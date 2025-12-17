package qlog

import (
	"github.com/sirupsen/logrus"
	"io"
)

var logger = New()

func SetOutput(out io.Writer) {
	logger.SetOutput(out)
}

func GetOutput() io.Writer {
	return logger.GetOutput()
}

func SetLevel(level logrus.Level) {
	logger.SetLevel(level)
}

func GetLevel() logrus.Level {
	return logger.GetLevel()
}

func Trace(args ...interface{}) {
	logger.Trace(args...)
}

func Tracef(format string, args ...interface{}) {
	logger.Tracef(format, args...)
}

func Debug(args ...interface{}) {
	logger.Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	logger.Debugf(format, args...)
}

func Info(args ...interface{}) {
	logger.Info(args...)
}

func Infof(format string, args ...interface{}) {
	logger.Infof(format, args...)
}

func Print(args ...interface{}) {
	logger.Print(args...)
}

func Printf(format string, args ...interface{}) {
	logger.Printf(format, args...)
}

func Warn(args ...interface{}) {
	logger.Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	logger.Warnf(format, args...)
}

func Error(args ...interface{}) {
	logger.Error(args...)
}

func Errorf(format string, args ...interface{}) {
	logger.Errorf(format, args...)
}

func Panicf(format string, args ...interface{}) {
	logger.Panicf(format, args...)
}

func WithFields(fields Fields) Logger {
	return logger.WithFields(fields)
}

func WithField(key string, value interface{}) Logger {
	return logger.WithField(key, value)
}

func WithError(err error) Logger {
	return logger.WithError(err)
}
