package migrate

import (
	"fmt"

	"github.com/underbek/examples-go/logger"
)

type gooseLogger struct {
	internal *logger.Logger
}

func newLogger(log *logger.Logger) *gooseLogger {
	return &gooseLogger{
		internal: log.Named("goose").WithOptions(logger.AddCallerSkip(1)),
	}
}

func (l *gooseLogger) Fatal(v ...interface{}) {
	l.internal.Fatal(fmt.Sprint(v...))
}

func (l *gooseLogger) Fatalf(format string, v ...interface{}) {
	l.internal.Fatal(fmt.Sprintf(format, v...))
}

func (l *gooseLogger) Print(v ...interface{}) {
	l.internal.Info(fmt.Sprint(v...))
}

func (l *gooseLogger) Println(v ...interface{}) {
	l.internal.Info(fmt.Sprint(v...))
}

func (l *gooseLogger) Printf(format string, v ...interface{}) {
	l.internal.Info(fmt.Sprintf(format, v...))
}
