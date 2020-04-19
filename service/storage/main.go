package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const (
	version = "v0.1"
)

var (
	configFile = flag.String("config", "storage.json", "config file")
	port       = flag.String("port", ":5001", "port number")
	endpoint   = flag.String("endpoint", "127.0.0.1:9000", "minio endpoint")
	accessKey  = flag.String("ak", "", "access key")
	secretKey  = flag.String("sk", "", "secret key")
	useSSL     = flag.Bool("ssl", false, "minio use ssl")
	debug      = flag.Bool("debug", false, "debug mode")
)

func main() {
	log.Infoln("Starting storage", version)

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
	authorized.DELETE("/delete", deleteFile)

	r.Run(*port)
}
