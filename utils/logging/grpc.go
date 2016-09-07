package logging

import (
	"fmt"
	"strings"

	"google.golang.org/grpc/grpclog"

	"github.com/apex/log"
)

// NewGRPCLogger wraps the ctx and returns a grpclog.Logger
func NewGRPCLogger(ctx log.Interface) grpclog.Logger {
	return &gRPCLogger{ctx}
}

// gRPCLogger implements the grpc/grpclog.Logger interface
type gRPCLogger struct {
	ctx log.Interface
}

var filteredLogs = []string{
	"failed to complete security handshake",
}

func (l *gRPCLogger) shouldLog(log string) bool {
	for _, filter := range filteredLogs {
		if strings.Contains(log, filter) {
			return false
		}
	}
	return true
}

// Fatal implements the grpc/grpclog.Logger interface
func (l *gRPCLogger) Fatal(args ...interface{}) {
	l.fatalln(fmt.Sprint(args...))
}

// Fatalf implements the grpc/grpclog.Logger interface
func (l *gRPCLogger) Fatalf(format string, args ...interface{}) {
	l.fatalln(fmt.Sprintf(format, args...))
}

// Fatalln implements the grpc/grpclog.Logger interface
func (l *gRPCLogger) Fatalln(args ...interface{}) {
	l.fatalln(fmt.Sprint(args...))
}

func (l *gRPCLogger) fatalln(in string) {
	l.ctx.Error(in)
}

// Print implements the grpc/grpclog.Logger interface
func (l *gRPCLogger) Print(args ...interface{}) {
	l.println(fmt.Sprint(args...))
}

// Printf implements the grpc/grpclog.Logger interface
func (l *gRPCLogger) Printf(format string, args ...interface{}) {
	l.println(fmt.Sprintf(format, args...))
}

// Println implements the grpc/grpclog.Logger interface
func (l *gRPCLogger) Println(args ...interface{}) {
	l.println(fmt.Sprint(args...))
}

func (l *gRPCLogger) println(in string) {
	if l.shouldLog(in) {
		l.ctx.Debug(in)
	}
}
