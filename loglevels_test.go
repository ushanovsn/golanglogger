package golanglogger

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func Test_LoggingLevelValue(t *testing.T) {
	// values for testing
	lvls := []struct {
		testName string
		testLvlValue LoggingLevel
		testingNames []string
		
	} {
		{
			testName: "Test Debug level",
			testLvlValue: DebugLvl,
			testingNames: []string{"Debug", "DEBUG", "debug", "DeBuG"},
		},
		{
			testName: "Test Info level",
			testLvlValue: InfoLvl,
			testingNames: []string{"Info", "INFO", "info", "InFo"},
		},
		{
			testName: "Test Warning level",
			testLvlValue: WarningLvl,
			testingNames: []string{"Warning", "WARNING", "warning", "WaRnInG"},
		},
		{
			testName: "Test Error level",
			testLvlValue: ErrorLvl,
			testingNames: []string{"Error", "ERROR", "error", "ErRoR"},
		},
	}


	// aggregate counts of levels
	var cntLvls int
	for i:=0; i<=int(ErrorLvl); i++ {
		cntLvls++
	}

	t.Run("Check counts", func(t *testing.T) {
		assert.Equal(t, cntLvls, len(lvls))
	})


	for _, tt := range lvls {
		t.Run(tt.testName, func(t *testing.T){
			for _, s := range tt.testingNames {
				v, err := LoggingLevelValue(s)
				
				if assert.NoError(t, err) {
					assert.Equal(t, tt.testLvlValue, v)
				}
			}
		})
	}


	

	// bad values for testing
	testingNameIncorrect := []string {"Tebug", "", "big_text_value_for_test", "12345", "text with spaces"}


	for _, tt := range testingNameIncorrect {
		t.Run(("Check wrong values and errors: \"" + tt + "\""), func(t *testing.T) {
			_, err := LoggingLevelValue(tt)
			assert.Error(t, err)
		})
	}
}

func Test_Name(t *testing.T) {
	// values for testing
	lvls := []struct {
		testName string
		testLvlValue LoggingLevel
		testingName string
		
	} {
		{
			testName: "Test Debug level",
			testLvlValue: DebugLvl,
			testingName: "DEBUG",
		},
		{
			testName: "Test Info level",
			testLvlValue: InfoLvl,
			testingName:"INFO",
		},
		{
			testName: "Test Warning level",
			testLvlValue: WarningLvl,
			testingName: "WARNING",
		},
		{
			testName: "Test Error level",
			testLvlValue: ErrorLvl,
			testingName: "ERROR",
		},
	}


	// aggregate counts of levels
	var cntLvls int
	for i:=0; i<=int(ErrorLvl); i++ {
		cntLvls++
	}

	t.Run("Check count and max value", func(t *testing.T) {
		// cntLvls is more then levels count by 1
		s := LoggingLevel(cntLvls).Name()
		assert.Equal(t, "", s)
	})


	for _, tt := range lvls {
		t.Run(tt.testName, func(t *testing.T){
			s := LoggingLevel(tt.testLvlValue).Name()
			
			assert.Equal(t, tt.testingName, s)
		})
	}
}