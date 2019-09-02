/*
Package logger define a Logger component which can be used to wrap
the logger in your application.

It extends the printing role from traditional logger with the ability
to stack meta-data with entry log while maintaining an immutable state.
*/
package logger

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Vorian-Atreides/scaffolder"
)

// MetaPrinter define the interface for logger which allows
// you to attach meta-data over log entry.
type MetaPrinter interface {
	WithMeta(map[string]interface{}) Logger
}

// Printer define the inner logger to be used to write the log.
type Printer interface {
	Printf(format string, args ...interface{})
}

var std = log.New(os.Stderr, "", log.LstdFlags)

// Logger interface exported from this package.
type Logger interface {
	Printer

	// Debugf write a debug level log.
	Debugf(format string, args ...interface{})
	// Infof write an info level log.
	Infof(format string, args ...interface{})
	// Warnf write a warning level log.
	Warnf(format string, args ...interface{})
	// Errorf write an error level log.
	Errorf(format string, args ...interface{})
	// With attach a new meta data field or replace an existing one.
	With(name string, value interface{}) Logger
}

// Level define the verbosity level.
type Level uint8

const (
	// Debug for debug log.
	Debug Level = iota
	// Info for info log.
	Info
	// Warning for warning log.
	Warning
	// Error for error log.
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
	inner Printer
	meta  *entry
}

// New instantiate a new Logger and customize it with the given options.
func New(opts ...scaffolder.Option) Logger {
	logger := &logger{}
	scaffolder.Init(logger, opts...)
	return logger
}

func (l *logger) Default() {
	l.level = Debug
	l.inner = std
}

// WithLevel set the verbosity level to the new Logger.
func WithLevel(level Level) scaffolder.Option {
	return func(l *logger) error {
		l.level = level
		return nil
	}
}

// WithPrinter set the inner Printer to be used to write the log.
func WithPrinter(printer Printer) scaffolder.Option {
	return func(l *logger) error {
		l.inner = printer
		if logger, ok := printer.(*logger); ok {
			l.inner = logger.inner
		}
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

	if m, ok := l.inner.(MetaPrinter); ok {
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
