package main

import (
	"context"
	"fmt"
	homework "net-c12/homework/v1"
	"sync"
)

type Rosie struct {
	mu     sync.Mutex
	chores []*homework.Chore
	homework.UnimplementedRobotMaidServer
}

func (r *Rosie) Add(_ context.Context, chores *homework.Chores) (*homework.Response, error) {
	r.mu.Lock()

	r.chores = append(r.chores, chores.Chores...)
	r.mu.Unlock()
	return &homework.Response{Success: true}, nil
}

func (r *Rosie) Complete(_ context.Context,
	req *homework.CompleteRequest) (*homework.Response, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.chores == nil || req.ChoreNumber < 1 ||
		int(req.ChoreNumber) > len(r.chores) {
		return nil, fmt.Errorf("chore %d not found", req.ChoreNumber)
	}
	r.chores[req.ChoreNumber-1].Complete = true
	return &homework.Response{Success: true}, nil
}

func (r *Rosie) List(_ context.Context, _ *homework.Empty) (
	*homework.Chores, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.chores == nil {
		r.chores = make([]*homework.Chore, 0)
	}
	return &homework.Chores{Chores: r.chores}, nil
}

func (r *Rosie) Service() homework.RobotMaidServer {
	return r
}
