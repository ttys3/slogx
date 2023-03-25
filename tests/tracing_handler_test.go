package tests

import (
	"context"
	"github.com/ttys3/slogsimple"
	"github.com/ttys3/tracing-go"
	"go.opentelemetry.io/otel"
	"golang.org/x/exp/slog"
	"io"
	"testing"
)

func TestNewTracingHandler(t *testing.T) {
	logger := slogsimple.New(slogsimple.WithLevel("debug"), slogsimple.WithDisableTime(), slogsimple.WithTracing())
	slog.SetDefault(logger)

	logTracingTest(t)
}

func TestTracingFeatureDisabled(t *testing.T) {
	logger := slogsimple.New(slogsimple.WithLevel("debug"), slogsimple.WithDisableTime())
	slog.SetDefault(logger)

	logTracingTest(t)
}

func logTracingTest(t *testing.T) {
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
	log.InfoCtx(ctx, "hello world")
	log.With("foo", "bar").ErrorCtx(ctx, "have a nice day", "err", io.ErrClosedPipe)
	log.ErrorCtx(ctx, "example error", io.ErrClosedPipe)

	func() {
		ctx, span := otel.Tracer("my-tracer-name").Start(ctx, "hello.SlogSubFunc001")
		defer span.End()

		log := slog.Default()
		log.InfoCtx(ctx, "second tracing span")
		log.With("foo", "bar2").ErrorCtx(ctx, "have a nice day", "err", io.ErrClosedPipe)
		log.ErrorCtx(ctx, "example error2", "err", io.ErrClosedPipe)
	}()
}
