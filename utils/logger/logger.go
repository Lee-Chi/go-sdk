package logger

import (
	"fmt"
	"path"
	"runtime"
	"time"
)

var (
	ignoreDir string = ""
)

func Init() {
	ignoreDir = getCurrentDir()
}

func getCurrentDir() string {
	_, file, _, ok := runtime.Caller(2)
	if !ok {
		return ""
	}

	return path.Dir(file)
}

const (
	Level_Error = "ERROR"
	Level_Warn  = "WARN"
	Level_Info  = "INFO"
	Level_Debug = "DEBUG"
)

func log(level string, format string, args ...interface{}) {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "???"
		line = 0
	}

	if ignoreDir != "" {
		file = file[len(ignoreDir)+1:]
	}

	fmt.Printf("%s [%s] %s:%d %s\n", time.Now().Format(time.RFC3339Nano), level, file, line, fmt.Sprintf(format, args...))
}

func Error(format string, args ...interface{}) {
	log(Level_Error, format, args...)
}

func Warn(format string, args ...interface{}) {
	log(Level_Warn, format, args...)
}

func Info(format string, args ...interface{}) {
	log(Level_Info, format, args...)
}

func Debug(format string, args ...interface{}) {
	log(Level_Debug, format, args...)
}
