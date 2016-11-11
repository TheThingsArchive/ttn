package api

import "github.com/apex/log"

// Logger used in mqtt package
type Logger interface {
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
	Fatal(msg string)
	Debugf(msg string, v ...interface{})
	Infof(msg string, v ...interface{})
	Warnf(msg string, v ...interface{})
	Errorf(msg string, v ...interface{})
	Fatalf(msg string, v ...interface{})
	WithField(string, interface{}) Logger
	WithFields(log.Fielder) Logger
	WithError(error) Logger
}

// Apex wraps apex/log
func Apex(ctx log.Interface) Logger {
	return &apexInterfaceWrapper{ctx}
}

type apexInterfaceWrapper struct {
	log.Interface
}

func (w *apexInterfaceWrapper) WithField(k string, v interface{}) Logger {
	return &apexEntryWrapper{w.Interface.WithField(k, v)}
}

func (w *apexInterfaceWrapper) WithFields(fields log.Fielder) Logger {
	return &apexEntryWrapper{w.Interface.WithFields(fields)}
}

func (w *apexInterfaceWrapper) WithError(err error) Logger {
	return &apexEntryWrapper{w.Interface.WithError(err)}
}

type apexEntryWrapper struct {
	*log.Entry
}

func (w *apexEntryWrapper) WithField(k string, v interface{}) Logger {
	return &apexEntryWrapper{w.Entry.WithField(k, v)}
}

func (w *apexEntryWrapper) WithFields(fields log.Fielder) Logger {
	return &apexEntryWrapper{w.Entry.WithFields(fields)}
}

func (w *apexEntryWrapper) WithError(err error) Logger {
	return &apexEntryWrapper{w.Entry.WithError(err)}
}

var logger Logger = noopLogger{}

// GetLogger returns the API Logger
func GetLogger() Logger {
	return logger
}

// SetLogger returns the API Logger
func SetLogger(log Logger) {
	logger = log
}

// noopLogger just does nothing
type noopLogger struct{}

func (l noopLogger) Debug(msg string)                     {}
func (l noopLogger) Info(msg string)                      {}
func (l noopLogger) Warn(msg string)                      {}
func (l noopLogger) Error(msg string)                     {}
func (l noopLogger) Fatal(msg string)                     {}
func (l noopLogger) Debugf(msg string, v ...interface{})  {}
func (l noopLogger) Infof(msg string, v ...interface{})   {}
func (l noopLogger) Warnf(msg string, v ...interface{})   {}
func (l noopLogger) Errorf(msg string, v ...interface{})  {}
func (l noopLogger) Fatalf(msg string, v ...interface{})  {}
func (l noopLogger) WithField(string, interface{}) Logger { return l }
func (l noopLogger) WithFields(log.Fielder) Logger        { return l }
func (l noopLogger) WithError(error) Logger               { return l }
