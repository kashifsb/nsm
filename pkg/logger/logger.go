package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

var globalLogger zerolog.Logger

type Config struct {
	Level      string
	File       string
	Console    bool
	Pretty     bool
	MaxSize    int
	MaxAge     int
	MaxBackups int
}

func Init(debug bool) {
	cfg := Config{
		Level:      "info",
		Console:    true,
		Pretty:     true,
		MaxSize:    10, // megabytes
		MaxAge:     30, // days
		MaxBackups: 5,
	}

	if debug {
		cfg.Level = "debug"
	}

	InitWithConfig(cfg)
}

func InitWithConfig(cfg Config) {
	// Parse log level
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	var writers []io.Writer

	// Console writer
	if cfg.Console {
		console := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "15:04:05",
			NoColor:    !cfg.Pretty,
			FormatLevel: func(i interface{}) string {
				if !cfg.Pretty {
					return fmt.Sprintf("[%s]", i)
				}

				s, _ := i.(string)
				switch s {
				case "trace":
					return "\033[35m🔍\033[0m"
				case "debug":
					return "\033[36m🐛\033[0m"
				case "info":
					return "\033[32mℹ️\033[0m"
				case "warn":
					return "\033[33m⚠️\033[0m"
				case "error":
					return "\033[31m❌\033[0m"
				case "fatal":
					return "\033[91m💀\033[0m"
				case "panic":
					return "\033[91m🚨\033[0m"
				default:
					return "\033[37m📝\033[0m"
				}
			},
			FormatMessage: func(i interface{}) string {
				return fmt.Sprintf("%s", i)
			},
			FormatFieldName: func(i interface{}) string {
				return fmt.Sprintf("\033[36m%s\033[0m=", i)
			},
			FormatFieldValue: func(i interface{}) string {
				return fmt.Sprintf("\033[37m%s\033[0m", i)
			},
		}
		writers = append(writers, console)
	}

	// File writer
	if cfg.File != "" {
		// Ensure log directory exists
		if err := os.MkdirAll(filepath.Dir(cfg.File), 0o755); err != nil {
			fmt.Printf("Failed to create log directory: %v\n", err)
		} else {
			fileWriter := &lumberjack.Logger{
				Filename:   cfg.File,
				MaxSize:    cfg.MaxSize,
				MaxBackups: cfg.MaxBackups,
				MaxAge:     cfg.MaxAge,
				Compress:   true,
			}
			writers = append(writers, fileWriter)
		}
	}

	// Create multi-writer
	var writer io.Writer
	if len(writers) == 0 {
		// Fallback to stdout if no writers configured
		writer = os.Stdout
	} else if len(writers) == 1 {
		writer = writers[0]
	} else {
		writer = io.MultiWriter(writers...)
	}

	// Create logger
	globalLogger = zerolog.New(writer).With().
		Timestamp().
		Caller().
		Logger()

	// Set global logger
	log.Logger = globalLogger
}

// Convenience functions
func Debug(msg string, keysAndValues ...interface{}) {
	event := globalLogger.Debug()
	addFields(event, keysAndValues...)
	event.Msg(msg)
}

func Info(msg string, keysAndValues ...interface{}) {
	event := globalLogger.Info()
	addFields(event, keysAndValues...)
	event.Msg(msg)
}

func Warn(msg string, keysAndValues ...interface{}) {
	event := globalLogger.Warn()
	addFields(event, keysAndValues...)
	event.Msg(msg)
}

func Error(msg string, keysAndValues ...interface{}) {
	event := globalLogger.Error()
	addFields(event, keysAndValues...)
	event.Msg(msg)
}

func Fatal(msg string, keysAndValues ...interface{}) {
	event := globalLogger.Fatal()
	addFields(event, keysAndValues...)
	event.Msg(msg)
}

func Panic(msg string, keysAndValues ...interface{}) {
	event := globalLogger.Panic()
	addFields(event, keysAndValues...)
	event.Msg(msg)
}

