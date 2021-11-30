package di

import (
	"io"
	"os"
)

type Log interface {
	DebugMode(bool)
	Debug(string)
	Info(string)
	Warn(string)
	Fatal(string)
}

type logger struct {
	debugMode bool
	writer    io.Writer
	errWriter io.Writer
}

func stdLogger() Log {
	return logger{
		debugMode: false,
		writer:    os.Stdout,
		errWriter: os.Stderr,
	}
}

func (l logger) DebugMode(b bool) {
	l.debugMode = b
}

func (l logger) Debug(s string) {
	if !l.debugMode {
		return
	}
	_, _ = l.writer.Write([]byte("[DI-DEBUG] : " + s + "\n"))
}

func (l logger) Info(s string) {
	_, _ = l.writer.Write([]byte("[DI-INFO] : " + s + "\n"))
}

func (l logger) Warn(s string) {
	_, _ = l.errWriter.Write([]byte("[DI-WARN] : " + s + "\n"))
}

func (l logger) Fatal(s string) {
	_, _ = l.errWriter.Write([]byte("[DI-FATAL] : " + s + "\n"))
	os.Exit(1)
}
