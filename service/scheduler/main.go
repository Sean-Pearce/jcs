package main

import (
	"flag"
	"net"

	pb "github.com/Sean-Pearce/jcs/service/scheduler/proto"
	"google.golang.org/grpc"
)

var (
	port = flag.String("port", ":5000", "grpc service port number")
)

func main() {
	flag.Parse()

	s := newScheduler("")
	gs := grpc.NewServer()
	pb.RegisterSchedulerServer(gs, s)

	lis, err := net.Listen("tcp", *port)
	if err != nil {
		panic(err)
	}

	panic(gs.Serve(lis))
}
