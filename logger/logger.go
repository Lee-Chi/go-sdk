package logger

import (
	"fmt"
	"path"
	"runtime"
	"time"
)

var (
	skip      int    = 2
	ignoreDir string = ""
)

func Init() {
	ignoreDir = getCurrentDir()
}

func getCurrentDir() string {
	_, file, _, ok := runtime.Caller(skip)
	if !ok {
		return ""
	}

	return path.Dir(file)
}

const (
	Level_Error = "E"
	Level_Info  = "I"
	Level_Warn  = "W"
	Level_Debug = "D"
)

const (
	TimeLayout string = "2006-01-02 15:04:05.000"
)

func log(level string, format string, args ...interface{}) {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		file = "???"
		line = 0
	}

	if ignoreDir != "" {
		file = file[len(ignoreDir)+1:]
	}

	fmt.Printf("[%s] %s | %s:%d | %s\n", level, time.Now().UTC().Format(TimeLayout), file, line, fmt.Sprintf(format, args...))
}

func Printf(format string, args ...interface{}) {
	_, file, line, ok := runtime.Caller(1)

	if !ok {
		file = "???"
		line = 0
	}

	if ignoreDir != "" {
		file = file[len(ignoreDir)+1:]
	}

	fmt.Printf("%s | %s:%d | %s", time.Now().UTC().Format(TimeLayout), file, line, fmt.Sprintf(format, args...))
}

func Sprintf(format string, args ...interface{}) string {
	_, file, line, ok := runtime.Caller(1)

	if !ok {
		file = "???"
		line = 0
	}

	if ignoreDir != "" {
		file = file[len(ignoreDir)+1:]
	}

	return fmt.Sprintf("%s | %s:%d | %s", time.Now().UTC().Format(TimeLayout), file, line, fmt.Sprintf(format, args...))
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
