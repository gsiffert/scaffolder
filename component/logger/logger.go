package logger

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Vorian-Atreides/scaffolder"
)

type Meta interface {
	WithMeta(map[string]interface{}) Logger
}

type Log interface {
	Printf(format string, args ...interface{})
}

var Std = log.New(os.Stderr, "", log.LstdFlags)

type Logger interface {
	Log

	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	With(name string, value interface{}) Logger
}

type Level uint8

const (
	Debug Level = iota
	Info
	Warning
	Error
)

// we use a backward list to be able to efficiently stack the new entry
// without modifying the parent reference.
type entry struct {
	name  string
	value interface{}
	prev  *entry
}

func (e *entry) fields() map[string]interface{} {
	m := make(map[string]interface{})

	for e != nil {
		if _, ok := m[e.name]; !ok {
			m[e.name] = e.value
		}
		e = e.prev
	}
	return m
}

type logger struct {
	level Level
	inner Log
	meta  *entry
}

func New(opts ...scaffolder.Option) Logger {
	logger := &logger{}
	scaffolder.Init(logger, opts...)
	return logger
}

func (l *logger) Default() {
	l.level = Debug
	l.inner = Std
}

func WithLevel(level Level) scaffolder.Option {
	return func(l *logger) error {
		l.level = level
		return nil
	}
}

func WithLog(log Log) scaffolder.Option {
	return func(l *logger) error {
		l.inner = log
		return nil
	}
}

func (l *logger) stringifyFields(fields map[string]interface{}) string {
	var builder strings.Builder
	for key, value := range fields {
		if builder.Len() > 0 {
			builder.WriteString(" ")
		}

		token := fmt.Sprintf("%s=%v", key, value)
		builder.WriteString(token)
	}
	return builder.String()
}

func (l *logger) printf(level Level, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	fields := l.meta.fields()
	if len(fields) == 0 {
		l.inner.Printf(format, args...)
		return
	}

	if m, ok := l.inner.(Meta); ok {
		m.WithMeta(fields).Printf(format, args...)
		return
	}
	stringifiedFields := l.stringifyFields(fields)
	l.inner.Printf(format+"\t%s", append(args, stringifiedFields)...)
}

func (l *logger) Printf(format string, args ...interface{}) {
	l.printf(Info, format, args...)
}

func (l *logger) Debugf(format string, args ...interface{}) {
	l.printf(Debug, format, args...)
}

func (l *logger) Infof(format string, args ...interface{}) {
	l.printf(Info, format, args...)
}

func (l *logger) Warnf(format string, args ...interface{}) {
	l.printf(Warning, format, args...)
}

func (l *logger) Errorf(format string, args ...interface{}) {
	l.printf(Error, format, args...)
}

func (l *logger) With(name string, value interface{}) Logger {
	cp := *l
	cp.meta = &entry{
		name:  name,
		value: value,
		prev:  l.meta,
	}
	return &cp
}
