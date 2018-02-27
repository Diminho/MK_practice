package simplelog

//EXAMPES:
// logger.WithFields(simplelog.Fields{"user": "JohnDoe", "file": "kaboom.txt"}).Info("FIRST")
// logger.WithField("user", "JohnDoe").Info("FIRST")
// logger.Info("JUST MESSAGE")

import (
	"fmt"
	"io"
	stdlog "log"
	"os"
	"sync"
	"time"
)

type Handler interface {
	HandleLog(*Record) error
}

var Now = time.Now

type Log struct {
	mu             sync.Mutex // ensures atomic writes; protects the following fields
	Out            io.Writer  // destination for output
	buf            []byte     // for accumulating text to write
	isolationLevel Level      // level from what logging should perform
	handler        Handler
}

func (l *Log) SetHandler(h Handler) {
	l.handler = h
}

func (l *Log) SetLevel(lvl Level) {
	l.isolationLevel = lvl
}

// Fields represents a map of entry level data used for structured logging.
type Fields map[string]interface{}

// Record represents a single log entry.
type Record struct {
	Level     Level     `json:"level"`
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
	Fields    Fields    `json:"fields"`
}

var log *Log

func NewLog(out io.Writer) *Log {
	log = &Log{
		Out:            out,
		isolationLevel: WarnLevel,
		handler:        NewJSONHandler(out), //default hadler
	}
	return log
}

func NewRecord() *Record {
	return &Record{
		Fields: Fields{},
	}
}

func (l *Log) WithField(key string, value interface{}) *Record {
	record := l.WithFields(Fields{key: value})
	return record
}

func (l *Log) WithFields(fields Fields) *Record {
	record := &Record{
		Fields: fields,
	}

	return record
}

func (r *Record) Info(message interface{}) {
	log.log(InfoLevel, r, fmt.Sprintf("%v", message))
}

//log just message without fields
func (l *Log) Info(message interface{}) {

	log.log(InfoLevel, NewRecord(), fmt.Sprintf("%v", message))
}

func (r *Record) Trace(message interface{}) {
	log.log(TraceLevel, r, fmt.Sprintf("%v", message))
}

//log just message without fields
func (l *Log) Trace(message interface{}) {
	log.log(TraceLevel, NewRecord(), fmt.Sprintf("%v", message))
}

func (r *Record) Debug(message interface{}) {
	log.log(DebugLevel, r, fmt.Sprintf("%v", message))
}

//log just message without fields
func (l *Log) Debug(message interface{}) {
	log.log(DebugLevel, NewRecord(), fmt.Sprintf("%v", message))
}

func (r *Record) Warn(message interface{}) {
	log.log(WarnLevel, r, fmt.Sprintf("%v", message))
}

//log just message without fields
func (l *Log) Warn(message interface{}) {
	log.log(WarnLevel, NewRecord(), fmt.Sprintf("%v", message))
}

func (r *Record) Error(err interface{}) {
	log.log(ErrorLevel, r, fmt.Sprintf("%v", err))
}

//log just message without fields
func (l *Log) Error(err interface{}) {
	log.log(ErrorLevel, NewRecord(), fmt.Sprintf("%v", err))
}

func (r *Record) Fatal(err interface{}) {
	log.log(FatalLevel, r, fmt.Sprintf("%v", err))
	os.Exit(1)
}

//log just message without fields
func (l *Log) Fatal(err interface{}) {
	log.log(FatalLevel, NewRecord(), fmt.Sprintf("%v", err))
	os.Exit(1)
}

func (l *Log) log(lvl Level, r *Record, msg string) {
	if l.isolationLevel <= lvl {
		if err := l.handler.HandleLog(r.prepare(lvl, msg)); err != nil {
			stdlog.Println("error logging %v", err)
		}
	}
}

func (r *Record) prepare(lvl Level, msg string) *Record {
	r.Message = msg
	r.Level = lvl
	r.Timestamp = Now()
	return r
}
