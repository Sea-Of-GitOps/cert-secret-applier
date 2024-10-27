package logger

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Init(logLevel string) error {
	zapConfig := config(logLevel)
	logger, _ := zapConfig.Build(zap.AddCallerSkip(1))
	zap.ReplaceGlobals(logger)
	zap.RedirectStdLog(logger)
	return nil
}

func encoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		TimeKey:        "ts",
		CallerKey:      "caller",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

func config(level string) *zap.Config {
	atomicLevel, _ := zap.ParseAtomicLevel(level)
	return &zap.Config{
		Level:             atomicLevel,
		Development:       true,
		DisableCaller:     false,
		DisableStacktrace: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         "json",
		EncoderConfig:    encoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
}

type RequestKey string

const RequestIdKey RequestKey = "request-id"

func Sync() error {
	return zap.L().Sync()
}

func appendRequestId(ctx context.Context, fields ...zap.Field) []zap.Field {
	requestId := ctx.Value(RequestIdKey)
	if requestId != nil {
		fields = append(fields, zap.String("requestId", fmt.Sprintf("%s", ctx.Value(RequestIdKey))))
	}
	return fields
}

func DebugCtx(ctx context.Context, message string, fields ...zap.Field) {
	zap.L().Debug(message, appendRequestId(ctx, fields...)...)
}

func Infof(message string, fields ...interface{}) {
	zap.S().WithOptions(zap.AddCallerSkip(0)).Infof(message, fields...)
}

func InfoCtx(ctx context.Context, message string, fields ...zap.Field) {
	zap.L().Info(message, appendRequestId(ctx, fields...)...)
}

func Errorf(message string, fields ...interface{}) {
	zap.S().WithOptions(zap.AddCallerSkip(0)).Errorf(message, fields...)
}

func ErrorCtx(ctx context.Context, message string, fields ...zap.Field) {
	zap.L().Error(message, appendRequestId(ctx, fields...)...)
}

func WarnCtx(ctx context.Context, message string, fields ...zap.Field) {
	zap.L().Warn(message, appendRequestId(ctx, fields...)...)
}
