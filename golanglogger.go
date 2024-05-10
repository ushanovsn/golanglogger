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
	// delay for check file size function (in seconds)
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
	// logger running flag
	fLogRun bool
}

// create logger w params (if file path is "" then no logging into file)
func New(l LoggingLevel, filePath string) *Logger {
	var log Logger
	var err error
	log.param = getBaseParam()
	
	// set received level
	log.param.logLvl = l
	// create channel for communications with logger
	log.logChan = make(chan logData, log.param.lBuf)
	// init log linking objects
	log.wg = &sync.WaitGroup{}
	log.rmu = &sync.RWMutex{}

	// set out into file if needed
	if filePath != "" {
		log.param.logFile, err = getFileOut(filePath)
		if err == nil {
			log.param.fileOutPath = filePath
		}
		// write error after logger starting
	}

	// memorized starting logger
	log.wg.Add(1)
	// start new logger
	go logger(&log)

	// indicate logger started
	log.fLogRun = true

	// the greeting line after logger was started
	log.Out("Logger starting w log level = " + l.Name())
	if err != nil {
		log.OutError("Error while init log-file: " + err.Error())
	}

	return &log
}

// restart logger
func (log *Logger) restart() {
	var err error

	// create new channel for communications with logger
	log.logChan = make(chan logData, log.param.lBuf)

	// set out into file if needed
	if log.param.fileOutPath != "" {
		log.param.logFile, err = getFileOut(log.param.fileOutPath)
		// write error after logger starting
	}

	// memorized starting logger
	log.wg.Add(1)
	// start new logger
	go logger(log)

	// indicate logger started
	log.fLogRun = true

	// the greeting line after logger was started
	log.Out("Logger starting w log level = " + log.param.logLvl.Name())
	if err != nil {
		log.OutError("Error while init log-file: " + err.Error())
	}
}


// stopping logger
func (log *Logger) StopLog() {
	log.Out("Logger stopping by command \"STOP\"")
	log.cmdOut(log.logChan, cmdStop)
	// wait when logger stopping
	log.wg.Wait()

	// now logger stopped

	if log.param.logFile != nil {
		err := log.param.logFile.Close()
		if err != nil {
			fmt.Printf("Error while log file closing:  \"%s\"\n", err)
		}
	}

	log.param.logFile = nil	
	log.fLogRun = false
}


// set log level
func (log *Logger) SetLevel(l LoggingLevel) {
	log.rmu.Lock()
	log.param.logLvl = l
	log.rmu.Unlock()
	log.cmdOut(log.logChan, cmdReloadConf)
	log.Out("Set logging level = " + l.Name())
}


// set log level
func (log *Logger) SetBufferSize(bSize int) {
	if bSize < minBufferSize || bSize > maxBufferSize {
		log.OutError(fmt.Sprintf("Wrong buffer size: %d", bSize))
		return
	}

	log.Out(fmt.Sprintf("Logger will be restarted w new buffer size: %d", bSize))

	// stopping logger
	log.StopLog()

	// set new value to bufer
	log.rmu.Lock()
	log.param.lBuf = bSize
	log.rmu.Unlock()

	log.restart()

	log.Out(fmt.Sprintf("Change log buffer size changed to %d", bSize))
}


// set log to file
func (log *Logger) SetFileParam(mbSize int, daySize int) {
	log.rmu.Lock()
	log.param.fileMbSize = mbSize
	log.param.fileDaySize = daySize
	log.rmu.Unlock()

	log.OutDebug(fmt.Sprintf("Logger will restart with new file parameters: %d Mb size, %d days duration", mbSize, daySize))
	log.cmdOut(log.logChan, cmdReloadConf)
}


// set log out writers parameters (con - console out flag, stdErr - standart error out flag)
func (log *Logger) SetErrOut(con bool, stdErr bool) bool {

	if !con && !stdErr {
		log.OutError("Error parameter for out log. Received STDOUT and STDERR false for both. Skip this.")
		return false
	}

	log.rmu.Lock()
	log.param.fNoCon = !con
	log.param.fStdErr = stdErr
	log.rmu.Unlock()

	log.OutDebug("Logger will restart with new outs parameters")
	log.cmdOut(log.logChan, cmdReloadConf)

	return true
}

