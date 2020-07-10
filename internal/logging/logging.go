package logging

import (
	"fmt"
	"io"

	"github.com/logrusorgru/aurora"
)

type Options struct {
	Silent       bool
	Colorized    bool
	ResultWriter io.Writer
}

type Logger struct {
	writer       io.Writer
	resultWriter io.Writer
	silent       bool
	a            aurora.Aurora
}

func NewLogger(writer io.Writer, options Options) *Logger {
	logger := Logger{
		writer: writer,
		silent: options.Silent,
		a:      aurora.NewAurora(options.Colorized),
	}

	if options.ResultWriter != nil {
		logger.resultWriter = options.ResultWriter
	} else {
		logger.resultWriter = writer
	}

	return &logger
}

func (l Logger) Infoln(msg string) {
	l.println(l.writer, l.a.Sprintf(l.a.Blue(msg)))
}

func (l Logger) Infof(format string, args ...interface{}) {
	l.println(l.writer, l.a.Sprintf(l.a.Blue(format), args...))
}

func (l Logger) Errorf(format string, args ...interface{}) {
	l.println(l.writer, l.a.Sprintf(l.a.Red(format), args...))
}

func (l Logger) Err(err error) {
	l.println(l.writer, l.a.Sprintf(l.a.Red("%v"), err))
}

func (l Logger) Resultf(format string, args ...interface{}) {
	fmt.Fprintln(l.resultWriter, l.a.Sprintf(l.a.Green(format), args...))
}

func (l Logger) println(writer io.Writer, msg string) {
	if l.silent {
		return
	}

	fmt.Fprintln(writer, msg)
}
