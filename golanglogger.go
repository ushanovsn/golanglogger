package golanglogger

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// logger cmd enums
type loggingCmd int

const (
	cmdIdle loggingCmd = iota
	cmdWrite
	cmdStop
	cmdReloadConf
	cmdChangeFile
)

// struct for logger interaction
type logData struct {
	t   string
	msg string
	cmd loggingCmd
}

// struct for logger parameters
type logParam struct {
	// current log level
	logLvl LoggingLevel
	// disable console out log flag
	fNoCon bool
	// stderr out log flag
	fStdErr bool
	// logger buffer
	lBuf int
	// logger file
	logFile *os.File
	// file name for logging
	fileOutPath string
	// file size
	fileMbSize int
	// file duration
	fileDaySize int
	// check file delay size period in seconds
	checkFileTime int
}

// logger object type
type Logger struct {
	// logger mutex for param interactions
	rmu *sync.RWMutex
	// logger parameters
	param *logParam
	// logger channel
	logChan chan logData
	// logger counter
	wg *sync.WaitGroup
}

// create logger w params
func New(l LoggingLevel) *Logger {
	var log Logger
	log.param = getBaseParam()
	

	// set received parameters
	log.param.logLvl = l
	log.logChan = make(chan logData, log.param.lBuf)

	// memorized starting logger
	log.wg.Add(1)
	// start new logger
	go logger(&log)
	log.Out("Logger starting w log level = " + l.Name())

	return &log
}

// initializing base parameters (set parameters what not standart init values)
func getBaseParam() *logParam {
	return &logParam{logLvl: initLogLevel, lBuf: initLogBuffer, checkFileTime: initCheckTime}
}

// stopping logger
func (log *Logger) StopLog() {
	writeCmd(log.logChan, cmdStop)
	log.wg.Wait()
}


// set log level
func (log *Logger) SetLevel(l LoggingLevel) {
	log.rmu.Lock()
	log.param.logLvl = l
	log.rmu.Unlock()
	writeCmd(log.logChan, cmdReloadConf)
	log.Out("Set logging level = " + l.Name())
}

// set log to file
func (log *Logger) SetFile(fPath string, mbSize int, daySize int) bool {
	logFile, err := os.OpenFile(fPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.OutError(`Error parameter for log file. Wrong parameter: "` + fPath + `". Error: ` + err.Error())
		return false
	}

	log.rmu.Lock()
	log.param.logFile = logFile
	log.param.fileOutPath = fPath
	log.param.fileMbSize = mbSize
	log.param.fileDaySize = daySize
	log.rmu.Unlock()

	log.OutDebug("Logger will restart with new file parameters")
	writeCmd(log.logChan, cmdReloadConf)

	return true
}

// set log out writers parameters
func (log *Logger) SetErrOut(con bool, stdErr bool) bool {

	if !con && !stdErr {
		log.OutError("Error parameter for out log. Received STDOUT and STDERR false for both")
		return false
	}

	log.rmu.Lock()
	log.param.fNoCon = !con
	log.param.fStdErr = stdErr
	log.rmu.Unlock()

	log.OutDebug("Logger will restart with new outs parameters")
	writeCmd(log.logChan, cmdReloadConf)

	return true
}

// write logging mesage
func writeLog(c chan logData, t time.Time, s string) {
	c <- logData{t.String(), s, cmdWrite}
}

// write logging command
func writeCmd(c chan<- logData, cmd loggingCmd) {
	c <- logData{"", "", cmd}
}

// out Debug string into log
func (log *Logger) OutDebug(msg string) {
	// don't use mutex, the level will  changing rare and it not important if reading old or new value of it
	if int(log.param.logLvl) <= int(DebugLvl) {
		writeLog(log.logChan, time.Now(), "[DBG]: "+msg)
	}
}

// out Info string into log
func (log *Logger) OutInfo(msg string) {
	// don't use mutex, the level will  changing rare and it not important if reading old or new value of it
	if int(log.param.logLvl) <= int(InfoLvl) {
		writeLog(log.logChan, time.Now(), "[INF]: "+msg)
	}
}

// out Warning string into log
func (log *Logger) OutWarning(msg string) {
	// don't use mutex, the level will  changing rare and it not important if reading old or new value of it
	if int(log.param.logLvl) <= int(WarningLvl) {
		writeLog(log.logChan, time.Now(), "[WRN]: "+msg)
	}
}

// out Error string into log
func (log *Logger) OutError(msg string) {
	// don't use mutex, the level will  changing rare and it not important if reading old or new value of it
	if int(log.param.logLvl) <= int(ErrorLvl) {
		writeLog(log.logChan, time.Now(), "[ERR]: "+msg)
	}
}

// always out string into log
func (log *Logger) Out(msg string) {
	writeLog(log.logChan, time.Now(), "[MSG]: "+msg)
}

