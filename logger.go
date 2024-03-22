package agent

type Logger interface {
	Printf(format string, v ...interface{})
}

type defaultLogger struct{}

func (l defaultLogger) Printf(format string, v ...interface{}) {}
