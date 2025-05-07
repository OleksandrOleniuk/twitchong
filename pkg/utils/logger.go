package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Logger is the global logger instance
	Logger *zap.Logger
)

// Custom color encoders
const (
	// ANSI color codes
	red     = "\x1b[31m"
	green   = "\x1b[32m"
	yellow  = "\x1b[33m"
	blue    = "\x1b[34m"
	magenta = "\x1b[35m"
	cyan    = "\x1b[36m"
	white   = "\x1b[37m"

	// Bright colors
	brightRed     = "\x1b[91m"
	brightGreen   = "\x1b[92m"
	brightYellow  = "\x1b[93m"
	brightBlue    = "\x1b[94m"
	brightMagenta = "\x1b[95m"
	brightCyan    = "\x1b[96m"

	// Muted colors
	mutedRed     = "\x1b[38;5;88m"
	mutedGreen   = "\x1b[38;5;28m"
	mutedYellow  = "\x1b[38;5;58m"
	mutedBlue    = "\x1b[38;5;24m"
	mutedMagenta = "\x1b[38;5;90m"
	mutedCyan    = "\x1b[38;5;30m"

	reset = "\x1b[0m"
)

func ZapString(key string, value string) zap.Field {
	return zap.String(key, value)
}

// Custom color encoders
func colorTimeEncoder(t zapcore.TimeEncoder) zapcore.TimeEncoder {
	return func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(white + t.Format("2006-01-02 15:04:05.000") + reset)
	}
}

func colorCallerEncoder(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(mutedMagenta + caller.TrimmedPath() + reset)
}

func init() {
	// Configure the encoder
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     colorTimeEncoder(zapcore.ISO8601TimeEncoder),
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   colorCallerEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}

	// Configure the core
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		zapcore.DebugLevel,
	)

	// Create the logger with additional options
	Logger = zap.New(core,
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
}

// Info logs an info message
func Info(msg string, fields ...zapcore.Field) {
	Logger.Info(msg, fields...)
}

// Error logs an error message
func Error(msg string, fields ...zapcore.Field) {
	Logger.Error(msg, fields...)
}

// Warn logs a warning message
func Warn(msg string, fields ...zapcore.Field) {
	Logger.Warn(msg, fields...)
}

// Debug logs a debug message
func Debug(msg string, fields ...zapcore.Field) {
	Logger.Debug(msg, fields...)
}

// Fatal logs a fatal message and exits
func Fatal(msg string, fields ...zapcore.Field) {
	Logger.Fatal(msg, fields...)
}

// Panic logs a panic message and then panics
func Panic(msg string, fields ...zapcore.Field) {
	Logger.Panic(msg, fields...)
}

// With creates a child logger and adds structured context to it
func With(fields ...zapcore.Field) *zap.Logger {
	return Logger.With(fields...)
}

// Object creates a field that formats an object as JSON
func Object(key string, obj interface{}) zapcore.Field {
	// Marshal the object to JSON with indentation
	jsonData, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return zap.String(key, fmt.Sprintf("error marshaling object: %v", err))
	}
	return zap.String(key, string(jsonData))
}

// PrettyObject logs an object in a nicely formatted way
func PrettyObject(msg string, key string, obj interface{}) {
	Logger.Info(msg, Object(key, obj))
}

