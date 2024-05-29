package gocql

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"strings"
)

type StdLogger interface {
	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}

type nopLogger struct{}

func (n nopLogger) Print(_ ...interface{}) {}

func (n nopLogger) Printf(_ string, _ ...interface{}) {}

func (n nopLogger) Println(_ ...interface{}) {}

func (n nopLogger) Error(_ string, _ ...LogField) {}

func (n nopLogger) Warning(_ string, _ ...LogField) {}

func (n nopLogger) Info(_ string, _ ...LogField) {}

func (n nopLogger) Debug(_ string, _ ...LogField) {}

type testLogger struct {
	capture bytes.Buffer
}

func (l *testLogger) Print(v ...interface{})                 { fmt.Fprint(&l.capture, v...) }
func (l *testLogger) Printf(format string, v ...interface{}) { fmt.Fprintf(&l.capture, format, v...) }
func (l *testLogger) Println(v ...interface{})               { fmt.Fprintln(&l.capture, v...) }
func (l *testLogger) String() string                         { return l.capture.String() }

type defaultLogger struct{}

func (l *defaultLogger) Print(v ...interface{})                 { log.Print(v...) }
func (l *defaultLogger) Printf(format string, v ...interface{}) { log.Printf(format, v...) }
func (l *defaultLogger) Println(v ...interface{})               { log.Println(v...) }

// Logger for logging messages.
// Deprecated: Use ClusterConfig.Logger instead.
var Logger StdLogger = &defaultLogger{}

var nilInternalLogger internalLogger = loggerAdapter{
	minimumLogLevel: LogLevelNone,
	advLogger:       nopLogger{},
	legacyLogger:    nil,
}

type LogLevel int

const (
	LogLevelDebug = LogLevel(5)
	LogLevelInfo  = LogLevel(4)
	LogLevelWarn  = LogLevel(3)
	LogLevelError = LogLevel(2)
	LogLevelNone  = LogLevel(0)
)

func (recv LogLevel) String() string {
	switch recv {
	case LogLevelDebug:
		return "debug"
	case LogLevelInfo:
		return "info"
	case LogLevelWarn:
		return "warn"
	case LogLevelError:
		return "error"
	case LogLevelNone:
		return "none"
	default:
		// fmt.sprintf allocates so use strings.Join instead
		temp := [2]string{"invalid level ", strconv.Itoa(int(recv))}
		return strings.Join(temp[:], "")
	}
}

type LogField struct {
	Name  string
	Value interface{}
}

func NewLogField(name string, value interface{}) LogField {
	return LogField{
		Name:  name,
		Value: value,
	}
}

type AdvancedLogger interface {
	Error(msg string, fields ...LogField)
	Warning(msg string, fields ...LogField)
	Info(msg string, fields ...LogField)
	Debug(msg string, fields ...LogField)
}

type internalLogger interface {
	AdvancedLogger
	MinimumLogLevel() LogLevel
}

type loggerAdapter struct {
	minimumLogLevel LogLevel
	advLogger       AdvancedLogger
	legacyLogger    StdLogger
}

func (recv loggerAdapter) logLegacy(msg string, fields ...LogField) {
	var values []interface{}
	var small [5]interface{}
	l := len(fields)
	if l <= 5 { // small stack array optimization
		values = small[:l]
	} else {
		values = make([]interface{}, l)
	}
	var i int
	for _, v := range fields {
		values[i] = v.Value
		i++
	}
	recv.legacyLogger.Printf(msg, values...)
}

func (recv loggerAdapter) Error(msg string, fields ...LogField) {
	if LogLevelError <= recv.minimumLogLevel {
		if recv.advLogger != nil {
			recv.advLogger.Error(msg, fields...)
		} else {
			recv.logLegacy(msg, fields...)
		}
	}
}

func (recv loggerAdapter) Warning(msg string, fields ...LogField) {
	if LogLevelWarn <= recv.minimumLogLevel {
		if recv.advLogger != nil {
			recv.advLogger.Warning(msg, fields...)
		} else {
			recv.logLegacy(msg, fields...)
		}
	}
}

func (recv loggerAdapter) Info(msg string, fields ...LogField) {
	if LogLevelInfo <= recv.minimumLogLevel {
		if recv.advLogger != nil {
			recv.advLogger.Info(msg, fields...)
		} else {
			recv.logLegacy(msg, fields...)
		}
	}
}

func (recv loggerAdapter) Debug(msg string, fields ...LogField) {
	if LogLevelDebug <= recv.minimumLogLevel {
		if recv.advLogger != nil {
			recv.advLogger.Debug(msg, fields...)
		} else {
			recv.logLegacy(msg, fields...)
		}
	}
}

func (recv loggerAdapter) MinimumLogLevel() LogLevel {
	return recv.minimumLogLevel
}

func newInternalLoggerFromAdvancedLogger(logger AdvancedLogger, level LogLevel) loggerAdapter {
	return loggerAdapter{
		minimumLogLevel: level,
		advLogger:       logger,
		legacyLogger:    nil,
	}
}

func newInternalLoggerFromStdLogger(logger StdLogger, level LogLevel) loggerAdapter {
	return loggerAdapter{
		minimumLogLevel: level,
		advLogger:       nil,
		legacyLogger:    logger,
	}
}
