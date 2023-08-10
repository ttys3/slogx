package sslog

import (
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/ttys3/sslog/internal"
	"io"
	"log/slog"
	"sync"
)

var bold = color.New(color.Bold)

func init() {
	bold.EnableColor()
}

// Colors mapping.
var Colors = map[slog.Level]*color.Color{
	slog.LevelDebug: color.New(color.FgWhite),
	slog.LevelInfo:  color.New(color.FgBlue),
	slog.LevelWarn:  color.New(color.FgYellow),
	slog.LevelError: color.New(color.FgRed),
}

// Strings mapping.
var Strings = map[slog.Level]string{
	slog.LevelDebug: "•",
	slog.LevelInfo:  "•",
	slog.LevelWarn:  "•",
	slog.LevelError: "⨯",
}

type CliHandler struct {
	mu sync.Mutex
	w  io.Writer

	opts *CliHandlerOptions

	attrsPrefix []slog.Attr

	groupPrefix string
}

type CliHandlerOptions struct {
	ColoredLevel bool
	slog.HandlerOptions
}

func NewCliHandler(w io.Writer, opts *CliHandlerOptions) *CliHandler {
	return &CliHandler{w: w, opts: opts}
}

func (h *CliHandler) Enabled(ctx context.Context, l slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}
	return l >= minLevel
}

func (h *CliHandler) Handle(ctx context.Context, r slog.Record) error {
	// level time message attributes// get a buffer from the sync pool
	// get a buffer from the sync pool
	buf := internal.NewBuffer()
	defer buf.Free()

	theColor := Colors[r.Level]
	theColor.EnableColor()

	if h.opts.ColoredLevel {
		levelEmoji := Strings[r.Level]
		padding := 4
		coloredLevel := theColor.Sprintf("%s", bold.Sprintf("%*s", padding, levelEmoji))
		buf.WriteString(coloredLevel)
	} else {
		buf.WriteString(r.Level.String())
	}

	buf.WriteString(" ")
	buf.WriteString(fmt.Sprintf("%-25s", r.Message))

	buf.WriteString("\t\t")

	// write handler attributes
	if len(h.attrsPrefix) > 0 {
		for _, attr := range h.attrsPrefix {
			h.appendAttr(buf, attr, theColor, h.groupPrefix)
		}
	}

	// buf.WriteString(" ")

	// write attributes
	if r.NumAttrs() > 0 {
		r.Attrs(func(attr slog.Attr) bool {
			h.appendAttr(buf, attr, theColor, h.groupPrefix)
			return true
		})
	}

	buf.WriteByte('\n')

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := h.w.Write(buf.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func (h *CliHandler) appendAttr(buf *internal.Buffer, attr slog.Attr, theColor *color.Color, groupsPrefix string) {
	theColor.EnableColor()

	buf.Write([]byte(" "))
	if groupsPrefix != "" {
		buf.WriteString(theColor.Sprint(groupsPrefix))
		// buf.WriteString(theColor.Sprint("."))
	}
	buf.WriteString(theColor.Sprint(attr.Key))
	buf.Write([]byte("="))

	// needQuote := attr.Value.Kind() != slog.KindInt64
	// if needQuote {
	// 	buf.Write([]byte(`"`))
	// }
	buf.Write([]byte(attr.Value.String()))
	// if needQuote {
	// 	buf.Write([]byte(`"`))
	// }
}

func (h *CliHandler) clone() *CliHandler {
	attrsPrefix := make([]slog.Attr, len(h.attrsPrefix))
	copy(attrsPrefix, h.attrsPrefix)
	return &CliHandler{w: h.w, opts: h.opts, attrsPrefix: attrsPrefix, groupPrefix: h.groupPrefix}
}

func (h *CliHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}

	cloned := h.clone()
	cloned.attrsPrefix = append(cloned.attrsPrefix, attrs...)
	return cloned
}

func (h *CliHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	cloned := h.clone()
	cloned.groupPrefix += name + "."
	return cloned
}

var _ slog.Handler = (*CliHandler)(nil)
