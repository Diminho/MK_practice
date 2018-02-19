package simplelog

//EXAMPES:
// logger.WithFields(simplelog.Fields{"user": "JohnDoe", "file": "kaboom.txt"}).Info("FIRST")
// logger.WithField("user", "JohnDoe").Info("FIRST")
// logger.Info("JUST MESSAGE")

import (
	"io"
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

func (r *Record) Info(message string) {
	log.log(InfoLevel, r, message)
}

//log just message without fields
func (l *Log) Info(message string) {

	log.log(InfoLevel, NewRecord(), message)
}

func (r *Record) Trace(message string) {
	log.log(TraceLevel, r, message)
}

//log just message without fields
func (l *Log) Trace(message string) {
	log.log(TraceLevel, NewRecord(), message)
}

func (r *Record) Debug(message string) {
	log.log(DebugLevel, r, message)
}

//log just message without fields
func (l *Log) Debug(message string) {
	log.log(DebugLevel, NewRecord(), message)
}

func (r *Record) Warn(message string) {
	log.log(WarnLevel, r, message)
}

//log just message without fields
func (l *Log) Warn(message string) {
	log.log(WarnLevel, NewRecord(), message)
}

func (r *Record) Error(err error) {
	log.log(ErrorLevel, r, err.Error())
}

//log just message without fields
func (l *Log) Error(err error) {
	log.log(ErrorLevel, NewRecord(), err.Error())
}

func (r *Record) Fatal(err error) {
	log.log(FatalLevel, r, err.Error())
	os.Exit(1)
}

//log just message without fields
func (l *Log) Fatal(err error) {
	log.log(FatalLevel, NewRecord(), err.Error())
	os.Exit(1)
}

func (l *Log) log(lvl Level, r *Record, msg string) {
	if l.isolationLevel <= lvl {
		l.handler.HandleLog(r.prepare(lvl, msg))
	}

}

func (r *Record) prepare(lvl Level, msg string) *Record {
	r.Message = msg
	r.Level = lvl
	r.Timestamp = Now()
	return r
}
