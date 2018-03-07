# go-log - Yet another logger wrapper for golang

A simple logger for Go; it provides log messages colorising, automatic addition or source file, line number and/or calling function to log messages. All top level messages are synchronised, so it is safe to reconfigure the logger from different goroutines.

## Usage

To import the logger, simply add it as follows:
``` golang
import (
	"github.com/dihedron/go-log"
)
```
All methods are available under the ```log``` namespace. To configure the logger, use a combination of the following methods:
``` golang
	log.SetLevel(log.DBG)
	log.SetStream(os.Stdout, true)
	log.SetTimeFormat("15:04:05.000")
	log.SetPrintCallerInfo(true)
	log.SetPrintSourceInfo(log.SourceInfoShort)
```
where ```log.SetLevel()``` sets the current logging level to one of ```log.DBG``` (debugging messages or higher), ```log.INF``` (informational messages or more severe), ```log.WRN``` (warning messages or more severe), ```log.ERR``` (error messages only) or ```log.NUL``` (no messages at all).  

```log.SetStream()``` sets the ```io.Writer``` to which messages will be output; it can be ```os.Stdout``` or ```os.Stderr```, or it can be a file on disk or a socket. The second boolean parameter specifies whether messages should be colorised according to their severity; this really only applies to console output.  

```log.SetTimeFormat()``` sets the format for timestamps; the suggested format provides timestamping to the milliseconds.  

```log.SetPrintCallerInfo()``` instructs the logger to write the name of the calling method before the message; the name is retrieved at runtime by walking the stack, so it is quite cumbersome and can result in a significant slowdown.  

```log.SetPrintSourceInfo()``` instructs the logger to print the name of the file (```log.SourceInfoShort```) or the full path (```log.SourceInfoLong```) and the line number of the call site. Also this information is retrieved at runtime by walking the stack and can be quite cumbersome: use sparingly!  

To actually log messages, you can use two families of functions which follow the ```fmt.Printf``` and ```fmt.Println``` usage patterns, e.g.:
``` golang
log.Errorf("this is an error message: %v", err)

log.Infoln("this is an informational message")
```

## License

The code is released under an MIT License. All contributions are welcome provided they don't decrease the coverage of unit tests and are in line with the style of the rest of the library.


