# slogsimple
slog handler with opentelemetry tracing support and simple init helper method

## usage

```go
package main

import (
    "context"
    "github.com/ttys3/slogsimple"
    "github.com/ttys3/tracing-go"
    "go.opentelemetry.io/otel"
    "golang.org/x/exp/slog"
    "io"
)

func main() {
	// init a new slog json handler at info level with output to stderr
	// and wrap it within a tracing handler
	logger := slogsimple.New(slogsimple.WithTracing())
	slog.SetDefault(logger)

	ctx := context.Background()

	// set up a recording tracer, non-recording span will not get a trace_id
	tpShutdown, err := tracing.InitProvider(ctx, tracing.WithStdoutTrace())
	if err != nil {
		panic(err)
	}
	defer tpShutdown(context.Background())

	ctx, newSpan := otel.Tracer("my-tracer-name").Start(ctx, "hello.Slog")
	defer newSpan.End()
	ctxLog := slog.With("foo", "bar")
	ctxLog.InfoCtx(ctx, "hello world")
	ctxLog.With("foo", "bar").ErrorCtx(ctx, "have a nice day", io.ErrClosedPipe)
	ctxLog.ErrorCtx(ctx, "example error", io.ErrClosedPipe)
}
```
