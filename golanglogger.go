package golanglogger

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// base logger interface
type Golanglogger interface {
	// out Debug string into log
	OutDebug(msg string)
	// out Info string into log
	OutInfo(msg string)
	// out Warning string into log
	OutWarning(msg string)
	// out Error string into log
	OutError(msg string)
	// always out string into log
	Out(msg string)
	// set log level
	SetLevel(LoggingLevel)
	// set name of logger for write as prefix to log
	SetName(nameStr string)
	// set log file control parameters
	SetFileParam(mbSize int, daySize int)
	// set log channel buffer size
	SetBufferSize(int)
	// set log out writers parameters (con - console out flag, stdErr - standart error out flag)
	SetStdOut(con bool, stdErr bool)
	// get logger parameters for writing messages
	CurrentOutParams() (con bool, stdErr bool, filePath string)
	// get logger buffer size
	CurrentBufSize() (size int)
	// get logger level
	CurrentLevel() (lvl LoggingLevel)
	// get logger name
	CurrentName() (nameStr string)
	// get logger file control parameters
	CurrentFileControl() (mbSize int, daySize int)
	// stopping logger
	StopLog()
}

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
	// name of logger
	name string
	// constructed name prefix
	namePref string
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
func New(l LoggingLevel, filePath string) Golanglogger {
	var log Logger
	log.param = getBaseParam()

	// set received level
	log.param.logLvl = l
	// set received path
	log.param.fileOutPath = filePath
	// init log linking objects
	log.wg = &sync.WaitGroup{}
	log.rmu = &sync.RWMutex{}

	log.start()

	return &log
}

// start an logger
func (log *Logger) start() {
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
}

// set log level
func (log *Logger) SetLevel(l LoggingLevel) {
	log.rmu.Lock()
	log.param.logLvl = l
	log.rmu.Unlock()
	log.cmdOut(log.logChan, cmdReloadConf)
	log.Out("Set logging level = " + l.Name())
}

// set log channel buffer size
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

	// starting logger
	log.start()

	log.Out(fmt.Sprintf("Logger buffer size changed to %d", bSize))
}

// set log file control parameters
func (log *Logger) SetFileParam(mbSize int, daySize int) {
	if mbSize < 0 {
		log.OutError(fmt.Sprintf("Wrong file size control value: %d", mbSize))
		return
	}
	if daySize < 0 {
		log.OutError(fmt.Sprintf("Wrong day size for file control value: %d", daySize))
		return
	}

	log.rmu.Lock()
	log.param.fileMbSize = mbSize
	log.param.fileDaySize = daySize
	log.rmu.Unlock()

	log.OutDebug(fmt.Sprintf("Logger will restart with new file parameters: %d Mb size, %d days duration", mbSize, daySize))
	log.cmdOut(log.logChan, cmdReloadConf)
}

// set log out writers parameters (con - console out flag, stdErr - standart error out flag)
func (log *Logger) SetStdOut(con bool, stdErr bool) {

	if !con && !stdErr {
		log.OutError("Error parameter for out log. Received STDOUT and STDERR false for both. Skip this.")
		return
	}

	log.rmu.Lock()
	log.param.fNoCon = !con
	log.param.fStdErr = stdErr
	log.rmu.Unlock()

	log.OutDebug("Logger will restart with new outs parameters")
	log.cmdOut(log.logChan, cmdReloadConf)
}

// get logger parameters for writing messages
func (log *Logger) CurrentOutParams() (con bool, stdErr bool, filePath string) {

	log.rmu.RLock()
	con = !log.param.fNoCon
	stdErr = log.param.fStdErr
	filePath = log.param.fileOutPath
	log.rmu.RUnlock()

	return con, stdErr, filePath
}

// get logger buffer size
func (log *Logger) CurrentBufSize() (size int) {

	log.rmu.RLock()
	size = log.param.lBuf
	log.rmu.RUnlock()

	return size
}

// get logger level
func (log *Logger) CurrentLevel() (lvl LoggingLevel) {

	log.rmu.RLock()
	lvl = log.param.logLvl
	log.rmu.RUnlock()

	return lvl
}

// get logger file control parameters
func (log *Logger) CurrentFileControl() (mbSize int, daySize int) {

	log.rmu.RLock()
	mbSize = log.param.fileMbSize
	daySize = log.param.fileDaySize
	log.rmu.RUnlock()

	return mbSize, daySize
}

// write logging mesage
func writeLog(w *io.Writer, n string, d logData) error {
	_, err := io.WriteString(*w, d.t + " -> " + n + d.msg + "\n")

	if err != nil {
		return err
	}
	return nil
}

// send logging mesage
func sendToLog(c chan logData, t time.Time, s string) {
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
		sendToLog(log.logChan, time.Now(), "[DBG]: "+msg)
	}
}

// out Info string into log
func (log *Logger) OutInfo(msg string) {
	// don't use mutex, the level will  changing rare and it not important if reading old or new value of it
	if int(log.param.logLvl) <= int(InfoLvl) && log.fLogRun {
		sendToLog(log.logChan, time.Now(), "[INF]: "+msg)
	}
}

// out Warning string into log
func (log *Logger) OutWarning(msg string) {
	// don't use mutex, the level will  changing rare and it not important if reading old or new value of it
	if int(log.param.logLvl) <= int(WarningLvl) && log.fLogRun {
		sendToLog(log.logChan, time.Now(), "[WRN]: "+msg)
	}
}

