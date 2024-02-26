package golanglogger

import (
	"testing"
)



func TestLoggingLevelValue(t *testing.T){

	v, err := LoggingLevelValue("debug")
	if err != nil {
		t.Error("Error received: " + err.Error())
	}
	if v != Debug {
		t.Error("Expected \"1\", got", v)
	}
}