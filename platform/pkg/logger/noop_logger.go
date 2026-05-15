package logger

import "context"

type NoopLogger struct{}

func (l *NoopLogger) Error(msg string, args ...any)          {}
func (l *NoopLogger) Info(msg string, args ...any)           {}
func (l *NoopLogger) Warn(msg string, args ...any)           {}
func (l *NoopLogger) With(args ...any) Logger                { return l }
func (l *NoopLogger) WithContext(ctx context.Context) Logger { return l }
func (l *NoopLogger) Sync() error                            { return nil }
