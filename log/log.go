// Copyright 2017-present Andrea Funtò. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package log

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

// Level represents the log level.
type Level int8

const (
	// DBG is the Level for debug messages.
	DBG Level = iota
	// INF is the Level for informational messages.
	INF
	// WRN is the Level for warning messages.
	WRN
	// ERR is the Level for error messages.
	ERR
	// NUL is the Level corresponding to no log output.
	NUL
)

// logln is the prototype of log functions writing a line to a stream.
type logln func(writer io.Writer, args ...interface{}) (int, error)

// logf is the prototype of log functions writing a formatted output to a stream.
type logf func(writer io.Writer, format string, args ...interface{}) (int, error)

var (
	logLevel          Level
	logLevelLock      sync.RWMutex
	logStream         io.Writer
	logStreamLock     sync.RWMutex
	logTimeFormat     string
	logTimeFormatLock sync.RWMutex
	logColorise       bool
	logColoriseLock   sync.RWMutex
	logDebugf         logf
	logInfof          logf
	logWarnf          logf
	logErrorf         logf
	logDebugln        logln
	logInfoln         logln
	logWarnln         logln
	logErrorln        logln
)

func init() {
	SetLevel(DBG)
	SetStream(os.Stderr)
	SetTimeFormat("2006-01-02@15:04:05.000")
	if runtime.GOOS == "windows" {
		SetColorise(false)
	} else {
		SetColorise(true)
	}
}

// SetLevel sets the log level for the application.
func SetLevel(level Level) {
	logLevelLock.Lock()
	defer logLevelLock.Unlock()
	logLevel = level
}

// GetLevel retur s the current log level.
func GetLevel() Level {
	logLevelLock.RLock()
	defer logLevelLock.RUnlock()
	return logLevel
}

// SetStream sets the stream to write messages to.
func SetStream(stream io.Writer) {
	logStreamLock.Lock()
	defer logStreamLock.Unlock()
	logStream = stream
}

// GetStream returns the current log stream.
func GetStream() io.Writer {
	logStreamLock.RLock()
	defer logStreamLock.RUnlock()
	return logStream
}

// SetTimeFormat sets the format for log messages time.
func SetTimeFormat(format string) {
	logTimeFormatLock.Lock()
	defer logTimeFormatLock.Unlock()
	logTimeFormat = format
}

// GetTimeFormat returns the current format of log messages time.
func GetTimeFormat() string {
	logTimeFormatLock.RLock()
	defer logTimeFormatLock.RUnlock()
	return logTimeFormat
}

// SetColorise enables or disables the colouring of the log messages
// according to their severity. By default this is disabled on
// Windows and enabled on *NIX systems; this function is the way
// to toggle it.
func SetColorise(enabled bool) {
	logColoriseLock.Lock()
	defer logColoriseLock.Unlock()
	if enabled {
		logDebugf = color.New(color.FgWhite).Fprintf
		logInfof = color.New(color.FgGreen).Fprintf
		logWarnf = color.New(color.FgYellow).Fprintf
		logErrorf = color.New(color.FgRed).Fprintf
		logDebugln = color.New(color.FgWhite).Fprintln
		logInfoln = color.New(color.FgGreen).Fprintln
		logWarnln = color.New(color.FgYellow).Fprintln
		logErrorln = color.New(color.FgRed).Fprintln
	} else if !enabled {
		logDebugf = fmt.Fprintf
		logInfof = fmt.Fprintf
		logWarnf = fmt.Fprintf
		logErrorf = fmt.Fprintf
		logDebugln = fmt.Fprintln
		logInfoln = fmt.Fprintln
		logWarnln = fmt.Fprintln
		logErrorln = fmt.Fprintln
	}
	logColorise = enabled
}

// IsDebug returns whether the debug (DBG) log elevel is enabled.
func IsDebug() bool {
	return GetLevel() <= DBG
}

// IsInfo returns whether the informational (INF) log elevel is enabled.
func IsInfo() bool {
	return GetLevel() <= INF
}

// IsWarning returns whether the warning (WRN) log elevel is enabled.
func IsWarning() bool {
	return GetLevel() <= WRN
}

// IsError returns whether the error (ERR) log elevel is enabled.
func IsError() bool {
	return GetLevel() <= ERR
}

// IsDisabled returns whether the log is disabled.
func IsDisabled() bool {
	return GetLevel() <= NUL
}

// Debugln writes a debug message to the current output stream,
// appending a new line.
func Debugln(args ...interface{}) (int, error) {
	if IsDebug() {
		args = append([]interface{}{"[D]", time.Now().Format(GetTimeFormat()), "-"}, args...)
		return logDebugln(GetStream(), args...)
	}
	return 0, nil
}

