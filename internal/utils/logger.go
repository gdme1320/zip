package utils

import (
	"errors"
	"fmt"
	"log"
	"os"
)

// LogLevel 日志级别
type LogLevel int

const (
	// Quiet 静默模式，只输出错误
	Quiet LogLevel = iota
	// Normal 正常模式，输出重要信息
	Normal
	// Verbose 详细模式，输出所有信息
	Verbose
)

var (
	currentLevel = Normal
	infoLogger   *log.Logger
	errorLogger  *log.Logger
	debugLogger  *log.Logger
)

// InitLogger 初始化日志系统
func InitLogger(level LogLevel) {
	currentLevel = level

	// 错误日志总是输出到 stderr
	errorLogger = log.New(os.Stderr, "ERROR: ", log.LstdFlags)

	// 信息日志输出到 stdout
	infoLogger = log.New(os.Stdout, "", log.LstdFlags)

	// 调试日志输出到 stdout
	debugLogger = log.New(os.Stdout, "DEBUG: ", log.LstdFlags)
}

// Info 输出信息日志（Normal 和 Verbose 级别）
func Info(format string, v ...interface{}) {
	if currentLevel >= Normal {
		infoLogger.Printf(format, v...)
	}
}

// Debug 输出调试日志（仅 Verbose 级别）
func Debug(format string, v ...interface{}) {
	if currentLevel >= Verbose {
		debugLogger.Printf(format, v...)
	}
}

// Error 输出错误日志（所有级别，输出到 stderr）
func Error(format string, v ...interface{}) {
	errorLogger.Printf(format, v...)
}

// Errorf 输出格式化的错误日志（所有级别，输出到 stderr）
func Errorf(format string, v ...interface{}) error {
	msg := fmt.Sprintf(format, v...)
	errorLogger.Print(msg)
	return errors.New(msg)
}

// Fatal 输出致命错误并退出（所有级别，输出到 stderr）
func Fatal(format string, v ...interface{}) {
	errorLogger.Fatalf(format, v...)
}

// GetCurrentLevel 获取当前日志级别
func GetCurrentLevel() LogLevel {
	return currentLevel
}
