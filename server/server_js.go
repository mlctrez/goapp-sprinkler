//go:build wasm

package server

import (
	"context"
	"github.com/mlctrez/goapp-sprinkler/schedule"
)

func Run(s *schedule.Schedule) (shutdownFunc func(ctx context.Context) error, err error) { return }
