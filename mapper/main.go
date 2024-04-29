package main

import (
	"bytes"
	"encoding/gob"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/healthcheck", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "OK"})
	})

	r.GET("/map", func(c *gin.Context) {
		str := c.Query("str")

		words := strings.Split(str, " ")

		mapping := map[string]int{}

		for _, word := range words {
			if _, prs := mapping[word]; prs {
				mapping[word] += 1
			} else {
				mapping[word] = 1
			}
		}

		buf := new(bytes.Buffer)
		encoder := gob.NewEncoder(buf)
		_ = encoder.Encode(mapping)

		c.Data(http.StatusOK, "application/octet-stream", buf.Bytes())
	})

	r.Run(":8081")
}
