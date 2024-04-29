package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	var client = &http.Client{}

	r.GET("/healthcheck", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "OK"})
	})

	r.GET("/compute", func(c *gin.Context) {
		text := c.Query("text")

		words := strings.Split(text, " ")

		// MAPPING

		mapperHost := os.Getenv("MAPPER_HOST")

		var mapperIps []string
		ips, _ := net.LookupIP(mapperHost)
		for _, ip := range ips {
			mapperIps = append(mapperIps, ip.String())
		}

		mapSplitCount := int(math.Ceil(float64(len(words)) / float64(len(mapperIps))))

		var mapSplits = map[string][]string{}

		for idx, mapperIp := range mapperIps {
			if idx*mapSplitCount >= len(words) {
				break
			}
			mapSplits[mapperIp] = words[idx*mapSplitCount : min(idx*mapSplitCount+mapSplitCount, len(words))]
		}

		var mapping = map[string]map[string]int{}
		var mappingMutex sync.Mutex

		var wgm sync.WaitGroup
		wgm.Add(len(mapSplits))

		for host, split := range mapSplits {
			go func(host string, split []string) {
				defer wgm.Done()

				for {
					req, err := http.NewRequest("GET", fmt.Sprintf("http://%s:%s/map", host, os.Getenv("MAPPER_PORT")), nil)
					if err != nil {
						fmt.Println("Error creating request:", err)
						continue
					}

					q := req.URL.Query()
					q.Add("str", strings.Join(split, " "))
					req.URL.RawQuery = q.Encode()

					res, err := client.Do(req)
					if err != nil {
						fmt.Println("Error sending request:", err)
						time.Sleep(time.Second)
						continue
					}
					defer res.Body.Close()

					body, err := io.ReadAll(res.Body)
					if err != nil {
						fmt.Println("Error reading response body:", err)
						continue
					}

					buf := bytes.NewBuffer(body)

					var decodedMap map[string]int
					decoder := gob.NewDecoder(buf)
					if err := decoder.Decode(&decodedMap); err != nil {
						fmt.Println("Error decoding response body:", err)
						continue
					}

					mappingMutex.Lock()
					mapping[host] = decodedMap
					mappingMutex.Unlock()

					break
				}
			}(host, split)
		}

		wgm.Wait()

		// SHUFFLING

		var shuffling = map[string][]int{}

		for _, host := range mapping {
			for word, count := range host {
				shuffling[word] = append(shuffling[word], count)
			}
		}

		// REDUCING

		reducerHost := os.Getenv("REDUCER_HOST")

		var reducerIps []string
		ips, _ = net.LookupIP(reducerHost)
		for _, ip := range ips {
			reducerIps = append(reducerIps, ip.String())
		}

		var shuffleWords []string
		for word := range shuffling {
			shuffleWords = append(shuffleWords, word)
		}

		reduceSplitCount := int(math.Ceil(float64(len(shuffleWords)) / float64(len(reducerIps))))

		var reduceSplits = map[string]map[string][]int{}

		for idx, reducerIp := range reducerIps {
			if idx*reduceSplitCount >= len(shuffleWords) {
				break
			}
			reduceWords := shuffleWords[idx*reduceSplitCount : min(idx*reduceSplitCount+reduceSplitCount, len(shuffleWords))]

			reduceSplits[reducerIp] = map[string][]int{}
			for _, reduceKey := range reduceWords {
				reduceSplits[reducerIp][reduceKey] = shuffling[reduceKey]
			}
		}

		var wgr sync.WaitGroup
		wgr.Add(len(reduceSplits))

		var reducing = map[string]map[string]int{}

		for host, split := range reduceSplits {
			go func(host string, split map[string][]int) {
				defer wgr.Done()

				for {
					req, err := http.NewRequest("GET", fmt.Sprintf("http://%s:%s/reduce", host, os.Getenv("REDUCER_PORT")), nil)
					if err != nil {
						fmt.Println("Error creating request:", err)
						continue
					}

					buf := new(bytes.Buffer)
					encoder := gob.NewEncoder(buf)
					if err := encoder.Encode(split); err != nil {
						fmt.Println("Error encoding request body:", err)
						continue
					}

					q := req.URL.Query()
					q.Add("body", buf.String())
					req.URL.RawQuery = q.Encode()

					res, err := client.Do(req)
					if err != nil {
						fmt.Println("Error sending request:", err)
						time.Sleep(time.Second)
						continue
					}
					defer res.Body.Close()

					body, err := io.ReadAll(res.Body)
					if err != nil {
						fmt.Println("Error reading response body:", err)
						continue
					}

					buf = bytes.NewBuffer(body)

					var decodedReduce = map[string]int{}
					decoder := gob.NewDecoder(buf)
					if err := decoder.Decode(&decodedReduce); err != nil {
						fmt.Println("Error decoding response body:", err)
						continue
					}

					mappingMutex.Lock()
					reducing[host] = decodedReduce
					mappingMutex.Unlock()

					break
				}
			}(host, split)
		}

		wgr.Wait()

		c.JSON(http.StatusOK, reducing)
	})

	r.Run(":8080")
}
