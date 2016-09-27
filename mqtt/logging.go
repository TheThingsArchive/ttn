package mqtt

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
}

// noopLogger just does nothing
type noopLogger struct{}

func (l noopLogger) Debug(msg string)                    {}
func (l noopLogger) Info(msg string)                     {}
func (l noopLogger) Warn(msg string)                     {}
func (l noopLogger) Error(msg string)                    {}
func (l noopLogger) Fatal(msg string)                    {}
func (l noopLogger) Debugf(msg string, v ...interface{}) {}
func (l noopLogger) Infof(msg string, v ...interface{})  {}
func (l noopLogger) Warnf(msg string, v ...interface{})  {}
func (l noopLogger) Errorf(msg string, v ...interface{}) {}
func (l noopLogger) Fatalf(msg string, v ...interface{}) {}
