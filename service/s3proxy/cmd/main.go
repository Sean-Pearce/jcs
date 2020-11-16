package main

import (
	"flag"
	"fmt"
	"net/http"
	"path"
	"runtime"
	"time"

	"github.com/Sean-Pearce/jcs/service/s3proxy"
	log "github.com/sirupsen/logrus"
)

var (
	flagMongoURL = flag.String("mongoURL", "mongodb://localhost:27017", "mongodb server address")
	flagAk       = flag.String("ak", "minioadmin", "minio access key")
	flagSk       = flag.String("sk", "minioadmin", "minio secret key")
	flagEndpoint = flag.String("endpoint", "http://127.0.0.1:9000", "minio endpoint")
	flagPort     = flag.String("port", ":5002", "server port number")
	flagTmpPath  = flag.String("tmp", "/tmp/jcs", "folder for tmp files")
)

func main() {
	flag.Parse()

	log.SetReportCaller(true)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			filename := path.Base(f.File)
			return fmt.Sprintf("%s()", f.Function), fmt.Sprintf("%s:%d", filename, f.Line)
		},
	})
	log.Infoln("s3proxy is running ...")

	p, err := s3proxy.NewProxy(*flagEndpoint, *flagAk, *flagSk, *flagMongoURL, *flagTmpPath)
	if err != nil {
		panic(err)
	}

	panic(http.ListenAndServe(*flagPort, p))
}
