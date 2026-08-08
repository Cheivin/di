package di

import (
	"io"
	"os"
)

// Log 是容器的日志接口。
// 容器内部所有日志（包括 Fatal）都通过它输出。
// v0.4.0 起 Fatal 签名由 string 改为 error，且实现应 panic 而非 os.Exit，
// 以便调用方 recover 后用 errors.Is 判断具体错误。
type Log interface {
	DebugMode(bool)
	Debug(string)
	Info(string)
	Warn(string)
	Fatal(error)
}

type logger struct {
	debugMode bool
	writer    io.Writer
	errWriter io.Writer
}

func stdLogger() Log {
	return &logger{
		debugMode: false,
		writer:    os.Stdout,
		errWriter: os.Stderr,
	}
}

func (l *logger) DebugMode(b bool) {
	l.debugMode = b
}

func (l *logger) Debug(s string) {
	if !l.debugMode {
		return
	}
	_, _ = l.writer.Write([]byte("[DI-DEBUG] : " + s + "\n"))
}

func (l *logger) Info(s string) {
	_, _ = l.writer.Write([]byte("[DI-INFO] : " + s + "\n"))
}

func (l *logger) Warn(s string) {
	_, _ = l.errWriter.Write([]byte("[DI-WARN] : " + s + "\n"))
}

func (l *logger) Fatal(err error) {
	_, _ = l.errWriter.Write([]byte("[DI-FATAL] : " + err.Error() + "\n"))
	panic(err)
}