func addFields(event *zerolog.Event, keysAndValues ...interface{}) {
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			key, ok := keysAndValues[i].(string)
			if ok {
				value := keysAndValues[i+1]
				event.Interface(key, value)
			}
		}
	}
}

// Context-aware logging
type ContextLogger struct {
	logger zerolog.Logger
	fields map[string]interface{}
}

func WithContext(fields map[string]interface{}) *ContextLogger {
	return &ContextLogger{
		logger: globalLogger,
		fields: fields,
	}
}

func (cl *ContextLogger) Debug(msg string, keysAndValues ...interface{}) {
	event := cl.logger.Debug()
	cl.addContextFields(event)
	addFields(event, keysAndValues...)
	event.Msg(msg)
}

func (cl *ContextLogger) Info(msg string, keysAndValues ...interface{}) {
	event := cl.logger.Info()
	cl.addContextFields(event)
	addFields(event, keysAndValues...)
	event.Msg(msg)
}

func (cl *ContextLogger) Warn(msg string, keysAndValues ...interface{}) {
	event := cl.logger.Warn()
	cl.addContextFields(event)
	addFields(event, keysAndValues...)
	event.Msg(msg)
}

func (cl *ContextLogger) Error(msg string, keysAndValues ...interface{}) {
	event := cl.logger.Error()
	cl.addContextFields(event)
	addFields(event, keysAndValues...)
	event.Msg(msg)
}

func (cl *ContextLogger) Fatal(msg string, keysAndValues ...interface{}) {
	event := cl.logger.Fatal()
	cl.addContextFields(event)
	addFields(event, keysAndValues...)
	event.Msg(msg)
}

func (cl *ContextLogger) Panic(msg string, keysAndValues ...interface{}) {
	event := cl.logger.Panic()
	cl.addContextFields(event)
	addFields(event, keysAndValues...)
	event.Msg(msg)
}

func (cl *ContextLogger) addContextFields(event *zerolog.Event) {
	for key, value := range cl.fields {
		event.Interface(key, value)
	}
}

// Logger instance methods for when you need the logger directly
func GetLogger() zerolog.Logger {
	return globalLogger
}

func WithFields(fields map[string]interface{}) *ContextLogger {
	return WithContext(fields)
}

// Structured logging helpers
func LogError(err error, msg string, keysAndValues ...interface{}) {
	event := globalLogger.Error().Err(err)
	addFields(event, keysAndValues...)
	event.Msg(msg)
}

func LogErrorWithStack(err error, msg string, keysAndValues ...interface{}) {
	event := globalLogger.Error().Err(err).Stack()
	addFields(event, keysAndValues...)
	event.Msg(msg)
}

// Performance logging
func LogDuration(name string, start time.Time, keysAndValues ...interface{}) {
	duration := time.Since(start)
	event := globalLogger.Info().
		Str("operation", name).
		Dur("duration", duration)
	addFields(event, keysAndValues...)
	event.Msg("Operation completed")
}

// HTTP request logging
func LogHTTPRequest(method, path string, statusCode int, duration time.Duration, keysAndValues ...interface{}) {
	var event *zerolog.Event

	switch {
	case statusCode >= 500:
		event = globalLogger.Error()
	case statusCode >= 400:
		event = globalLogger.Warn()
	default:
		event = globalLogger.Info()
	}

	event = event.
		Str("method", method).
		Str("path", path).
		Int("status", statusCode).
		Dur("duration", duration)

	addFields(event, keysAndValues...)
	event.Msg("HTTP request")
}

// Configuration logging
func LogConfig(cfg interface{}, keysAndValues ...interface{}) {
	event := globalLogger.Info().Interface("config", cfg)
	addFields(event, keysAndValues...)
	event.Msg("Configuration loaded")
}

