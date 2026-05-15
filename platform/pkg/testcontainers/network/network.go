package network

import (
	"context"
	"fmt"
	"time"
)

type Network struct {
	name string
}

func NewNetwork(ctx context.Context, projectName string) (*Network, error) {
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("context canceled: %w", ctx.Err())
	default:
	}

	return &Network{
		name: fmt.Sprintf("%s-%d", projectName, time.Now().UnixNano()),
	}, nil
}

func (n *Network) Name() string {
	return n.name
}
