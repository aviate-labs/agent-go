package agent

type Logger interface {
	Printf(format string, v ...any)
}

type defaultLogger struct{}

func (l defaultLogger) Printf(format string, v ...any) {}