// System event logging
func LogSystemEvent(event string, keysAndValues ...interface{}) {
	logEvent := globalLogger.Info().Str("event", event)
	addFields(logEvent, keysAndValues...)
	logEvent.Msg("System event")
}

// Security logging
func LogSecurityEvent(event string, severity string, keysAndValues ...interface{}) {
	var logEvent *zerolog.Event

	switch severity {
	case "critical":
		logEvent = globalLogger.Error()
	case "high":
		logEvent = globalLogger.Warn()
	default:
		logEvent = globalLogger.Info()
	}

	logEvent = logEvent.
		Str("security_event", event).
		Str("severity", severity)

	addFields(logEvent, keysAndValues...)
	logEvent.Msg("Security event")
}

// Development helpers
func DevDebug(msg string, keysAndValues ...interface{}) {
	if globalLogger.GetLevel() <= zerolog.DebugLevel {
		event := globalLogger.Debug().Str("dev", "true")
		addFields(event, keysAndValues...)
		event.Msg(msg)
	}
}

func DevInfo(msg string, keysAndValues ...interface{}) {
	event := globalLogger.Info().Str("dev", "true")
	addFields(event, keysAndValues...)
	event.Msg(msg)
}

// File rotation helper
func RotateLogFile() error {
	// The lumberjack library handles this automatically, but this provides manual control
	// In a real implementation, you'd need to keep a reference to the lumberjack.Logger
	return nil
}

// Cleanup function - simplified since zerolog.Logger doesn't need explicit closing
func Close() error {
	// zerolog doesn't require explicit cleanup, but we can flush any pending writes
	return nil
}

// Test helpers
func SetTestMode() {
	globalLogger = zerolog.New(io.Discard).Level(zerolog.Disabled)
	log.Logger = globalLogger
}

func ResetLogger() {
	globalLogger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	log.Logger = globalLogger
}

// Additional utility functions
func SetLevel(level string) error {
	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		return fmt.Errorf("invalid log level: %w", err)
	}
	zerolog.SetGlobalLevel(lvl)
	return nil
}

func GetLevel() string {
	return globalLogger.GetLevel().String()
}

// Component logger for better organization
func Component(name string) *ContextLogger {
	return WithContext(map[string]interface{}{
		"component": name,
	})
}

// Request ID logger for tracing
func WithRequestID(requestID string) *ContextLogger {
	return WithContext(map[string]interface{}{
		"request_id": requestID,
	})
}

// User context logger
func WithUser(userID string) *ContextLogger {
	return WithContext(map[string]interface{}{
		"user_id": userID,
	})
}

// Project context logger
func WithProject(projectName, projectType string) *ContextLogger {
	return WithContext(map[string]interface{}{
		"project_name": projectName,
		"project_type": projectType,
	})
}

// Network context logger
func WithNetwork(host string, port int) *ContextLogger {
	return WithContext(map[string]interface{}{
		"host": host,
		"port": port,
	})
}

// Error with context
func ErrorWithContext(err error, context map[string]interface{}, msg string) {
	event := globalLogger.Error().Err(err)
	for key, value := range context {
		event.Interface(key, value)
	}
	event.Msg(msg)
}

// Panic recovery logging
func LogPanicRecovery(recovered interface{}, keysAndValues ...interface{}) {
	event := globalLogger.Error().
		Interface("panic", recovered).
		Stack()
	addFields(event, keysAndValues...)
	event.Msg("Panic recovered")
}

// Startup/shutdown logging
func LogStartup(service string, version string, keysAndValues ...interface{}) {
	event := globalLogger.Info().
		Str("service", service).
		Str("version", version).
		Str("event", "startup")
	addFields(event, keysAndValues...)
	event.Msg("Service starting")
}

func LogShutdown(service string, keysAndValues ...interface{}) {
	event := globalLogger.Info().
		Str("service", service).
		Str("event", "shutdown")
	addFields(event, keysAndValues...)
	event.Msg("Service shutting down")
}