// write logging mesage
func writeLog(c chan logData, t time.Time, s string) {
	c <- logData{t.Format("2006-01-02 15:04:05.000"), s, cmdWrite}
}

// write logging command
func writeCmd(c chan<- logData, cmd loggingCmd) {
	c <- logData{"", "", cmd}
}

// out Debug string into log
func (log *Logger) OutDebug(msg string) {
	// don't use mutex, the level will  changing rare and it not important if reading old or new value of it
	if int(log.param.logLvl) <= int(DebugLvl) && log.fLogRun {
		writeLog(log.logChan, time.Now(), "[DBG]: "+msg)
	}
}

// out Info string into log
func (log *Logger) OutInfo(msg string) {
	// don't use mutex, the level will  changing rare and it not important if reading old or new value of it
	if int(log.param.logLvl) <= int(InfoLvl) && log.fLogRun {
		writeLog(log.logChan, time.Now(), "[INF]: "+msg)
	}
}

// out Warning string into log
func (log *Logger) OutWarning(msg string) {
	// don't use mutex, the level will  changing rare and it not important if reading old or new value of it
	if int(log.param.logLvl) <= int(WarningLvl) && log.fLogRun {
		writeLog(log.logChan, time.Now(), "[WRN]: "+msg)
	}
}

// out Error string into log
func (log *Logger) OutError(msg string) {
	// don't use mutex, the level will  changing rare and it not important if reading old or new value of it
	if int(log.param.logLvl) <= int(ErrorLvl) && log.fLogRun {
		writeLog(log.logChan, time.Now(), "[ERR]: "+msg)
	}
}

// always out string into log
func (log *Logger) Out(msg string) {
	if log.fLogRun {
		writeLog(log.logChan, time.Now(), "[MSG]: "+msg)
	}
	
}

// send command to logger
func (log *Logger) cmdOut(c chan<- logData, cmd loggingCmd) {
	if log.fLogRun {
		log.OutDebug(fmt.Sprintf("Internal command for logger: %d", cmd))
		writeCmd(c, cmd)
	}
	
}




