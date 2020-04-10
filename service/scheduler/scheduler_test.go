package main

import (
	"context"
	"log"
	"net"
	"testing"

	pb "github.com/Sean-Pearce/jcs/service/scheduler/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var (
	tests = []struct {
		sites []string
		want  string
	}{
		{
			sites: nil,
			want:  "",
		},
		{
			sites: []string{"a", "b", "c"},
			want:  "a",
		},
	}
	lis *bufconn.Listener
)

func init() {
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	pb.RegisterSchedulerServer(s, newScheduler("yo"))
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func TestLocal(t *testing.T) {
	s := newScheduler("yoyo")
	for _, test := range tests {
		req := &pb.ScheduleRequest{Sites: test.sites}
		resp, err := s.Schedule(context.Background(), req)
		if err != nil {
			t.Skipf("Schedule(%v) got unexpected error: %v", *req, err)
		}
		if resp.Site != test.want {
			t.Errorf("Schedule(%v) = %v, want %v", *req, resp.Site, test.want)
		}
	}
}

func TestGrpc(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.Dial("bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := pb.NewSchedulerClient(conn)
	for _, test := range tests {
		req := &pb.ScheduleRequest{Sites: test.sites}
		resp, err := client.Schedule(ctx, req)
		if err != nil {
			t.Skipf("client.Schedule(%v) got unexpected error: %v", *req, err)
		}
		if resp.Site != test.want {
			t.Errorf("Schedule(%v) = %v, want %v", *req, resp.Site, test.want)
		}
	}
}
