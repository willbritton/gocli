package gocli

import (
	"io"
	"log"
)

type Logger struct {
	base    *log.Logger
	verbose bool
}

func NewLogger() *Logger {
	l := &Logger{base: log.Default()}
	l.base.SetFlags(0)
	return l
}

var logger *Logger = NewLogger()

func DefaultLogger() *Logger {
	return logger
}

func (l *Logger) Wrap(b *log.Logger) {
	l.base = b
}

func (l *Logger) SetVerbose() {
	l.verbose = true
}

func (l *Logger) SetQuiet() {
	l.verbose = false
}

// SetOutput sets the output destination for the logger.
func (l *Logger) SetOutput(w io.Writer) {
	if l.base != nil {
		l.base.SetOutput(w)
	}
}

// Output writes the output for a logging event. The string s contains
// the text to print after the prefix specified by the flags of the
// Logger. A newline is appended if the last character of s is not
// already a newline. Calldepth is used to recover the PC and is
// provided for generality, although at the moment on all pre-defined
// paths it will be 2.
func (l *Logger) Output(calldepth int, s string) error {
	if l.base != nil && l.verbose {
		return l.base.Output(calldepth, s)
	}
	return nil
}

// Printf calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Printf(format string, v ...any) {
	if l.base != nil && l.verbose {
		l.base.Printf(format, v...)
	}
}

// Print calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Print(v ...any) {
	if l.base != nil && l.verbose {
		l.base.Print(v...)
	}
}

// Println calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Println.
func (l *Logger) Println(v ...any) {
	if l.base != nil && l.verbose {
		l.base.Println(v...)
	}
}

// Fatal is equivalent to l.Print() followed by a call to os.Exit(1).
func (l *Logger) Fatal(v ...any) {
	if l.base != nil {
		l.base.Fatal(v...)
	}
}

// Fatalf is equivalent to l.Printf() followed by a call to os.Exit(1).
func (l *Logger) Fatalf(format string, v ...any) {
	if l.base != nil {
		l.base.Fatalf(format, v...)
	}
}

// Fatalln is equivalent to l.Println() followed by a call to os.Exit(1).
func (l *Logger) Fatalln(v ...any) {
	if l.base != nil {
		l.base.Fatalln(v...)
	}
}

// Panic is equivalent to l.Print() followed by a call to panic().
func (l *Logger) Panic(v ...any) {
	if l.base != nil {
		l.base.Panic(v...)
	}
}

// Panicf is equivalent to l.Printf() followed by a call to panic().
func (l *Logger) Panicf(format string, v ...any) {
	if l.base != nil {
		l.base.Panicf(format, v...)
	}
}

// Panicln is equivalent to l.Println() followed by a call to panic().
func (l *Logger) Panicln(v ...any) {
	if l.base != nil {
		l.base.Panicln(v...)
	}
}

// Flags returns the output flags for the logger.
// The flag bits are Ldate, Ltime, and so on.
func (l *Logger) Flags() int {
	if l.base != nil {
		return l.base.Flags()
	}
	return 0
}

// SetFlags sets the output flags for the logger.
// The flag bits are Ldate, Ltime, and so on.
func (l *Logger) SetFlags(flag int) {
	if l.base != nil {
		l.base.SetFlags(flag)
	}
}

// Prefix returns the output prefix for the logger.
func (l *Logger) Prefix() string {
	if l.base != nil {
		return l.base.Prefix()
	}
	return ""
}

// SetPrefix sets the output prefix for the logger.
func (l *Logger) SetPrefix(prefix string) {
	if l.base != nil {
		l.base.SetPrefix(prefix)
	}
}

// Writer returns the output destination for the logger.
func (l *Logger) Writer() io.Writer {
	if l.base != nil {
		return l.base.Writer()
	}
	return nil
}
