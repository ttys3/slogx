package sslog

import (
	"fmt"
	"io"
	"os"
	"strings"

	"log/slog"
)

func InitDefault() {
	slog.SetDefault(slog.New(NewTracingHandler(NewHandler(&options{}))))
}

// New create a new *slog.Logger with tracing handler wrapper
func New(opts ...Option) *slog.Logger {
	options := &options{
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
		o(options)
	}

	h := NewHandler(options)
	if options.Tracing {
		h = NewTracingHandler(h)
	}
	return slog.New(h)
}

func NewHandler(options *options) slog.Handler {
	var w io.Writer
	if options.Writer != nil {
		w = options.Writer
	} else {
		switch options.Output {
		case "stdout":
			w = os.Stdout
		case "stderr", "":
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

	opts := NewHandlerOptions(lvl, &options.HandlerOptions)
	var th slog.Handler
	switch options.Format {
	case "text":
		th = slog.NewTextHandler(w, &opts)
	case "json":
		fallthrough
	default:
		th = slog.NewJSONHandler(w, &opts)
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
				if src, ok := a.Value.Any().(*slog.Source); ok {
					// File is the absolute path to the file
					short := src.File

					// using short file like stdlog
					// for i := len(file) - 1; i > 0; i-- {
					// 	if file[i] == '/' {
					// 		short = file[i+1:]
					// 		break
					// 	}
					// }

					// zap like short file
					// https://github.com/uber-go/zap/blob/a55bdc32f526699c3b4cc51a2cc97e944d02fbbf/zapcore/entry.go#L102-L136
					idx := strings.LastIndexByte(src.File, '/')
					if idx > 0 {
						// Find the penultimate separator.
						idx = strings.LastIndexByte(src.File[:idx], '/')
						if idx > 0 {
							short = src.File[idx+1:]
						}
					}

					file = fmt.Sprintf("%s:%d", short, src.Line)
				}

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