// logger goroutine
func logger(l *Logger) {
	// reset memorysed logger when exit
	defer l.wg.Done()


	// base log writer
	var writer io.Writer
	// parameters copy values
	var param logParam
	// channel for stop goroutine check time file 
	var cCancel chan struct{}
	// current received command
	var cmd loggingCmd = cmdIdle
	// message what was generating error in cycle before
	var lostLogMsg string
	// error while cmd processing
	var procErrorMsg string



	// Now base cycle of logger func
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
			// if one gorotine started before - it was stopped and channel closed when cmd received in main cycle
			cCancel = make(chan struct{})
			go fileTimeControl(l, cCancel)
		}

		// check lost messages
		if lostLogMsg != "" {
			_, err := io.WriteString(writer, lostLogMsg+"\n")
			if err != nil {
				// when error - just do it again and again
				fmt.Printf("Multiple ERROR while write log string \"%s\"; Error: %s\n", lostLogMsg, err.Error())
				time.Sleep(time.Second)
				continue
			}
			lostLogMsg = ""
		}
		// check old errors
		if procErrorMsg != "" {
			_, err := io.WriteString(writer, procErrorMsg+"\n")
			if err != nil {
				// when error - just do it again and again
				fmt.Printf("Multiple ERROR while write old error into log. Old error: \"%s\"; Current error: %s\n", procErrorMsg, err.Error())
				time.Sleep(time.Second)
				continue
			}
			procErrorMsg = ""
		}


		// main cycle for getting data from channel cycle
		for msg := range l.logChan {
			// received message:
			if msg.cmd == cmdWrite {
				// write log
				_, err := io.WriteString(writer, msg.t+" -> "+msg.msg+"\n")
				if err != nil {
					// need restart writers and reload config
					// memorize message
					lostLogMsg = msg.t+" -> "+msg.msg
					// memorize error
					procErrorMsg = fmt.Sprintf("Error while write log string \"%s\"; Error: %s", lostLogMsg, err.Error())
					cmd = cmdReloadConf
					break
				}

				// if check file MB size control enabled
				if param.fileMbSize > 0 && param.logFile != nil {
					f, err := getFileMbSize(param.fileOutPath)
					if err != nil {
						// need restart writers and reload config
						procErrorMsg = fmt.Sprintf("Error while getting File Mb Size; File: \"%s\"; Error: %s", param.fileOutPath, err.Error())
						cmd = cmdReloadConf
						break
					}

					// when file is too big - generate cmd for change it
					if f >= param.fileMbSize {
						cmd = cmdChangeFile
						break
					}
				}
			} else if msg.cmd == cmdIdle {
				// if nothing to do - continue listen the channel
				continue
			} else {
				// need processing the received cmd - break the cycle
				cmd = msg.cmd
				break
			}
		}



		// stop file control (this place is one who closing channel, and checking for nil - is enought)
		if cCancel != nil {
			close(cCancel)
		}

		// processing commands
		switch cmd {
		case cmdChangeFile:
			// need changing the log file
			var err error
			param.logFile, err = changeFile(param.logFile, param.fileOutPath, true)
			if err != nil {
				procErrorMsg = fmt.Sprintf("Error while changing log file at time (%v). Err: %s", time.Now(), err.Error())
			} else {
				// set file to main parameters of base logger
				l.rmu.Lock()
				l.param.logFile = param.logFile
				l.rmu.Unlock()
			}

		case cmdReloadConf:
			// restart base cycle and copy param at start

		case cmdStop:
			// indcate logger stopped
			l.rmu.Lock()
			l.fLogRun = false
			l.rmu.Unlock()
			// and channel closing
			close(l.logChan)

			// read all lost messages and stop
			for msg := range l.logChan {
				// received message:
				if msg.cmd == cmdWrite {
					// write log as can
					_, _ = io.WriteString(writer, msg.t+" -> "+msg.msg+"\n")
				}
			}
			// just line for visual control
			_, _ = io.WriteString(writer,"********************************************************************************************\n\n\n")
			return
		}

		// check errors msg and emergency write it into console
		if procErrorMsg != "" {
			fmt.Printf("No logged Error:  \"%s\"\n", procErrorMsg)
		}
	}
}



// control of log file time duration 
func fileTimeControl(l *Logger, cCancel chan struct{}) {

	const secondInDay = 60 * 60 * 24

	// timezone offset
	_, zOffset := time.Now().In(time.Local).Zone()

	// set parameters at starting monitoring
	l.rmu.RLock()
	// file
	filePath := l.param.fileOutPath
	// control time
	daySize := l.param.fileDaySize * secondInDay
	// period of control operations
	checkTime := l.param.checkFileTime
	l.rmu.RUnlock()

	l.OutDebug(fmt.Sprintf("Time control starting for control log file duration: %d days", (daySize/secondInDay)))


	// base cycle for control duration
	for {
		// time file creation in second
		tSec, err := getDateTimeFile(filePath)

		if err != nil {
			l.OutError("Log time control will stopped. Error while getting datetime creating of file. Err: " + err.Error())
			return
		}

		// time in seconds at day starting for file creating time + size log in seconds
		fileChangeDaySec := ((tSec + int64(zOffset)) / secondInDay)*secondInDay + int64(daySize)

		// if time is already gone...
		if (time.Now().Unix() + int64(zOffset)) >= fileChangeDaySec {
			l.rmu.RLock()
			// read logger status
			fRun := l.fLogRun
			l.rmu.RUnlock()

			if fRun {
				l.cmdOut(l.logChan, cmdChangeFile)
			}
			
			return
		}


		// wait timers or signal to stop
		select {
		case <-cCancel:
			l.OutDebug("Time control stopping by signal \"STOP\"")
			return
		case <-time.After(time.Duration(checkTime) * time.Second):
			// period timer occured
		}
	}
}