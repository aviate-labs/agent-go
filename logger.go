package agent

type Logger interface {
	Printf(format string, v ...any)
}

type NoopLogger struct{}

func (l NoopLogger) Printf(format string, v ...any) {}
