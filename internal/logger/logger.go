package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gitlab-hiring.cabify.tech/cabify/interviewing/car-pooling-challenge-go/utils"
)

type LogLevel string

const (
	Debug LogLevel = "DEBUG"
	Info  LogLevel = "INFO"
	Warn  LogLevel = "WARN"
	Error LogLevel = "ERROR"
)

type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     LogLevel               `json:"level"`
	Message   string                 `json:"message"`
	RequestID string                 `json:"request_id,omitempty"`
	Service   string                 `json:"service"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

type Logger struct {
	service  string
	minLevel LogLevel
}

func New(service string) *Logger {
	level := ParseLevel(utils.GetEnv("LOG_LEVEL", "ERROR"))
	return &Logger{
		service:  service,
		minLevel: level,
	}
}

// map priority
func levelPriority(level LogLevel) int {
	switch level {
	case Debug:
		return 1
	case Info:
		return 2
	case Warn:
		return 3
	case Error:
		return 4
	default:
		return 99
	}
}

// ParseLevel parses a string to LogLevel, defaults to INFO
func ParseLevel(s string) LogLevel {
	switch s {
	case "DEBUG":
		return Debug
	case "INFO":
		return Info
	case "WARN":
		return Warn
	case "ERROR":
		return Error
	default:
		return Info
	}
}

func (l *Logger) log(level LogLevel, message string, fields map[string]interface{}) {
	if levelPriority(level) < levelPriority(l.minLevel) {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Service:   l.service,
		Fields:    fields,
	}

	jsonData, _ := json.Marshal(entry)
	fmt.Fprintln(os.Stdout, string(jsonData))
}

func (l *Logger) Debug(message string, fields ...map[string]interface{}) {
	l.log(Debug, message, mergeFields(fields...))
}

func (l *Logger) Info(message string, fields ...map[string]interface{}) {
	l.log(Info, message, mergeFields(fields...))
}

func (l *Logger) Warn(message string, fields ...map[string]interface{}) {
	l.log(Warn, message, mergeFields(fields...))
}

func (l *Logger) Error(message string, fields ...map[string]interface{}) {
	l.log(Error, message, mergeFields(fields...))
}

func mergeFields(fields ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, field := range fields {
		for k, v := range field {
			result[k] = v
		}
	}
	return result
}

// RequestID utilities

const RequestIDKey = "request_id"

func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}

func SetRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, RequestIDKey, id)
}

func GenerateRequestID() string {
	return uuid.New().String()
}

func GinMiddleware(logger *Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		requestID := GenerateRequestID()

		// Add request ID to context
		c.Set("request_id", requestID)
		ctx := SetRequestID(c.Request.Context(), requestID)
		c.Request = c.Request.WithContext(ctx)

		// Log request start
		logger.Info("Request started", map[string]interface{}{
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"user_agent": c.Request.UserAgent(),
			"remote_ip":  c.ClientIP(),
		})

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()

		logger.Info("Request completed", map[string]interface{}{
			"status_code":   status,
			"duration_ms":   duration.Milliseconds(),
			"response_size": c.Writer.Size(),
		})

		// Log errors
		if status >= 400 {
			logger.Error("Request failed", map[string]interface{}{
				"status_code": status,
				"duration_ms": duration.Milliseconds(),
			})
		}
	}
}
