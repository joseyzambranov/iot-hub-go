package logging

import (
	"log"
	"os"
)

type Logger struct {
	*log.Logger
}

func NewLogger() *Logger {
	return &Logger{
		Logger: log.New(os.Stdout, "", log.LstdFlags),
	}
}

func (l *Logger) Info(msg string) {
	l.Printf("ℹ️ INFO: %s", msg)
}

func (l *Logger) Warning(msg string) {
	l.Printf("⚠️ WARNING: %s", msg)
}

func (l *Logger) Error(msg string) {
	l.Printf("❌ ERROR: %s", msg)
}

func (l *Logger) Security(msg string) {
	l.Printf("🔒 SECURITY: %s", msg)
}

func (l *Logger) Anomaly(msg string) {
	l.Printf("🚨 ANOMALY: %s", msg)
}

func (l *Logger) Success(msg string) {
	l.Printf("✅ SUCCESS: %s", msg)
}