package logger

import (
	"fmt"
	"io"
	"log"
	"os"
)

const (
	Sys = "SYSTEM"
	Upl = "UPLOAD"
	Hom = "HOME"
	Del = "DELETE"
	Sto = "STORAGE"
	Mid = "MIDDLEWARE"
	Dat = "DATABASE"
	Jan = "JANITOR"

	BLUE = "\033[34m"
	BLUE_BOLD = "\033[34;1m"
	YELLOW = "\033[33m"
	YELLOW_BOLD = "\033[33;1m"
	RED = "\033[31m"
	RED_BOLD = "\033[31;1m"
	BOLD = "\033[1m"
	RESET = "\033[0m"
)

type Logger struct {
	logger *log.Logger
	//ErrorLogger *log.Logger
	w io.Writer
	file *os.File
}

func NewLogger(logFilePath string) (*Logger, error) {
	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &Logger{
		logger: log.New(os.Stdout, "[\033[34;1mSYSTEM INFO\033[0m] ", log.Ldate|log.Ltime|log.Lshortfile),
		//ErrorLogger: log.New(file, "[\033[31mSYSTEM ERR\033[0m] ", log.Ldate|log.Ltime|log.Lshortfile),
		file: file,
	}, nil
}

func (l *Logger) Info (s ...string) *Logger{
	var mod string
	if len(s) != 1 || s[0] == "" {
		mod = Sys
	} else {
		mod = s[0]
	}
	l.logger.SetPrefix(fmt.Sprintf("%s[%s%s INFO%s%s]%s ", BOLD, BLUE_BOLD, mod, RESET, BOLD, RESET))
	l.logger.SetOutput(os.Stdout)
	return l
}

func (l *Logger) Error (s ...string) *Logger {
	var mod string
	if len(s) != 1 || s[0] == ""{
		mod = Sys
	} else {
		mod = s[0]
	}
	l.logger.SetPrefix(fmt.Sprintf("%s[%s%s ERR%s%s]%s ", BOLD, RED_BOLD, mod, RESET, BOLD, RESET))
	//l.logger.SetPrefix(fmt.Sprintf("[\033[31;1m%s ERR\033[0m] ", mod))
	l.logger.SetOutput(l.file)
	return l
}

func (l *Logger) Warn(s ...string) *Logger {
	var mod string
	if len(s) != 1 || s[0] == ""{
		mod = Sys
	} else {
		mod = s[0]
	}
	l.logger.SetPrefix(fmt.Sprintf("%s[%s%s WARN%s%s]%s ", BOLD, YELLOW_BOLD, mod, RESET, BOLD, RESET))
	//l.logger.SetPrefix(fmt.Sprintf("\033[1m[\033[0m\033[33;1m%s WARN\033[0m] ", mod))
	l.logger.SetOutput(l.file)
	return l
}

func (l *Logger) Write(msg string) {
	l.logger.Output(2, fmt.Sprintf("%s", msg))
}

func (l *Logger) Writef(msg string, err error) {
	l.logger.Output(2, fmt.Sprintf("%s: %v", msg, err))
}

//func (l *Logger) Error(msg string, err error) {
//	l.ErrorLogger.Printf("%s: %v", msg, err)
//}

func (l *Logger) Close(){
	l.file.Close()
}
