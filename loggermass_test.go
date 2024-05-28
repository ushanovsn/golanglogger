package golanglogger

import (
	"os"
	"testing"
	"sync"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func Test_OneLoggerOneFile(t *testing.T) {
	// values for testing

	// count of goroutines
	cnt := 10
	// count of writing strings for each goroutine
	strCnt := 10
	// waiting goroutines
	var grp  sync.WaitGroup
	// loger level
	logLevel := DebugLvl
	// logger type
	var lType string
	// time for waiting writing async logger
	tDur := time.Millisecond * 300
	// finalize text size
	var fSize  int64
	// string-text in log  size
	var strSize  int


	for _, v := range ([]int{1,2}){

		// temp file
		logFileName := uuid.NewString() + ".log"
		defer os.Remove(logFileName)

		var log Golanglogger

		if v == 1 {
			// creating logger
			log = New(logLevel, logFileName)
			lType = "async"
			// finalize text size for async logger
			fSize = 228
		} else {
			// creating logger
			log = NewSync(logLevel, logFileName)
			lType = "sync"
			// finalize text size for sync logger
			fSize = 170
		}


		// check logger created
		if v == 1 {
			// creating logger
			assert.NotNil(t, log, "Check logger async exists")
			if log == nil {
				t.Fatal()
			}
		} else {
			// creating logger
			assert.NotNil(t, log, "Check logger syncronious exists")
			if log == nil {
				t.Fatal()
			}
		}

		t.Run(fmt.Sprintf("Check logger %s creating", lType), func(t *testing.T) {
			logCon, logErr, logFile := log.CurrentOutParams()
			lvl := log.CurrentLevel()
			assert.True(t, logCon, "Console out enabled")
			assert.False(t, logErr, "stdError out enabled")
			assert.Equal(t, logFileName, logFile, "Logger file out")
			assert.Equal(t, logLevel, lvl, "Logger level")
			assert.FileExists(t, logFileName, "Logger file")
		})

		time.Sleep(tDur)

		// check file
		f, err := os.Stat(logFileName)
		assert.NoError(t, err, "Fix starting file size")
		if err != nil {
			t.Fatal()
			log.StopLog()
		}

		fileSizeBefore := f.Size()

		// starting tests
		for i:= 0; i<cnt; i++ {
			grp.Add(1)
			go func(k int){
				for j:=k*strCnt;j<(k+1)*strCnt;j++ {
					log.OutDebug(fmt.Sprintf("Logger: %3d. Value %9d", k, j))
					// string-text in log size for current message
					strSize = 63
				}
				grp.Done()
			}(i)
		}
		grp.Wait()
		log.StopLog()

		f, err = os.Stat(logFileName)
		assert.NoError(t, err, "Fix starting file size")
		if err != nil {
			t.Fatal()
			log.StopLog()
		}
		fileSizeAfter := f.Size()

		assert.Equal(t, (fileSizeBefore + int64(cnt*strCnt*strSize) + fSize), fileSizeAfter, fmt.Sprintf("Logger %s file size incorrect", lType))

	}
}
