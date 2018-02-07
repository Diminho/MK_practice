package simplelog

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
