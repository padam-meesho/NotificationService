package utils

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type contextKey string

const traceIDKey contextKey = "trace_id"

// GetTraceID extracts trace ID from context
func GetTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value(traceIDKey).(string); ok {
		return traceID
	}
	return ""
}

// LogWithContext returns a logger with trace ID from context
func LogWithContext(ctx context.Context) zerolog.Logger {
	return zerolog.Ctx(ctx).With().Str("trace_id", GetTraceID(ctx)).Logger()
}

// ComponentLogger returns a logger for a specific component
func ComponentLogger(component string) zerolog.Logger {
	return log.With().Str("component", component).Logger()
}

// OperationLogger returns a logger for a specific operation within a component
func OperationLogger(component, operation string) zerolog.Logger {
	return log.With().
		Str("component", component).
		Str("operation", operation).
		Logger()
}

// RequestLogger returns a logger with request context
func RequestLogger(ctx context.Context, component, operation string) zerolog.Logger {
	logger := log.With().
		Str("component", component).
		Str("operation", operation)

	if traceID := GetTraceID(ctx); traceID != "" {
		logger = logger.Str("trace_id", traceID)
	}

	return logger.Logger()
}

// DatabaseLogger returns a logger for database operations
func DatabaseLogger(ctx context.Context, operation, table, requestID string) zerolog.Logger {
	logger := log.With().
		Str("component", "database").
		Str("operation", operation).
		Str("table", table)

	if requestID != "" {
		logger = logger.Str("request_id", requestID)
	}

	if traceID := GetTraceID(ctx); traceID != "" {
		logger = logger.Str("trace_id", traceID)
	}

	return logger.Logger()
}

// KafkaLogger returns a logger for Kafka operations
func KafkaLogger(operation, topic string) zerolog.Logger {
	return log.With().
		Str("component", "kafka").
		Str("operation", operation).
		Str("topic", topic).
		Logger()
}
