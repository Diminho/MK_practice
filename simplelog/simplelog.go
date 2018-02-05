package simplelog

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

type HandlerFunc func(*Record) error

var Now = time.Now

type Log struct {
	Mu      sync.Mutex // ensures atomic writes; protects the following fields
	Out     io.Writer  // destination for output
	buf     []byte     // for accumulating text to write
	Handler HandlerFunc
}

func (l *Log) SetHandler(f HandlerFunc) {
	l.Handler = f
}

// Fields represents a map of entry level data used for structured logging.
type Fields map[string]interface{}

// Record represents a single log entry.
type Record struct {
	Log       *Log      `json:"-"`
	Level     Level     `json:"level"`
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
	Fields    Fields    `json:"fields"`
}

var log *Log

func NewLog(out io.Writer) *Log {
	return &Log{Out: out}
}

func NewRecord(l *Log) *Record {
	return &Record{
		Log:    l,
		Fields: Fields{},
	}
}

func (l *Log) WithField(key string, value interface{}) *Record {
	record := l.WithFields(Fields{key: value})
	return record
}

func (l *Log) WithFields(fields Fields) *Record {
	record := &Record{
		Log:    l,
		Fields: fields,
	}
	return record
}

func (r *Record) Info(message string) {
	r.Log.log(InfoLevel, r, message)
}

//log just message without fields
func (l *Log) Info(message string) {
	log.log(InfoLevel, NewRecord(l), message)
}

// Infof level formatted message.
func (l *Log) Infof(msg string, v ...interface{}) {
	log.log(InfoLevel, NewRecord(l), fmt.Sprintf(msg, v...))
}

func (r *Record) Trace(message string) {
	r.Log.log(TraceLevel, r, message)
}

//log just message without fields
func (l *Log) Trace(message string) {
	log.log(TraceLevel, NewRecord(l), message)
}

func (r *Record) Debug(message string) {
	r.Log.log(DebugLevel, r, message)
}

//log just message without fields
func (l *Log) Debug(message string) {
	log.log(DebugLevel, NewRecord(l), message)
}

func (r *Record) Error(err error) {
	r.Log.log(ErrorLevel, r, err.Error())
}

//log just message without fields
func (l *Log) Error(err error) {
	log.log(ErrorLevel, NewRecord(l), err.Error())
}

func (r *Record) Fatal(err error) {

	r.Log.log(FatalLevel, r, err.Error())
	os.Exit(1)
}

//log just message without fields
func (l *Log) Fatal(err error) {
	log.log(FatalLevel, NewRecord(l), err.Error())
	os.Exit(1)
}

func (l *Log) log(lvl Level, r *Record, msg string) {
	l.Handler(r.prepare(lvl, msg))
}

func (r *Record) prepare(lvl Level, msg string) *Record {
	r.Message = msg
	r.Level = lvl
	r.Timestamp = Now()
	return r
}
