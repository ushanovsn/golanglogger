# golanglogger

Logger module (project within of language learning)


<!--Info block-->

## Use logger:
Asynchronous Logger created as goroutine and will works in async mode (logger implements "Golanglogger" interface).  Messages for log are transmitted through the channel with no zero buffer size. <br/>
Synchronous Logger implements "Golanglogger" interface and work synchronous with app flow.

File day size control works in another goroutine and check the time for period 60 sec (and check is performed in base logger when recording message every new day).
It not enough accurate variant, but low count of file accesses (to prevent old logs from getting into the file, a time check is used for each writing).
A new file will be created after the specified duration at about 00:00.

1. At startup, need to create a logger:
```Go
log := golanglogger.New(golanglogger.DebugLvl, "test.log")
```
for synchronous:
```Go
log := golanglogger.NewSync(golanglogger.DebugLvl, "test.log")
```
define logging level and file name for log, if log file name is empty - no loging to file

2. At stop, need to stop a logger for writing out all buffered messages (and closing log-file):
```Go
defer log.StopLog()
```

3. After logger creation, it is possible to change the some parameters while program running:

    * Change logging level:
    ```Go
    log.SetLevel(golanglogger.ErrorLvl)
    ```

    * Change logger name:
    ```Go
    log.SetName("Iam LOGGER!")
    ```
    The name is writing as a prefix (after time and before log level) of the output log message

    * Change buffer size (default value: 20):
    ```Go
    log.SetBufferSize(100)
    ```
    available values at 1 to 1000

    * Set file control parameters (default switched off values: 0, 0):
    ```Go
    log.SetFileParam(10, 1)
    ```
     available zero or any positive values, when zero - is disabling value. <br/>
     the value of the file size in MB <br/>
     the value of the file day size in days (day fixed in 00:00 for localtime with automatic using timezone) <br/>

4. Use writing log functions, for logging with necessary levels:
```Go
log.Out("This write everithing, without level control")
log.Debug("This write only when Debug level available")
log.Info("This write only when Info level available and lower")
log.Warn("This write only when Warning level available and lower")
log.Error("This write only when Error level available and lower")
```
"Debug" is minimum level, "Error" is maximum

