package closer

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/ossirdian/magistral/platform/pkg/logger"
)

const shutdownTimeout = 5 * time.Second

type Logger interface {
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

type Closer struct {
	mu     sync.Mutex
	once   sync.Once
	done   chan struct{}
	funcs  []func(context.Context) error
	logger Logger
}

var globalCloser = NewWithLogger(&logger.NoopLogger{})

func AddNamed(name string, f func(context.Context) error) {
	globalCloser.AddNamed(name, f)
}

func Add(f ...func(context.Context) error) {
	globalCloser.Add(f...)
}

func CloseAll(ctx context.Context) error {
	return globalCloser.CloseAll(ctx)
}

func SetLogger(l Logger) {
	globalCloser.SetLogger(l)
}

func Configure(signals ...os.Signal) {
	go globalCloser.handleSignals(signals...)
}

func New(signals ...os.Signal) *Closer {
	return NewWithLogger(logger.Global(), signals...)
}

func NewWithLogger(l Logger, signals ...os.Signal) *Closer {
	if l == nil {
		l = &logger.NoopLogger{}
	}

	c := &Closer{
		done:   make(chan struct{}),
		logger: l,
	}

	if len(signals) > 0 {
		go c.handleSignals(signals...)
	}

	return c
}

func (c *Closer) SetLogger(l Logger) {
	if l == nil {
		l = &logger.NoopLogger{}
	}

	c.logger = l
}

func (c *Closer) handleSignals(signals ...os.Signal) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, signals...)
	defer signal.Stop(ch)

	select {
	case <-ch:
		c.logger.Info("received shutdown signal, initiating graceful shutdown")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if err := c.CloseAll(shutdownCtx); err != nil {
			c.logger.Error("shutdown failed", "error", err)
		}
	case <-c.done:
	}
}

func (c *Closer) AddNamed(name string, f func(context.Context) error) {
	c.Add(func(ctx context.Context) error {
		start := time.Now()
		c.logger.Info("starting shutdown task", "name", name)

		err := f(ctx)
		duration := time.Since(start)
		if err != nil {
			c.logger.Error("shutdown task failed", "name", name, "duration", duration.String(), "error", err)
		} else {
			c.logger.Info("shutdown task completed", "name", name, "duration", duration.String())
		}

		return err
	})
}

func (c *Closer) Add(f ...func(context.Context) error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.funcs = append(c.funcs, f...)
}

func (c *Closer) CloseAll(ctx context.Context) error {
	var result error

	c.once.Do(func() {
		defer close(c.done)

		c.mu.Lock()
		funcs := c.funcs
		c.funcs = nil
		c.mu.Unlock()

		if len(funcs) == 0 {
			c.logger.Info("no shutdown tasks to execute")
			return
		}

		c.logger.Info("starting graceful shutdown", "tasks", len(funcs))

		errCh := make(chan error, len(funcs))
		var wg sync.WaitGroup

		for i := len(funcs) - 1; i >= 0; i-- {
			f := funcs[i]
			wg.Add(1)
			go func(f func(context.Context) error) {
				defer wg.Done()

				defer func() {
					if r := recover(); r != nil {
						errCh <- fmt.Errorf("panic recovered in closer: %v", r)
					}
				}()

				if err := f(ctx); err != nil {
					errCh <- err
				}
			}(f)
		}

		go func() {
			wg.Wait()
			close(errCh)
		}()

		for {
			select {
			case <-ctx.Done():
				c.logger.Error("shutdown context cancelled", "error", ctx.Err())
				if result == nil {
					result = ctx.Err()
				}
				return
			case err, ok := <-errCh:
				if !ok {
					c.logger.Info("all shutdown tasks completed")
					return
				}

				c.logger.Error("shutdown task returned error", "error", err)
				if result == nil {
					result = err
				} else {
					result = errors.Join(result, err)
				}
			}
		}
	})

	return result
}
