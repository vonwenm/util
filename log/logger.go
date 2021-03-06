/*
Copyright (c) 2014 Ashley Jeffs

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, sub to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package log

import (
	"fmt"
	"io"
	"strings"
	"time"
)

/*--------------------------------------------------------------------------------------------------
 */

/*
Logger level constants
*/
const (
	LogOff   int = 0
	LogFatal int = 1
	LogError int = 2
	LogWarn  int = 3
	LogInfo  int = 4
	LogDebug int = 5
	LogTrace int = 6
	LogAll   int = 7
)

/*
intToLogLevel - Converts an integer into a human readable log level.
*/
func intToLogLevel(i int) string {
	switch i {
	case LogOff:
		return "OFF"
	case LogFatal:
		return "FATAL"
	case LogError:
		return "ERROR"
	case LogWarn:
		return "WARN"
	case LogInfo:
		return "INFO"
	case LogDebug:
		return "DEBUG"
	case LogTrace:
		return "TRACE"
	case LogAll:
		return "ALL"
	}
	return "ALL"
}

/*
logLevelToInt - Converts a human readable log level into an integer value.
*/
func logLevelToInt(level string) int {
	levelUpper := strings.ToUpper(level)
	switch levelUpper {
	case "OFF":
		return LogOff
	case "FATAL":
		return LogFatal
	case "ERROR":
		return LogError
	case "WARN":
		return LogWarn
	case "INFO":
		return LogInfo
	case "DEBUG":
		return LogDebug
	case "TRACE":
		return LogTrace
	case "ALL":
		return LogAll
	}
	return -1
}

/*--------------------------------------------------------------------------------------------------
 */

/*
LoggerConfig - Holds configuration options for a logger object.
*/
type LoggerConfig struct {
	Prefix       string `json:"prefix" yaml:"prefix"`
	LogLevel     string `json:"log_level" yaml:"log_level"`
	AddTimeStamp bool   `json:"add_timestamp" yaml:"add_timestamp"`
}

/*
DefaultLoggerConfig - Returns a fully defined logger configuration with the default values for each
field.
*/
func DefaultLoggerConfig() LoggerConfig {
	return LoggerConfig{
		Prefix:       "service",
		LogLevel:     "INFO",
		AddTimeStamp: true,
	}
}

/*--------------------------------------------------------------------------------------------------
 */

/*
Logger - A logger object with support for levelled logging and modular components.
*/
type Logger struct {
	stream        io.Writer
	config        LoggerConfig
	level         int
	riemannClient *RiemannClient
}

/*
NewLogger - Create and return a new logger object.
*/
func NewLogger(stream io.Writer, config LoggerConfig) *Logger {
	logger := Logger{
		stream: stream,
		config: config,
		level:  logLevelToInt(config.LogLevel),
	}
	return &logger
}

/*
NewModule - Creates a new logger object from the previous, using the same configuration, but adds
an extra prefix to represent a submodule.
*/
func (l *Logger) NewModule(prefix string) *Logger {
	config := l.config
	config.Prefix = fmt.Sprintf("%v%v", config.Prefix, prefix)

	return &Logger{
		stream:        l.stream,
		config:        config,
		level:         l.level,
		riemannClient: l.riemannClient,
	}
}

/*
UseRiemann - Register a RiemannClient object to be used for pushing log events to a riemann service.
*/
func (l *Logger) UseRiemann(client *RiemannClient) error {
	if client == nil {
		return ErrClientNil
	}
	l.riemannClient = client
	return nil
}

/*--------------------------------------------------------------------------------------------------
 */

/*
printf - Prints a log message with any configured extras prepended.
*/
func (l *Logger) printf(message, level string, other ...interface{}) {
	timestampStr := ""
	if l.config.AddTimeStamp {
		timestampStr = fmt.Sprintf("%v | ", time.Now().Format(time.RFC3339))
	}

	fmt.Fprintf(l.stream, fmt.Sprintf("%v%v | %v | %v", timestampStr, level, l.config.Prefix, message), other...)
}

