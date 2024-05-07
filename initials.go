package golanglogger

// Default log level
const initLogLevel = ErrorLvl

// Default logger channel buffer size
const initLogBuffer = 20

// Default time checking file size (in seconds)
const initCheckTime = 10




// initializing base parameters (set parameters what not standart init values)
func getBaseParam() *logParam {
	return &logParam{logLvl: initLogLevel, lBuf: initLogBuffer, checkFileTime: initCheckTime}
}