// Package log wraps log/slog for more easier and convinient use.
package log

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"time"
)

var (
	// Handler exposes slog handler struct.
	Handler slog.Handler

	// Ctx log context. For now it is just sits here and do nothing.
	Ctx = context.Background()

	// Writer go-writer eported for capturing other types of logs, even for std log go facility.
	Writer io.Writer

	// reader is required for Writer to work properly. It is connected via io.Pipe() to Writer.
	reader io.Reader

	// Logger *log.Logger thing required for compatibility with old-good log package.
	Logger log.Logger

	// LogFileHandler is currently opened log file descriptor. Stderr by default.
	LogFileHandler = os.Stderr

	// LogLevel current loglevel. Info by default.
	LogLevel = slog.LevelInfo
)

// Init setup logger stuff.
// level can be error, warn, info, debug if something other supplied info level selected.
// fileDescriptor should be opened before supplying it to Init().
func Init(level string, fileDescriptor *os.File) {
	LogFileHandler = fileDescriptor

	// no panic, no trace.
	switch level {
	case "debug":
		LogLevel = slog.LevelDebug

	case "info":
		LogLevel = slog.LevelInfo

	case "warn":
		LogLevel = slog.LevelWarn

	case "error":
		LogLevel = slog.LevelError

	default:
		LogLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{ //nolint: exhaustruct
		// Use the ReplaceAttr function on the handler options
		// to be able to replace any single attribute in the log output.
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr { //nolint: revive
			// Check that we are handling the time key.
			if a.Key != slog.TimeKey {
				return a
			}

			t := a.Value.Time()

			// change the value from a time.Time to a String
			// where the string has the correct time format.
			a.Value = slog.StringValue(t.Format(time.DateTime))

			return a
		},

		Level: LogLevel,
	}

	// Documentation says that we can use *os.File in place of io.Writer, but in this case it did not work.
	// So we have to cast fileDescriptor to io.Writer.
	w := io.Writer(fileDescriptor)

	Handler = slog.NewTextHandler(w, opts)

	slog.SetDefault(
		slog.New(Handler),
	)

	// Catch log.Logger message at info loglevel of slog logging facility.
	Logger = *slog.NewLogLogger(Handler, slog.LevelError)

	reader, Writer = io.Pipe()

	go readIoPipe()
}

// readIoPipe reads pipe and writes it to slog.
func readIoPipe() {
	scanner := bufio.NewScanner(reader)

	for {
		for scanner.Scan() {
			line := scanner.Text() // Get the current line as a string.
			slog.Debug(line)
		}

		if err := scanner.Err(); err != nil {
			// io.EOF is expected when done reading, e.g. on exit.
			if errors.Is(err, io.EOF) {
				return
			}

			// something weird happen, we should it log to guess later what it was.
			Errorf("Error during scanning log writer pipe: %v\n", err)
		}
	}
}

// GetLevel returns current loglevel as string.
func GetLevel() string {
	switch {
	case Handler.Enabled(Ctx, slog.LevelDebug):
		return "debug"

	case Handler.Enabled(Ctx, slog.LevelInfo):
		return "info"

	case Handler.Enabled(Ctx, slog.LevelWarn):
		return "warn"

	case Handler.Enabled(Ctx, slog.LevelError):
		return "error"

	default:
		return "unknown"
	}
}

// Fatal passes message directly to slog.Error() and exits with error code 1.
func Fatal(message string) {
	slog.Error(message)
	os.Exit(1)
}

// Fatalf logs message on error log level and exits with error code 1. It allows to supply arguments in printf() style .
func Fatalf(format string, a ...any) {
	slog.Error(fmt.Sprintf(format, a...))
	os.Exit(1)
}

// Error passes message directly to slog.Error().
func Error(message string) {
	slog.Error(message)
}

// Errorf logs message on error log level and allows to supply arguments in printf() style.
func Errorf(format string, a ...any) {
	slog.Error(fmt.Sprintf(format, a...))
}

// Warn passes message directly to slog.Warn().
func Warn(message string) {
	slog.Warn(message)
}

// Warnf logs message on warn log level and allows to supply arguments in printf() style.
func Warnf(format string, a ...any) {
	slog.Warn(fmt.Sprintf(format, a...))
}

// Info passes message directly to slog.Info().
func Info(message string) {
	slog.Info(message)
}

// Infof logs message on info log level and allows to supply arguments in printf() style.
func Infof(format string, a ...any) {
	slog.Info(fmt.Sprintf(format, a...))
}

// Debug passes message directly to slog.Debug().
func Debug(message string) {
	slog.Debug(message)
}

// Debugf logs message on debug log level and allows to supply arguments in printf() style.
func Debugf(format string, a ...any) {
	slog.Debug(fmt.Sprintf(format, a...))
}

// Close closes opened log file descriptors.
func Close() {
	LogFileHandler.Close()
}

// ReOpenLog closes and opens log file again.
func ReOpenLog() {
	filename := LogFileHandler.Name()

	// If we writing to stderr, we should not re-open it.
	if filename == os.Stderr.Name() {
		return
	}

	loglevel := GetLevel()

	LogFileHandler.Close()

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)

	if err != nil {
		// Try to emulate our own log message format.
		fmt.Fprintf(
			os.Stderr,
			"time=\"%s\" level=ERROR msg=\"Unable to open logfile %s: %s\"",
			time.Now().Format(time.DateTime),
			filename,
			err,
		)

		return
	}

	Init(loglevel, file)
}

/* vim: set ft=go noet ai ts=4 sw=4 sts=4: */
