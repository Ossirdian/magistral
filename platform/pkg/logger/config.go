package logger

import "go.uber.org/zap"

type Config struct {
	Level        string
	AsJSON       bool
	StaticFields []Field
}

func String(key, val string) Field    { return zap.String(key, val) }
func Bool(key string, val bool) Field { return zap.Bool(key, val) }
func Int(key string, val int) Field   { return zap.Int(key, val) }
func Int64(key string, val int64) Field {
	return zap.Int64(key, val)
}
func Any(key string, val any) Field { return zap.Any(key, val) }
func Error(err error) Field         { return zap.Error(err) }
