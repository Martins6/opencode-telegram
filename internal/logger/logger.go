package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Level string

const (
	INPUT  Level = "INPUT"
	OUTPUT Level = "OUTPUT"
	DEBUG  Level = "DEBUG"
	ERROR  Level = "ERROR"
)

type Logger struct {
	mu        sync.Mutex
	workspace string
	dateStr   string
	file      *os.File
	logger    *log.Logger
}

var (
	globalLogger *Logger
	once         sync.Once
)

func Initialize(workspace string) error {
	var err error
	once.Do(func() {
		globalLogger = &Logger{
			workspace: workspace,
			dateStr:   time.Now().Format("2006-01-02"),
		}
		err = globalLogger.setupLogFile()
		if err == nil {
			globalLogger.CleanupOldLogs()
		}
	})
	return err
}

func (l *Logger) setupLogFile() error {
	logsDir := filepath.Join(l.workspace, ".logs")
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return err
	}

	logFile := filepath.Join(logsDir, l.dateStr+".log")
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	l.file = f
	multiWriter := io.MultiWriter(os.Stdout, f)
	l.logger = log.New(multiWriter, "", 0)

	return nil
}

func (l *Logger) checkDateChange() {
	newDateStr := time.Now().Format("2006-01-02")
	if newDateStr != l.dateStr {
		l.mu.Lock()
		if newDateStr != l.dateStr {
			if l.file != nil {
				l.file.Close()
			}
			l.dateStr = newDateStr
			l.setupLogFile()
		}
		l.mu.Unlock()
	}
}

func Log(level Level, userID int64, message string) {
	if globalLogger == nil {
		return
	}

	globalLogger.checkDateChange()

	if globalLogger.logger == nil {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	format := fmt.Sprintf("[%s]  [%s] User %d: %s", level, timestamp, userID, message)
	globalLogger.logger.Println(format)
}

func LogDebug(format string, args ...interface{}) {
	if globalLogger == nil {
		return
	}

	globalLogger.checkDateChange()

	if globalLogger.logger == nil {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, args...)
	format = fmt.Sprintf("[%s]  [%s] %s", DEBUG, timestamp, message)
	globalLogger.logger.Println(format)
}

func (l *Logger) CleanupOldLogs() {
	go func() {
		logsDir := filepath.Join(l.workspace, ".logs")
		files, err := os.ReadDir(logsDir)
		if err != nil {
			return
		}

		cutoff := time.Now().AddDate(0, 0, -30)
		for _, f := range files {
			if f.IsDir() {
				continue
			}
			info, err := f.Info()
			if err != nil {
				continue
			}
			if info.ModTime().Before(cutoff) {
				os.Remove(filepath.Join(logsDir, f.Name()))
			}
		}
	}()
}

func Close() {
	if globalLogger != nil && globalLogger.file != nil {
		globalLogger.file.Close()
	}
}
