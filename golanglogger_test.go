package golanglogger

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func Test_New(t *testing.T) {
	// values for testing
	testLvls := []struct {
		testName     string
		testLvlValue LoggingLevel
		outStat      bool
		outDebugStat bool
		outInfoStat  bool
		outWarnStat  bool
		outErrorStat bool
	}{
		{
			testName:     "Test writing Debug level",
			testLvlValue: DebugLvl,
			outStat:      true,
			outDebugStat: true,
			outInfoStat:  true,
			outWarnStat:  true,
			outErrorStat: true,
		},
		{
			testName:     "Test writing Info level",
			testLvlValue: InfoLvl,
			outStat:      true,
			outDebugStat: false,
			outInfoStat:  true,
			outWarnStat:  true,
			outErrorStat: true,
		},
		{
			testName:     "Test writing Warning level",
			testLvlValue: WarningLvl,
			outStat:      true,
			outDebugStat: false,
			outInfoStat:  false,
			outWarnStat:  true,
			outErrorStat: true,
		},
		{
			testName:     "Test writing Error level",
			testLvlValue: ErrorLvl,
			outStat:      true,
			outDebugStat: false,
			outInfoStat:  false,
			outWarnStat:  false,
			outErrorStat: true,
		},
	}

	// temp file
	logFileName := uuid.NewString() + ".log"
	defer os.Remove(logFileName)

	// loger level
	logLevel := DebugLvl
	// delay for change control
	tDur := time.Millisecond * 300

	// creating logger
	log := New(logLevel, logFileName)
	defer log.StopLog()

	// check logger created
	assert.NotNil(t, log, "Check logger exists")
	if log == nil {
		t.Fatal()
	}

	t.Run("Check logger creating", func(t *testing.T) {
		logCon, logErr, logFile := log.CurrentOutParams()
		lvl := log.CurrentLevel()
		assert.True(t, logCon, "Console out enabled")
		assert.False(t, logErr, "stdError out enabled")
		assert.Equal(t, logFileName, logFile, "Logger file out")
		assert.Equal(t, logLevel, lvl, "Logger level")
		assert.FileExists(t, logFileName, "Logger file")
	})

	// check file
	f, err := os.Stat(logFileName)
	assert.NoError(t, err, "Fix starting file size")
	if err != nil {
		t.Fatal()
	}
	fileSizeBefore := f.Size()
	fileSizeAfter := f.Size()

	for _, tt := range testLvls {
		t.Run(tt.testName, func(t *testing.T) {
			// set level and check setted level
			logLevel = tt.testLvlValue
			log.SetLevel(logLevel)
			lvl := log.CurrentLevel()
			assert.Equal(t, logLevel, lvl, "Logger level while test writing")
			if logLevel != lvl {
				t.Fatal()
			}
			// wait untill system messages will writing
			time.Sleep(tDur)

			//*********************************** OUT *******************************************
			// save file param
			f, err = os.Stat(logFileName)
			assert.NoError(t, err, "File control while test writing - before stage")
			if err != nil {
				t.Fatal()
			}
			fileSizeBefore = f.Size()
			// write log
			log.Out("Test message: Out")
			// wait untill messages writing
			time.Sleep(tDur)
			f, err = os.Stat(logFileName)
			assert.NoError(t, err, "File control while test writing - after stage")
			if err != nil {
				t.Fatal()
			}
			fileSizeAfter = f.Size()
			// check file changed
			assert.Equal(t, fileSizeAfter > fileSizeBefore, tt.outStat, "Equal file size after message writing")
			//***********************************************************************************

			//*********************************** OUTDEBUG **************************************
			// save file param
			fileSizeBefore = fileSizeAfter
			// write log
			log.OutDebug("Test message: OutDebug")
			// wait untill messages writing
			time.Sleep(tDur)
			f, err = os.Stat(logFileName)
			assert.NoError(t, err, "File control while test writing - after stage")
			if err != nil {
				t.Fatal()
			}
			fileSizeAfter = f.Size()
			// check file changed
			assert.Equal(t, fileSizeAfter > fileSizeBefore, tt.outDebugStat, "Equal file size after message writing")
			//**********************************************************************************

			//*********************************** OUTINFO **************************************
			// save file param
			fileSizeBefore = fileSizeAfter
			// write log
			log.OutInfo("Test message: OutInfo")
			// wait untill messages writing
			time.Sleep(tDur)
			f, err = os.Stat(logFileName)
			assert.NoError(t, err, "File control while test writing - after stage")
			if err != nil {
				t.Fatal()
			}
			fileSizeAfter = f.Size()
			// check file changed
			assert.Equal(t, fileSizeAfter > fileSizeBefore, tt.outInfoStat, "Equal file size after message writing")
			//**********************************************************************************

			//*********************************** OUTWARNING ***********************************
			// save file param
			fileSizeBefore = fileSizeAfter
			// write log
			log.OutWarning("Test message: OutWarning")
			// wait untill messages writing
			time.Sleep(tDur)
			f, err = os.Stat(logFileName)
			assert.NoError(t, err, "File control while test writing - after stage")
			if err != nil {
				t.Fatal()
			}
			fileSizeAfter = f.Size()
			// check file changed
			assert.Equal(t, fileSizeAfter > fileSizeBefore, tt.outWarnStat, "Equal file size after message writing")
			//**********************************************************************************

			//*********************************** OUTERROR *************************************
			// save file param
			fileSizeBefore = fileSizeAfter
			// write log
			log.OutError("Test message: OutError")
			// wait untill messages writing
			time.Sleep(tDur)
			f, err = os.Stat(logFileName)
			assert.NoError(t, err, "File control while test writing - after stage")
			if err != nil {
				t.Fatal()
			}
			fileSizeAfter = f.Size()
			// check file changed
			assert.Equal(t, fileSizeAfter > fileSizeBefore, tt.outErrorStat, "Equal file size after message writing")
			//**********************************************************************************
		})
	}
}