// Infoln writes an informational message to the current output stream,
// appending a new line.
func Infoln(args ...interface{}) (int, error) {
	if IsInfo() {
		args = append([]interface{}{"[I]", time.Now().Format(GetTimeFormat()), "-"}, args...)
		return logInfoln(GetStream(), args...)
	}
	return 0, nil
}

// Warnln writes a warning message to the current output stream,
// appending a new line.
func Warnln(args ...interface{}) (int, error) {
	if IsWarning() {
		args = append([]interface{}{"[W]", time.Now().Format(GetTimeFormat()), "-"}, args...)
		return logWarnln(GetStream(), args...)
	}
	return 0, nil
}

// Errorln writes an error message to the current output stream,
// appending a new line.
func Errorln(args ...interface{}) (int, error) {
	if IsError() {
		args = append([]interface{}{"[E]", time.Now().Format(GetTimeFormat()), "-"}, args...)
		return logErrorln(GetStream(), args...)
	}
	return 0, nil
}

// Debugf writes a debug message to the current output stream,
// appending a new line.
func Debugf(format string, args ...interface{}) (int, error) {
	if IsDebug() {
		args = append([]interface{}{"[D]", time.Now().Format(GetTimeFormat())}, args...)
		format = "%s %s - " + format
		if !strings.HasSuffix(format, "\n") && !strings.HasSuffix(format, "\r") {
			format = format + "\n"
		}
		return logDebugf(GetStream(), format, args...)
	}
	return 0, nil
}

// Infof writes an informational message to the current output stream,
// appending a new line.
func Infof(format string, args ...interface{}) (int, error) {
	if IsInfo() {
		args = append([]interface{}{"[I]", time.Now().Format(GetTimeFormat())}, args...)
		format = "%s %s - " + format
		if !strings.HasSuffix(format, "\n") && !strings.HasSuffix(format, "\r") {
			format = format + "\n"
		}
		return logInfof(GetStream(), format, args...)
	}
	return 0, nil
}

// Warnf writes a warning message to the current output stream,
// appending a new line.
func Warnf(format string, args ...interface{}) (int, error) {
	if IsWarning() {
		args = append([]interface{}{"[W]", time.Now().Format(GetTimeFormat())}, args...)
		format = "%s %s - " + format
		if !strings.HasSuffix(format, "\n") && !strings.HasSuffix(format, "\r") {
			format = format + "\n"
		}
		return logWarnf(GetStream(), format, args...)
	}
	return 0, nil
}

// Errorf writes an error message to the current output stream,
// appending a new line.
func Errorf(format string, args ...interface{}) (int, error) {
	if IsError() {
		args = append([]interface{}{"[E]", time.Now().Format(GetTimeFormat())}, args...)
		format = "%s %s - " + format
		if !strings.HasSuffix(format, "\n") && !strings.HasSuffix(format, "\r") {
			format = format + "\n"
		}
		return logErrorf(GetStream(), format, args...)
	}
	return 0, nil
}

// Println is a raw version of the debug functions; it tries to interpret
// the message by checking if it starts with anthing like "[D]" or "[W]";
// if so, it delegates to the corresponding logging function, otherwise it
// just prints to the log stream as is, with no additional formatting.
func Println(args ...interface{}) (int, error) {
	if len(args) > 0 {
		if value, ok := args[0].(string); ok {
			switch {
			case strings.HasPrefix(value, "[D]"):
				return Debugln(args[1:]...)
			case strings.HasPrefix(value, "[I]"):
				return Infoln(args[1:]...)
			case strings.HasPrefix(value, "[W]"):
				return Warnln(args[1:]...)
			case strings.HasPrefix(value, "[E]"):
				return Errorln(args[1:]...)
			}
		}
	}
	return fmt.Fprintln(GetStream(), args...)
}

// Printf is a raw version of the debug functions; it tries to interpret
// the message by checking if it starts with anything like "[D]" or "[W]";
// if so, it delegates to the corresponding logging function, otherwise it
// just prints to the log stream as is, with no additional formatting.
func Printf(format string, args ...interface{}) (int, error) {
	re := regexp.MustCompile(`^\[(D|I|W|E)\]\s`)
	switch {
	case strings.HasPrefix(format, "[D]"):
		return Debugf(re.ReplaceAllString(format, ""), args...)
	case strings.HasPrefix(format, "[I]"):
		return Infof(re.ReplaceAllString(format, ""), args...)
	case strings.HasPrefix(format, "[W]"):
		return Warnf(re.ReplaceAllString(format, ""), args...)
	case strings.HasPrefix(format, "[E]"):
		return Errorf(re.ReplaceAllString(format, ""), args...)
	}
	return fmt.Fprintf(GetStream(), format, args...)
}

// ToJSON converts an object into pretty-printed JSON format.
func ToJSON(object interface{}) string {
	if bytes, err := json.MarshalIndent(object, "", "  "); err == nil {
		return string(bytes)
	}
	return ""
}
