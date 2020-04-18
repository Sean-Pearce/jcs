package main

import (
	"context"
	"errors"

	pb "github.com/Sean-Pearce/jcs/service/scheduler/proto"
)

type scheduler struct {
	name string
}

func newScheduler(name string) *scheduler {
	return &scheduler{name: name}
}

func (s *scheduler) Schedule(ctx context.Context, req *pb.ScheduleRequest) (*pb.ScheduleResponse, error) {
	res := &pb.ScheduleResponse{}

	if len(req.Sites) == 0 {
		return nil, errors.New("no site info provided")
	}

	res.Sites = req.Sites
	return res, nil
}
