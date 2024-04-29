package main

import (
	"bytes"
	"encoding/gob"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/healthcheck", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "OK"})
	})

	r.GET("/reduce", func(c *gin.Context) {
		body := c.Query("body")

		buf := bytes.NewBuffer([]byte(body))

		var reduceData = map[string][]int{}

		decoder := gob.NewDecoder(buf)
		_ = decoder.Decode(&reduceData)

		var reducing = map[string]int{}

		for key, value := range reduceData {
			reducing[key] = 0
			for _, count := range value {
				reducing[key] += count
			}
		}

		buf = new(bytes.Buffer)
		encoder := gob.NewEncoder(buf)
		_ = encoder.Encode(reducing)

		c.Data(http.StatusOK, "application/octet-stream", buf.Bytes())
	})

	r.Run(":8082")
}
