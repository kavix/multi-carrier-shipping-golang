package logger

import "go.uber.org/zap"

var log *zap.Logger

func Init() {
	var err error
	log, err = zap.NewProduction()
	if err != nil {
		panic(err)
	}
}

func Get() *zap.Logger {
	if log == nil {
		Init()
	}
	return log
}

func Info(msg string, fields ...zap.Field)  { Get().Info(msg, fields...) }
func Error(msg string, fields ...zap.Field) { Get().Error(msg, fields...) }
func Fatal(msg string, fields ...zap.Field) { Get().Fatal(msg, fields...) }
func String(key, val string) zap.Field     { return zap.String(key, val) }
func Int(key string, val int) zap.Field     { return zap.Int(key, val) }
func Float64(key string, val float64) zap.Field { return zap.Float64(key, val) }
