package logger

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// Logger is a wrapper around zap logger
type Logger interface {
	Debug(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
	Fatal(msg string, keysAndValues ...interface{})
	With(keysAndValues ...interface{}) Logger
	Sync() error
}

type logger struct {
	zap *zap.SugaredLogger
}

// NewLogger creates a new logger instance
func NewLogger() Logger {
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	zapLogger, err := config.Build()
	if err != nil {
		panic(err)
	}

	return &logger{
		zap: zapLogger.Sugar(),
	}
}

// NewDevelopmentLogger creates a new development logger
func NewDevelopmentLogger() Logger {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	zapLogger, err := config.Build()
	if err != nil {
		panic(err)
	}

	return &logger{
		zap: zapLogger.Sugar(),
	}
}

func (l *logger) Debug(msg string, keysAndValues ...interface{}) {
	l.zap.Debugw(msg, keysAndValues...)
}

func (l *logger) Info(msg string, keysAndValues ...interface{}) {
	l.zap.Infow(msg, keysAndValues...)
}

func (l *logger) Warn(msg string, keysAndValues ...interface{}) {
	l.zap.Warnw(msg, keysAndValues...)
}

func (l *logger) Error(msg string, keysAndValues ...interface{}) {
	l.zap.Errorw(msg, keysAndValues...)
}

func (l *logger) Fatal(msg string, keysAndValues ...interface{}) {
	l.zap.Fatalw(msg, keysAndValues...)
}

func (l *logger) With(keysAndValues ...interface{}) Logger {
	return &logger{
		zap: l.zap.With(keysAndValues...),
	}
}

func (l *logger) Sync() error {
	return l.zap.Sync()
}

// UnaryServerInterceptor returns a new unary server interceptor for logging
func UnaryServerInterceptor(logger Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Extract metadata from context
		md, _ := metadata.FromIncomingContext(ctx)
		requestID := ""
		if vals := md.Get("x-request-id"); len(vals) > 0 {
			requestID = vals[0]
		}

		// Create logger with context
		log := logger.With("method", info.FullMethod, "request_id", requestID)
		log.Debug("gRPC request started")

		// Call handler
		resp, err := handler(ctx, req)

		// Log result
		if err != nil {
			log.Error("gRPC request failed", "error", err)
		} else {
			log.Debug("gRPC request completed")
		}

		return resp, err
	}
}
