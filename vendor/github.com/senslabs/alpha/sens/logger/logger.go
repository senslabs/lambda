package logger

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/fluent/fluent-logger-golang/fluent"
)

type Logger interface {
	Debug(v ...interface{})
	Error(v ...interface{})
	Debugf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
}

type ConsoleLogger struct{}
type FileLogger struct{}
type FluentLogger struct {
	Logger *fluent.Fluent
	Tag    string
}

var l Logger

func InitConsoleLogger() {
	l = &ConsoleLogger{}
}

func InitFileLogger(name string) {
	err := os.MkdirAll("/var/sens/logs", 0666)
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.OpenFile("/var/sens/logs/"+name+".log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()
	log.SetOutput(f)
	log.SetFlags(log.Lshortfile)
	l = &FileLogger{}
}

func InitFluentLogger(name string) {
	fL, err := fluent.New(fluent.Config{
		FluentHost: os.Getenv("FLUENTD_HOST"),
		FluentPort: 24224,
		RequestAck: true,
		RetryWait:  32,
		MaxRetry:   3,
	})
	if err != nil {
		log.Fatal(err)
	}

	l = &FluentLogger{Logger: fL, Tag: name}
}

// -- INIT ENDS HERE -- //

func (this *ConsoleLogger) Error(v ...interface{}) {
	log.Println(v...)
}

func (this *ConsoleLogger) Debug(v ...interface{}) {
	log.Println(v...)
}

func (this *ConsoleLogger) Errorf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func (this *ConsoleLogger) Debugf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func (this *FileLogger) Error(v ...interface{}) {
	log.Println(v...)
}

func (this *FileLogger) Debug(v ...interface{}) {
	log.Println(v...)
}

func (this *FileLogger) Errorf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func (this *FileLogger) Debugf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func (this *FluentLogger) Error(v ...interface{}) {
	logLevel := "ERROR"
	logMessage := fmt.Sprint(v...)
	fluentMessage := createFluentMessage(logLevel, logMessage)
	err := this.Logger.PostWithTime(this.Tag, time.Now(), fluentMessage)
	if err != nil {
		log.Println(err)
	}
}

func (this *FluentLogger) Debug(v ...interface{}) {
	logLevel := "DEBUG"
	logMessage := fmt.Sprint(v...)
	fluentMessage := createFluentMessage(logLevel, logMessage)
	err := this.Logger.PostWithTime(this.Tag, time.Now(), fluentMessage)
	if err != nil {
		log.Println(err)
	}
}

func createFluentMessage(level string, message string) map[string]string {
	fluentMessage := make(map[string]string, 0)
	pc, file, line, ok := runtime.Caller(3)
	if ok {
		fluentMessage = map[string]string{
			"level":    level,
			"file":     file,
			"line":     strconv.Itoa(line),
			"function": runtime.FuncForPC(pc).Name(),
		}
	}
	fluentMessage["message"] = message
	return fluentMessage
}

func (this *FluentLogger) Errorf(format string, v ...interface{}) {
	logLevel := "ERROR"
	logMessage := fmt.Sprintf(format, v...)
	fluentMessage := createFluentMessage(logLevel, logMessage)
	err := this.Logger.PostWithTime(this.Tag, time.Now(), fluentMessage)
	if err != nil {
		log.Println(err)
	}
}

func (this *FluentLogger) Debugf(format string, v ...interface{}) {
	logLevel := "DEBUG"
	logMessage := fmt.Sprintf(format, v...)
	fluentMessage := createFluentMessage(logLevel, logMessage)
	err := this.Logger.PostWithTime(this.Tag, time.Now(), fluentMessage)
	if err != nil {
		log.Println(err)
	}
}

func LogMeta(level string) {
	pc, file, line, ok := runtime.Caller(2)
	if ok {
		log.Printf("%s: %s:%d in func: %v", level, file, line, runtime.FuncForPC(pc).Name())
	}
}

func InitLogger(arg interface{}) {
	logStore := os.Getenv("LOG_STORE")
	if logStore == "file" {
		InitFileLogger(arg.(string))
	} else if logStore == "fluentd" {
		InitFluentLogger(arg.(string))
	} else {
		InitConsoleLogger()
	}
}

func Error(v ...interface{}) {
	LogMeta("Error")
	l.Error(v...)
}

func Debug(v ...interface{}) {
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "DEBUG" {
		LogMeta("Debug")
		l.Debug(v...)
	}
}

func Errorf(format string, v ...interface{}) {
	LogMeta("Error")
	l.Errorf(format, v...)
}

func Debugf(format string, v ...interface{}) {
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "DEBUG" {
		LogMeta("Debug")
		l.Debugf(format, v...)
	}
}
