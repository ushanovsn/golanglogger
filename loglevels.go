package golanglogger

import (
	"errors"
	"strings"
)

// logger level enums
type LoggingLevel int

const (
	DebugLvl LoggingLevel = iota
	InfoLvl
	WarningLvl
	ErrorLvl
)

// returns LoggingLevel const by text value (initial level if wrong name)
func LoggingLevelValue(s string) (LoggingLevel, error) {
	if l, ok := map[string]LoggingLevel{"debug": DebugLvl, "info": InfoLvl, "warning": WarningLvl, "error": ErrorLvl}[strings.ToLower(s)]; ok {
		return l, nil
	}
	return initLogLevel, errors.New("Wrong LogLevel value: " + s)
}

// returns LoggingLevel name
func (l LoggingLevel) Name() string {
	lvlNames := []string{"Debug", "Info", "Warning", "Error"}
	if int(l) > (len(lvlNames) - 1) {
		return ""
	}
	return strings.ToUpper(lvlNames[int(l)])
}
