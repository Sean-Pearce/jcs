package main

import (
	"flag"
	"net"

	pb "github.com/Sean-Pearce/jcs/service/scheduler/proto"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	version = "v0.1"
)

var (
	port = flag.String("port", ":5001", "grpc service port number")
)

func main() {
	flag.Parse()

	log.Infoln("Starting scheduler", version)

	s := newScheduler("")
	gs := grpc.NewServer()
	pb.RegisterSchedulerServer(gs, s)

	lis, err := net.Listen("tcp", *port)
	if err != nil {
		panic(err)
	}

	panic(gs.Serve(lis))
}