// out Error string into log
func (log *Logger) OutError(msg string) {
	// don't use mutex, the level will  changing rare and it not important if reading old or new value of it
	if int(log.param.logLvl) <= int(ErrorLvl) && log.fLogRun {
		sendToLog(log.logChan, time.Now(), "[ERR]: "+msg)
	}
}

// always out string into log
func (log *Logger) Out(msg string) {
	if log.fLogRun {
		sendToLog(log.logChan, time.Now(), "[MSG]: "+msg)
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
	// previous day value for control file time duration
	var prepDay int
	// time in seconds for control file duration
	var daySize int

	// Now base cycle of logger func
	for {
		// parameters copy
		l.rmu.RLock()
		param = *l.param
		l.rmu.RUnlock()

		// current time zone offset
		_, zOffset := time.Now().In(time.Local).Zone()

		// reinit writers
		writer = getWriter(param.fNoCon, param.fStdErr, param.logFile)

		// check file day size control
		if param.fileDaySize > 0 && param.logFile != nil {
			// if one gorotine started before - it was stopped and channel closed when cmd received in main cycle
			cCancel = make(chan struct{})
			go fileTimeControl(l, zOffset, cCancel)

			// prepare day value for further use
			prepDay = time.Now().YearDay()
			daySize = param.fileDaySize * secondInDay
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

			// received message for logging:
			if msg.cmd == cmdWrite {
				// before check time control of file, not to write new values to old file
				if param.fileDaySize > 0 && param.logFile != nil {
					// if in this iteration - new day starting
					if time.Now().YearDay()-prepDay != 0 {
						// check file duration
						fChange, err := checkFileTime(zOffset, param.fileOutPath, daySize)

						if err != nil {
							// need restart writers and reload config
							// memorize message
							lostLogMsg = msg.t + " -> " + msg.msg
							// memorize error
							procErrorMsg = fmt.Sprintf("Error while check log file time duration. Error: %s", err.Error())
							cmd = cmdReloadConf
							break
						}

						// if time is already gone...
						if fChange {
							lostLogMsg = msg.t + " -> " + msg.msg
							cmd = cmdChangeFile
							break
						}
					}
				}

				// write log
				_, err := io.WriteString(writer, msg.t+" -> "+msg.msg+"\n")
				if err != nil {
					// need restart writers and reload config
					// memorize message
					lostLogMsg = msg.t + " -> " + msg.msg
					// memorize error
					procErrorMsg = fmt.Sprintf("Error while write log string \"%s\"; Error: %s", lostLogMsg, err.Error())
					cmd = cmdReloadConf
					break
				}

				// check file MB size control enabled after writing new message
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
			_, _ = io.WriteString(writer, "********************************************************************************************\n\n\n")
			return
		}

		// check errors msg and emergency write it into console
		if procErrorMsg != "" {
			fmt.Printf("No logged Error:  \"%s\"\n", procErrorMsg)
		}
	}
}

// independed control of log file time duration
func fileTimeControl(l *Logger, zOffset int, cCancel chan struct{}) {

	// set parameters at starting monitoring
	l.rmu.RLock()
	// file
	filePath := l.param.fileOutPath
	// control time
	daySize := l.param.fileDaySize * secondInDay
	// period of control operations
	checkTime := l.param.checkFileTime
	l.rmu.RUnlock()

	l.OutDebug(fmt.Sprintf("Time control starting for control log file duration: %d days", (daySize / secondInDay)))

	// base cycle for control duration
	for {

		// check file duration
		fChange, err := checkFileTime(zOffset, filePath, daySize)

		if err != nil {
			l.OutError("Log time control will stopped. Error while getting datetime creating of file. Err: " + err.Error())
			return
		}

		// if time is already gone...
		if fChange {
			l.rmu.RLock()
			// read logger status
			fRun := l.fLogRun
			l.rmu.RUnlock()

			// check cancel and send cmd for change file
			if fRun {
				select {
				case <-cCancel:
					l.OutDebug("Time control stopping by signal \"STOP\"")
					return
				default:
					l.cmdOut(l.logChan, cmdChangeFile)
				}
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

// check expiration control of log file time duration
//
// zOffset - TimeZone shift in hours;
// filePath - controlling file path;
// daySizeSec - maximum file durations in day*seconds;
func checkFileTime(zOffset int, filePath string, daySizeSec int) (bool, error) {

	// time file creation in second
	tSec, err := getDateTimeFile(filePath)

	if err != nil {
		return false, err
	}

	// maximum time in seconds up to which a file can exist
	fileChangeDaySec := ((tSec+int64(zOffset))/secondInDay)*secondInDay + int64(daySizeSec)

	// if time is already gone...
	if (time.Now().Unix() + int64(zOffset)) >= fileChangeDaySec {
		return true, nil
	} else {
		return false, nil
	}
}

// aggregating writers for logger
//
// fNoCon - flag for don't use console; fStdErr - flag for use stdError out; logFile - pointer to logfile
func getWriter(fNoCon bool, fStdErr bool, logFile *os.File) io.Writer {
	// reset writers
	var writer io.Writer

	// chk disabling console
	if !fNoCon {
		writer = io.Writer(os.Stdout)
	}

	// chk stdErr out (add it or init with it)
	if fStdErr {
		if writer == nil {
			writer = io.Writer(os.Stderr)
		} else {
			writer = io.MultiWriter(writer, os.Stderr)
		}
	}

	// add file to out
	if logFile != nil {
		if writer == nil {
			writer = io.Writer(logFile)
		} else {
			writer = io.MultiWriter(writer, logFile)
		}
	}

	return writer
}
