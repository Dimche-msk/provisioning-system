package logger

import (
	"log"
	"strings"
)

type Level int

const (
	LevelError Level = iota
	LevelWarn
	LevelInfo
	LevelDebug
)

var currentLevel = LevelError

func SetLevel(levelStr string) {
	switch strings.ToUpper(levelStr) {
	case "DEBUG":
		currentLevel = LevelDebug
	case "INFO":
		currentLevel = LevelInfo
	case "WARN", "WARNING":
		currentLevel = LevelWarn
	case "ERROR":
		currentLevel = LevelError
	default:
		currentLevel = LevelError
	}
}

func GetLevel() Level {
	return currentLevel
}

func Error(format string, v ...interface{}) {
	if currentLevel >= LevelError {
		log.Printf("[ERROR] "+format, v...)
	}
}

func Warn(format string, v ...interface{}) {
	if currentLevel >= LevelWarn {
		log.Printf("[WARN] "+format, v...)
	}
}

func Info(format string, v ...interface{}) {
	if currentLevel >= LevelInfo {
		log.Printf("[INFO] "+format, v...)
	}
}

func Debug(format string, v ...interface{}) {
	if currentLevel >= LevelDebug {
		log.Printf("[DEBUG] "+format, v...)
	}
}
