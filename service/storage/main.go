package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"

	"github.com/gin-gonic/gin"
)

var (
	configFile = flag.String("config", "storage.json", "config file")
	port       = flag.String("port", ":5001", "port number")
)

func main() {
	flag.Parse()

	config, err := ioutil.ReadFile(*configFile)
	if err != nil {
		panic(err)
	}
	var accounts map[string]string
	err = json.Unmarshal(config, &accounts)
	if err != nil {
		panic(err)
	}

	r := gin.Default()

	authorized := r.Group("/", gin.BasicAuth(accounts))
	authorized.GET("/ping", ping)
	authorized.POST("/upload", upload)
	authorized.GET("/download", download)

	r.Run(*port)
}
