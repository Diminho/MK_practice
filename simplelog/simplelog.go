package simplelog

import (
	"io"
	"log"
	"sync"
	"time"
)

var (
	Trace   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

type HandlerFunc func(*Record) error

type Handler interface {
	HandleLog(*Record) error
}

type Logger struct {
	Mu sync.Mutex // ensures atomic writes; protects the following fields
	// flag   int        // properties
	Out     io.Writer // destination for output
	buf     []byte    // for accumulating text to write
	Handler HandlerFunc
}

func (logger *Logger) SetHandler(f HandlerFunc) {
	logger.Handler = f
}

// type Logger interface {
// 	Log()
// }

type Level int

// Log levels.
const (
	TraceLevel Level = iota + 1
	DebugLevel
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

var levelNames = [...]string{
	TraceLevel: "trace",
	DebugLevel: "debug",
	InfoLevel:  "info",
	WarnLevel:  "warn",
	ErrorLevel: "error",
	FatalLevel: "fatal",
}

// String implementation.
func (l Level) String() string {
	return levelNames[l]
}

// MarshalJSON implementation.
func (l Level) MarshalJSON() ([]byte, error) {
	return []byte(`"` + l.String() + `"`), nil
}

// Fields represents a map of entry level data used for structured logging.
type Fields map[string]interface{}

// Record represents a single log entry.
type Record struct {
	Logger    *Logger   `json:"-"`
	Level     Level     `json:"level"`
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
	// Fields    []Fields  `json:"fields"`
	Fields Fields `json:"fields"`
}

// var logger = NewLogger(os.Stderr, "")
var logger *Logger

// func init()

func NewLogger(out io.Writer) *Logger {
	return &Logger{Out: out}
}

func NewRecord(logger *Logger) *Record {
	return &Record{
		Logger: logger,
	}
}

// func (logger *Logger) WithField(key string, value interface{}) *Record {
// 	record.Fields = Fields{key: value}
// 	return record
// }

func (logger *Logger) WithFields(fields Fields) *Record {
	record := &Record{
		Logger: logger,
		Fields: fields,
	}
	return record
}

func (record *Record) Info(message string) {
	record.Logger.log(InfoLevel, record, message)
}

func (logger *Logger) log(level Level, record *Record, message string) {
	logger.Handler(record.prepare(level, message))
}

func (record *Record) prepare(level Level, message string) *Record {
	record.Message = message
	record.Level = level
	record.Timestamp = time.Now()
	return record
}