/*
printLine - Prints a log message with any configured extras prepended.
*/
func (l *Logger) printLine(message, level string) {
	timestampStr := ""
	if l.config.AddTimeStamp {
		timestampStr = fmt.Sprintf("%v | ", time.Now().Format(time.RFC3339))
	}

	fmt.Fprintf(l.stream, fmt.Sprintf("%v%v | %v | %v\n", timestampStr, level, l.config.Prefix, message))
}

/*
sendRiemann - If a Riemann client has been set then we send a log event through it.
*/
func (l *Logger) sendRiemann(message, level string, other ...interface{}) {
	if l.riemannClient != nil {
		l.riemannClient.SendEvent(RiemannEvent{
			Service:     l.config.Prefix,
			State:       level,
			Description: fmt.Sprintf(message, other...),
		})
	}
}

/*--------------------------------------------------------------------------------------------------
 */

/*
Fatalf - Print a fatal message to the console. Does NOT cause panic.
*/
func (l *Logger) Fatalf(message string, other ...interface{}) {
	if LogFatal <= l.level {
		l.printf(message, "FATAL", other...)
	}
	l.sendRiemann(message, "FATAL", other...)
}

/*
Errorf - Print an error message to the console.
*/
func (l *Logger) Errorf(message string, other ...interface{}) {
	if LogError <= l.level {
		l.printf(message, "ERROR", other...)
	}
	l.sendRiemann(message, "ERROR", other...)
}

/*
Warnf - Print a warning message to the console.
*/
func (l *Logger) Warnf(message string, other ...interface{}) {
	if LogWarn <= l.level {
		l.printf(message, "WARN", other...)
	}
	l.sendRiemann(message, "WARN", other...)
}

/*
Infof - Print an information message to the console.
*/
func (l *Logger) Infof(message string, other ...interface{}) {
	if LogInfo <= l.level {
		l.printf(message, "INFO", other...)
	}
	l.sendRiemann(message, "INFO", other...)
}

/*
Debugf - Print a debug message to the console.
*/
func (l *Logger) Debugf(message string, other ...interface{}) {
	if LogDebug <= l.level {
		l.printf(message, "DEBUG", other...)
	}
}

/*
Tracef - Print a trace message to the console.
*/
func (l *Logger) Tracef(message string, other ...interface{}) {
	if LogTrace <= l.level {
		l.printf(message, "TRACE", other...)
	}
}

/*--------------------------------------------------------------------------------------------------
 */

/*
Fatalln - Print a fatal message to the console. Does NOT cause panic.
*/
func (l *Logger) Fatalln(message string) {
	if LogFatal <= l.level {
		l.printLine(message, "FATAL")
	}
	l.sendRiemann(message, "FATAL")
}

/*
Errorln - Print an error message to the console.
*/
func (l *Logger) Errorln(message string) {
	if LogError <= l.level {
		l.printLine(message, "ERROR")
	}
	l.sendRiemann(message, "ERROR")
}

/*
Warnln - Print a warning message to the console.
*/
func (l *Logger) Warnln(message string) {
	if LogWarn <= l.level {
		l.printLine(message, "WARN")
	}
	l.sendRiemann(message, "WARN")
}

/*
Infoln - Print an information message to the console.
*/
func (l *Logger) Infoln(message string) {
	if LogInfo <= l.level {
		l.printLine(message, "INFO")
	}
	l.sendRiemann(message, "INFO")
}

/*
Debugln - Print a debug message to the console.
*/
func (l *Logger) Debugln(message string) {
	if LogDebug <= l.level {
		l.printLine(message, "DEBUG")
	}
}

/*
Traceln - Print a trace message to the console.
*/
func (l *Logger) Traceln(message string) {
	if LogTrace <= l.level {
		l.printLine(message, "TRACE")
	}
}

/*--------------------------------------------------------------------------------------------------
 */
