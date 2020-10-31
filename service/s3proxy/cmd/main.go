package main

import (
	"flag"
	"net/http"

	"github.com/Sean-Pearce/jcs/service/s3proxy"
	log "github.com/sirupsen/logrus"
)

var (
	flagMongoURL = flag.String("mongoURL", "mongodb://localhost:27017", "mongodb server address")
	flagAk       = flag.String("ak", "minioadmin", "minio access key")
	flagSk       = flag.String("sk", "minioadmin", "minio secret key")
	flagEndpoint = flag.String("endpoint", "http://127.0.0.1:9000", "minio endpoint")
	flagPort     = flag.String("port", ":5002", "server port number")
)

func main() {
	flag.Parse()
	log.Infoln("s3proxy is running ...")

	p, err := s3proxy.NewProxy(*flagEndpoint, *flagAk, *flagSk, *flagMongoURL)
	if err != nil {
		panic(err)
	}

	panic(http.ListenAndServe(*flagPort, p))
}
