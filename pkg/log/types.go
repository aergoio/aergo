/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package log

type ILogger interface {

	// SetLevel changes the logging level to the passed level.
	SetLevel(logLevel string)

	// Debugf formats message according to format specifier and writes to
	// log with DebugLvl.
	Debugf(format string, args ...interface{})

	// Infof formats message according to format specifier and writes to
	// log with InfoLvl.
	Infof(format string, args ...interface{})

	// Warnf formats message according to format specifier and writes to
	// to log with WarnLvl.
	Warnf(format string, args ...interface{})

	// Errorf formats message according to format specifier and writes to
	// to log with ErrorLvl.
	Errorf(format string, args ...interface{})

	// Fatalf formats message according to format specifier and writes to
	// log with FatalLvl, and program exits.
	Fatalf(format string, args ...interface{})

	// Debug formats message using the default formats for its operands
	// and writes to log with DebugLvl.
	Debug(args ...interface{})

	// Info formats message using the default formats for its operands
	// and writes to log with InfoLvl.
	Info(args ...interface{})

	// Warn formats message using the default formats for its operands
	// and writes to log with WarnLvl.
	Warn(args ...interface{})

	// Error formats message using the default formats for its operands
	// and writes to log with ErrorLvl.
	Error(args ...interface{})

	// Fatal formats message using the default formats for its operands
	// and writes to log with FatalLvl, and program exits.
	Fatal(args ...interface{})

	// Level returns the current logging level.
	Level() string

	IsDebugEnabled() bool
}

type Levels struct {
	Default string
	Module  map[string]string
}
