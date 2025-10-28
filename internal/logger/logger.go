package logger

import (
	"fmt"
	"os"
	"strings"
	"sync/atomic"
)

type Level int32

const (
	PanicLevel Level = iota
	FatalLevel
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
	TraceLevel
)

var currentLevel int32 = int32(InfoLevel)

func SetLevel(level Level) {
	atomic.StoreInt32(&currentLevel, int32(level))
}

func GetLevel() Level {
	return Level(atomic.LoadInt32(&currentLevel))
}

func ParseLevel(level string) (Level, error) {
	switch strings.ToLower(level) {
	case "panic":
		return PanicLevel, nil
	case "fatal":
		return FatalLevel, nil
	case "error":
		return ErrorLevel, nil
	case "warn", "warning":
		return WarnLevel, nil
	case "info":
		return InfoLevel, nil
	case "debug":
		return DebugLevel, nil
	case "trace":
		return TraceLevel, nil
	case "":
		return InfoLevel, nil
	default:
		return InfoLevel, fmt.Errorf("unknown log level: %s", level)
	}
}

func levelEnabled(level Level) bool {
	return level <= GetLevel()
}

func logf(level Level, prefix string, format string, args ...interface{}) {
	if !levelEnabled(level) {
		return
	}
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}
	fmt.Fprintf(os.Stderr, "%s %s", prefix, fmt.Sprintf(format, args...))
}

func logln(level Level, prefix string, args ...interface{}) {
	if !levelEnabled(level) {
		return
	}
	fmt.Fprintf(os.Stderr, "%s %s\n", prefix, fmt.Sprint(args...))
}

func Tracef(format string, args ...interface{}) {
	logf(TraceLevel, "[TRACE]", format, args...)
}

func Debugf(format string, args ...interface{}) {
	logf(DebugLevel, "[DEBUG]", format, args...)
}

func Debug(args ...interface{}) {
	logln(DebugLevel, "[DEBUG]", args...)
}

func Infof(format string, args ...interface{}) {
	logf(InfoLevel, "[INFO]", format, args...)
}

func Info(args ...interface{}) {
	logln(InfoLevel, "[INFO]", args...)
}

func Warnf(format string, args ...interface{}) {
	logf(WarnLevel, "[WARN]", format, args...)
}

func Warn(args ...interface{}) {
	logln(WarnLevel, "[WARN]", args...)
}

func Errorf(format string, args ...interface{}) {
	logf(ErrorLevel, "[ERROR]", format, args...)
}

func Error(args ...interface{}) {
	logln(ErrorLevel, "[ERROR]", args...)
}

func Fatalf(format string, args ...interface{}) {
	logf(FatalLevel, "[FATAL]", format, args...)
	os.Exit(1)
}

func Fatal(args ...interface{}) {
	logln(FatalLevel, "[FATAL]", args...)
	os.Exit(1)
}

func Println(args ...interface{}) {
	logln(InfoLevel, "", args...)
}

func (l Level) String() string {
	switch l {
	case PanicLevel:
		return "panic"
	case FatalLevel:
		return "fatal"
	case ErrorLevel:
		return "error"
	case WarnLevel:
		return "warn"
	case InfoLevel:
		return "info"
	case DebugLevel:
		return "debug"
	case TraceLevel:
		return "trace"
	default:
		return "unknown"
	}
}
