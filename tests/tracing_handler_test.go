package tests

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"os"
	"testing"

	"github.com/ttys3/slogsimple"
	"github.com/ttys3/tracing-go"
	"go.opentelemetry.io/otel"
)

func TestNewTracingHandler(t *testing.T) {
	var buf bytes.Buffer
	mw := io.MultiWriter(&buf, os.Stderr)
	logger := slogsimple.New(slogsimple.WithLevel("debug"),
		slogsimple.WithDisableTime(),
		slogsimple.WithTracing(),
		slogsimple.WithWriter(mw))
	slog.SetDefault(logger)

	ctx := context.Background()
	// set up a recording tracer, non-recording span will not get a trace_id
	tpShutdown, err := tracing.InitProvider(ctx, tracing.WithStdoutTrace())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		tpShutdown(context.Background())
	})

	ctx, newSpan := otel.Tracer("my-tracer-name").Start(ctx, "hello.Slog")
	defer newSpan.End()

	log := slog.With("foo", "bar")
	log.InfoContext(ctx, "hello world")
	checkLogOutput(t, buf.String(), `{"level":"INFO","source":"tests/tracing_handler_test.go:\d+","msg":"hello world","foo":"bar","trace_id":"\w+"}`)
	buf.Reset()

	log.With("foo", "bar").ErrorContext(ctx, "have a nice day", "err", io.ErrClosedPipe)
	checkLogOutput(t, buf.String(), `{"level":"ERROR","source":"tests/tracing_handler_test.go:\d+","msg":"have a nice day","foo":"bar","foo":"bar","err":"io: read/write on closed pipe","trace_id":"\w+"}`)
	buf.Reset()

	log.ErrorContext(ctx, "example error", "err", io.ErrClosedPipe)
	checkLogOutput(t, buf.String(), `{"level":"ERROR","source":"tests/tracing_handler_test.go:\d+","msg":"example error","foo":"bar","err":"io: read/write on closed pipe","trace_id":"\w+"}`)
	buf.Reset()

	func() {
		ctx, span := otel.Tracer("my-tracer-name").Start(ctx, "hello.SlogSubFunc001")
		defer span.End()

		log := slog.Default()
		log.InfoContext(ctx, "second tracing span")

		log.With("foo", "bar2").ErrorContext(ctx, "have a nice day", "err", io.ErrClosedPipe)

		log.ErrorContext(ctx, "example error2", "err", io.ErrClosedPipe)
	}()
}

func TestTracingFeatureDisabled(t *testing.T) {
	var buf bytes.Buffer
	mw := io.MultiWriter(&buf, os.Stderr)
	logger := slogsimple.New(slogsimple.WithLevel("debug"), slogsimple.WithDisableTime(), slogsimple.WithWriter(mw))
	slog.SetDefault(logger)

	ctx := context.Background()
	// set up a recording tracer, non-recording span will not get a trace_id
	tpShutdown, err := tracing.InitProvider(ctx, tracing.WithStdoutTrace())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		tpShutdown(context.Background())
	})

	ctx, newSpan := otel.Tracer("my-tracer-name").Start(ctx, "hello.Slog")
	defer newSpan.End()

	log := slog.With("foo", "bar")
	log.InfoContext(ctx, "hello world")
	checkLogOutput(t, buf.String(), `{"level":"INFO","source":"tests/tracing_handler_test.go:\d+","msg":"hello world","foo":"bar"}`)
	buf.Reset()

	log.With("foo", "bar").ErrorContext(ctx, "have a nice day", "err", io.ErrClosedPipe)
	log.ErrorContext(ctx, "example error", "err", io.ErrClosedPipe)

	func() {
		ctx, span := otel.Tracer("my-tracer-name").Start(ctx, "hello.SlogSubFunc001")
		defer span.End()

		log := slog.Default()
		log.InfoContext(ctx, "second tracing span")
		log.With("foo", "bar2").ErrorContext(ctx, "have a nice day", "err", io.ErrClosedPipe)
		log.ErrorContext(ctx, "example error2", "err", io.ErrClosedPipe)
	}()
}
