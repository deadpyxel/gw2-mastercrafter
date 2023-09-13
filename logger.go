package main

import (
	"log"
	"log/slog"
	"os"
)

var logger = NewSLogLogger()

type SLogLogger struct {
	log *slog.Logger
}

func NewSLogLogger() *SLogLogger {
	sl := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	return &SLogLogger{log: sl}
}

func (sl *SLogLogger) Debug(msg string, metadata ...interface{}) {
	sl.log.Debug(msg, metadata...)
}

func (sl *SLogLogger) Info(msg string, metadata ...interface{}) {
	sl.log.Info(msg, metadata...)
}

func (sl *SLogLogger) Warn(msg string, metadata ...interface{}) {
	sl.log.Warn(msg, metadata...)
}

func (sl *SLogLogger) Error(msg string, metadata ...interface{}) {
	sl.log.Error(msg, metadata...)
}

func (sl *SLogLogger) Fatal(msg string, metadata ...interface{}) {
	sl.log.Error(msg, metadata...)
	log.Fatal(msg)
}
