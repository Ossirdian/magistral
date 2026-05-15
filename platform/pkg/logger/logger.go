package logger

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Key string

const (
	traceIDKey Key = "trace_id"
	userIDKey  Key = "user_id"
)

type Field = zap.Field

type Logger interface {
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	With(args ...any) Logger
	WithContext(ctx context.Context) Logger
	Sync() error
}

type zapLogger struct {
	base *zap.Logger
}

var (
	global Logger = &NoopLogger{}
	mu     sync.RWMutex
)

func Init(level string, asJSON bool) error {
	l, err := New(level, asJSON)
	if err != nil {
		return err
	}

	SetGlobal(l)
	return nil
}

func New(level string, asJSON bool) (Logger, error) {
	return NewWithConfig(Config{
		Level:  level,
		AsJSON: asJSON,
	})
}

func NewWithConfig(cfg Config) (Logger, error) {
	parsedLevel := parseLevel(cfg.Level)

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "ts"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	var encoder zapcore.Encoder
	if cfg.AsJSON {
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	} else {
		consoleCfg := encoderCfg
		consoleCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(consoleCfg)
	}

	core := zapcore.NewCore(
		encoder,
		zapcore.Lock(os.Stdout),
		parsedLevel,
	)

	return &zapLogger{base: zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1)).With(cfg.StaticFields...)}, nil
}

func SetGlobal(l Logger) {
	if l == nil {
		l = &NoopLogger{}
	}

	mu.Lock()
	defer mu.Unlock()

	global = l
}

func Global() Logger {
	mu.RLock()
	defer mu.RUnlock()

	return global
}

func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func (l *zapLogger) Info(msg string, args ...any) {
	l.base.Info(msg, toFields(args)...)
}

func (l *zapLogger) Warn(msg string, args ...any) {
	l.base.Warn(msg, toFields(args)...)
}

func (l *zapLogger) Error(msg string, args ...any) {
	l.base.Error(msg, toFields(args)...)
}

func (l *zapLogger) With(args ...any) Logger {
	return &zapLogger{base: l.base.With(toFields(args)...)}
}

func (l *zapLogger) WithContext(ctx context.Context) Logger {
	fields := make([]Field, 0, 2)

	if traceID, ok := ctx.Value(traceIDKey).(string); ok && traceID != "" {
		fields = append(fields, String(string(traceIDKey), traceID))
	}
	if userID, ok := ctx.Value(userIDKey).(string); ok && userID != "" {
		fields = append(fields, String(string(userIDKey), userID))
	}
	if len(fields) == 0 {
		return l
	}

	return &zapLogger{base: l.base.With(fields...)}
}

func (l *zapLogger) Sync() error {
	return l.base.Sync()
}

func parseLevel(level string) zapcore.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return zapcore.DebugLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

func toFields(args []any) []Field {
	if len(args) == 0 {
		return nil
	}

	fields := make([]Field, 0, len(args))
	allZapFields := true

	for _, arg := range args {
		field, ok := arg.(zap.Field)
		if !ok {
			allZapFields = false
			break
		}
		fields = append(fields, field)
	}
	if allZapFields {
		return fields
	}

	fields = fields[:0]
	for i := 0; i < len(args); i += 2 {
		if i+1 >= len(args) {
			fields = append(fields, Any(fmt.Sprintf("arg_%d", i), args[i]))
			break
		}

		key, ok := args[i].(string)
		if !ok || key == "" {
			key = fmt.Sprintf("arg_%d", i)
		}
		fields = append(fields, Any(key, args[i+1]))
	}

	return fields
}
