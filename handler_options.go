package slogsimple

import (
	"io"
	"os"
	"strings"

	"golang.org/x/exp/slog"
)

func InitDefault() {
	slog.SetDefault(slog.New(NewHandler("info", "json", "stderr")))
}

func SetDefault(level string, format string, output string) {
	slog.SetDefault(slog.New(NewHandler(level, format, output)))
}

func NewHandler(level string, format string, output string) slog.Handler {
	var w io.Writer
	switch output {
	case "stdout":
		w = os.Stdout
	case "stderr":
		w = os.Stderr
	case "discard":
		w = io.Discard
	default:
		var err error
		w, err = os.OpenFile(output, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			slog.Error("failed to open log file, fallback to stderr", err)
			w = os.Stderr
		}
	}

	var theLevel slog.Level
	switch level {
	case "debug":
		theLevel = slog.LevelDebug
	case "info":
		theLevel = slog.LevelInfo
	case "warn":
		theLevel = slog.LevelWarn
	case "error":
		theLevel = slog.LevelError
	default:
		theLevel = slog.LevelInfo
	}

	lvl := &slog.LevelVar{}
	lvl.Set(theLevel)

	h := NewHandlerOptions(lvl)
	var th slog.Handler
	switch format {
	case "text":
		th = h.NewTextHandler(w)
	case "json":
		fallthrough
	default:
		th = h.NewJSONHandler(w)
	}
	return th
}

func NewHandlerOptions(level slog.Leveler) slog.HandlerOptions {
	return slog.HandlerOptions{
		AddSource: true,
		Level:     level,

		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				// Remove time from the output.
				return slog.Attr{}
			}

			if a.Key == slog.SourceKey {

				file := a.Value.String()
				short := file

				// using short file like stdlog
				// for i := len(file) - 1; i > 0; i-- {
				// 	if file[i] == '/' {
				// 		short = file[i+1:]
				// 		break
				// 	}
				// }

				// zap like short file
				// https://github.com/uber-go/zap/blob/a55bdc32f526699c3b4cc51a2cc97e944d02fbbf/zapcore/entry.go#L102-L136
				idx := strings.LastIndexByte(file, '/')
				if idx > 0 {
					// Find the penultimate separator.
					idx = strings.LastIndexByte(file[:idx], '/')
					if idx > 0 {
						short = file[idx+1:]
					}
				}

				file = short
				return slog.Attr{
					Key:   slog.SourceKey,
					Value: slog.StringValue(file),
				}
			}
			return a
		},
	}
}