// logger goroutine
func logger(l *Logger) {
	// reset memorysed logger when exit
	defer l.wg.Done()

	// log writer
	var writer io.Writer

	// parameters copy values
	var param logParam

	// stop goroutine check filec hannel
	var cCancel chan struct{}
	// current received command
	var cmd loggingCmd = cmdIdle
	// message what was generating error in cycle before
	var lostLogMsg string
	// error while changing file
	var procErrorMsg string


	// Now base command cycle of logger func
	for {
		// parameters copy
		l.rmu.RLock()
		param = *l.param
		l.rmu.RUnlock()
		
		// reset writers
		writer = nil

		// chk disabling console
		if !param.fNoCon {
			writer = io.Writer(os.Stdout)
		}
		// chk stdErr out (add it or init with it)
		if param.fStdErr {
			if (writer == nil) {
				writer = io.Writer(os.Stderr)
			} else {
				writer = io.MultiWriter(writer, os.Stderr)
			}
		}

		// add file to out
		if param.logFile != nil {
			if (writer == nil) {
				writer = io.Writer(param.logFile)
			} else {
				writer = io.MultiWriter(writer, param.logFile)
			}
		}

		// check file day size control
		if param.fileDaySize > 0 && param.logFile != nil {
			// close old channel
			if _, ok := <- cCancel; ok {
				close(cCancel)
			}
			cCancel = make(chan struct{})
			go fileTimeControl(l, cCancel)
			defer close(cCancel)
		}



		// main cycle for getting data from channel cycle
		for msg := range l.logChan {
			switch msg.cmd {
			case cmdWrite:
				// write log
				_, err := io.WriteString(writer, msg.t+" -> "+msg.msg+"\n")
				if err != nil {
					lostLogMsg = msg.t+" -> "+msg.msg
					// need restart writers and reload config
					procErrorMsg = fmt.Sprintf("Error while write log string \"%s\"; Error: %s", lostLogMsg, err.Error())
					cmd = cmdReloadConf
					break
				}

				// check file MB size control
				if param.fileMbSize > 0 && param.logFile != nil {
					f, err := getFileMbSize(param.fileOutPath)
					if err != nil {
						procErrorMsg = fmt.Sprintf("Error while getting File Mb Size; File: \"%s\"; Error: %s", param.fileOutPath, err.Error())
						// need restart writers and reload config
						cmd = cmdReloadConf
						break
					}

					if f >= param.fileMbSize {
						cmd = cmdChangeFile
						break
					}
				}
			case cmdIdle:
				// nothing do
			default:
				cmd = msg.cmd
				break
			}
		}

		// stop file control
		if _, ok := <- cCancel; ok {
			close(cCancel)
		}

		// processing commands
		switch cmd {
		case cmdChangeFile:
			var err error
			param.logFile, err = changeFile(param.logFile, param.fileOutPath)
			if err != nil {
				procErrorMsg = fmt.Sprintf("Error while changing log file at time (%v). Err: %s", time.Now(), err.Error())
			}
		case cmdReloadConf:
			// restart cycle and copy param at start
		case cmdStop:
			return
		}

		// check errors msg
		if procErrorMsg != "" {
			fmt.Printf("No logged Error:  \"%s\"\n", procErrorMsg)
		}
	}
}

// returning file size and errors
func getFileMbSize(path string) (int, error) {
	f, err := os.Stat(path)
	if err == nil {
		return int(f.Size() / 1048576), nil
	} else {
		return 0, err
	}
}

// log file rotation (write into file must be paused while process it)
func changeFile(logFile *os.File, fileName string) (*os.File, error) {

	// close old file
	err := logFile.Close()
	if err != nil {
		return nil, err
	}

	// rename old file
	t := time.Now()
	formattedT := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
	err = os.Rename(fileName, fileName+"_"+formattedT)
	if err != nil {
		return nil, err
	}

	// new file
	logFile, err = os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return nil, err
	}

	return logFile, nil
}


// control of log file time duration 
func fileTimeControl(l *Logger, cCancel chan struct{}) {

	const secondInDay = 60 * 60 * 24

	l.rmu.RLock()
	filePath := l.param.fileOutPath
	daySize := l.param.fileDaySize * secondInDay
	checkTime := l.param.checkFileTime
	l.rmu.RUnlock()


	// base cycle for control duration
	for {
		// time file creation in second
		tSec, err := getDateTimeFile(filePath)

		if err != nil {
			l.OutError("Log time control will stopped. Error while getting datetime creating of file. Err: " + err.Error())
			return
		}

		// time in seconds at day starting for file creating time + size log in seconds
		fileChangeDaySec := (tSec / secondInDay) + int64(daySize)

		// if time is already gone...
		if time.Now().Unix() >= fileChangeDaySec {
			writeCmd(l.logChan, cmdChangeFile)
			return
		}


		// wait timers or signal to stop
		select {
		case <-cCancel:
			l.OutInfo("Time control stopping by signal \"STOP\"")
			return
		case <-time.After(time.Duration(checkTime) * time.Second):
			// period timer occured
		}
	}
}