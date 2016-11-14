package api

import (
	"fmt"

	logrus "github.com/Sirupsen/logrus"
	apex "github.com/apex/log"
	"google.golang.org/grpc/grpclog"
)

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
	WithFields(apex.Fielder) Logger
	WithError(error) Logger
}

// GRPC logger
func GRPC() grpclog.Logger {
	return &grpcWrapper{GetLogger()}
}

type grpcWrapper struct {
	Logger
}

func (w *grpcWrapper) Fatal(args ...interface{}) {
	w.Logger.Fatal(fmt.Sprint(args...))
}
func (w *grpcWrapper) Fatalln(args ...interface{}) {
	w.Fatal(args...)
}
func (w *grpcWrapper) Print(args ...interface{}) {
	w.Logger.Debug(fmt.Sprint(args...))
}
func (w *grpcWrapper) Printf(format string, args ...interface{}) {
	w.Logger.Debugf(format, args...)
}
func (w *grpcWrapper) Println(args ...interface{}) {
	w.Print(args...)
}

// StandardLogrus wraps the standard Logrus Logger into a Logger
func StandardLogrus() Logger {
	return Logrus(logrus.StandardLogger())
}

// Logrus wraps logrus into a Logger
func Logrus(logger *logrus.Logger) Logger {
	return &logrusEntryWrapper{logrus.NewEntry(logger)}
}

type logrusEntryWrapper struct {
	*logrus.Entry
}

func (w *logrusEntryWrapper) Debug(msg string) {
	w.Entry.Debug(msg)
}

func (w *logrusEntryWrapper) Info(msg string) {
	w.Entry.Info(msg)
}

func (w *logrusEntryWrapper) Warn(msg string) {
	w.Entry.Warn(msg)
}

func (w *logrusEntryWrapper) Error(msg string) {
	w.Entry.Error(msg)
}

func (w *logrusEntryWrapper) Fatal(msg string) {
	w.Entry.Fatal(msg)
}

func (w *logrusEntryWrapper) WithError(err error) Logger {
	return &logrusEntryWrapper{w.Entry.WithError(err)}
}

func (w *logrusEntryWrapper) WithField(k string, v interface{}) Logger {
	return &logrusEntryWrapper{w.Entry.WithField(k, v)}
}

func (w *logrusEntryWrapper) WithFields(fields apex.Fielder) Logger {
	return &logrusEntryWrapper{w.Entry.WithFields(
		map[string]interface{}(fields.Fields()),
	)}
}

// Apex wraps apex/log
func Apex(ctx apex.Interface) Logger {
	return &apexInterfaceWrapper{ctx}
}

type apexInterfaceWrapper struct {
	apex.Interface
}

func (w *apexInterfaceWrapper) WithField(k string, v interface{}) Logger {
	return &apexEntryWrapper{w.Interface.WithField(k, v)}
}

func (w *apexInterfaceWrapper) WithFields(fields apex.Fielder) Logger {
	return &apexEntryWrapper{w.Interface.WithFields(fields)}
}

func (w *apexInterfaceWrapper) WithError(err error) Logger {
	return &apexEntryWrapper{w.Interface.WithError(err)}
}

type apexEntryWrapper struct {
	*apex.Entry
}

func (w *apexEntryWrapper) WithField(k string, v interface{}) Logger {
	return &apexEntryWrapper{w.Entry.WithField(k, v)}
}

func (w *apexEntryWrapper) WithFields(fields apex.Fielder) Logger {
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

// SetLogger sets the API and gRPC Logger
func SetLogger(log Logger) {
	logger = log
	grpclog.SetLogger(GRPC())
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
func (l noopLogger) WithFields(apex.Fielder) Logger       { return l }
func (l noopLogger) WithError(error) Logger               { return l }
