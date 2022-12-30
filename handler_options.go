package slogsimple

import (
	"io"
	"os"
	"strings"

	"golang.org/x/exp/slog"
)

func InitDefault() {
	slog.SetDefault(slog.New(NewTracingHandler(NewHandler())))
}

// New create a new *slog.Logger with tracing handler wrapper
func New(opts ...Option) *slog.Logger {
	return slog.New(NewTracingHandler(NewHandler(opts...)))
}

func NewHandler(opts ...Option) slog.Handler {
	options := options{
		HandlerOptions: HandlerOptions{
			DisableSource: false,
			FullSource:    false,
			DisableTime:   false,
		},
		Level:  "info",
		Format: "json",
		Output: "stderr",
	}
	for _, o := range opts {
		o(&options)
	}

	var w io.Writer
	switch options.Output {
	case "stdout":
		w = os.Stdout
	case "stderr":
		w = os.Stderr
	case "discard":
		w = io.Discard
	default:
		var err error
		w, err = os.OpenFile(options.Output, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			slog.Error("failed to open log file, fallback to stderr", err)
			w = os.Stderr
		}
	}

	var theLevel slog.Level
	switch options.Level {
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

	h := NewHandlerOptions(lvl, &options.HandlerOptions)
	var th slog.Handler
	switch options.Format {
	case "text":
		th = h.NewTextHandler(w)
	case "json":
		fallthrough
	default:
		th = h.NewJSONHandler(w)
	}
	return th
}

func NewHandlerOptions(level slog.Leveler, opt *HandlerOptions) slog.HandlerOptions {
	ho := slog.HandlerOptions{
		AddSource: !opt.DisableSource,
		Level:     level,
	}

	if !opt.DisableTime && (opt.FullSource || opt.DisableSource) {
		return ho
	}

	ho.ReplaceAttr = func(groups []string, a slog.Attr) slog.Attr {
		if opt.DisableTime {
			if a.Key == slog.TimeKey {
				// Remove time from the output.
				return slog.Attr{}
			}
		}

		// handle short source file location
		if !opt.DisableSource && !opt.FullSource {
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
		}

		return a
	}
	return ho
}
