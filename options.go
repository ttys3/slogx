package slogsimple

// Option is an application option.
type Option func(o *options)

type HandlerOptions struct {
	DisableSource bool
	FullSource    bool
	DisableTime   bool
}

// options is an application options.
type options struct {
	HandlerOptions

	Level   string // debug, info, warn, error
	Format  string // json, text
	Output  string // stdout, stderr, discard, or a file path
	Tracing bool   // enable tracing feature
}

func WithDisableSource() Option {
	return func(o *options) { o.DisableSource = true }
}

func WithFullSource() Option {
	return func(o *options) { o.FullSource = true }
}

func WithDisableTime() Option {
	return func(o *options) { o.DisableTime = true }
}

func WithLevel(level string) Option {
	return func(o *options) {
		if level == "" {
			level = "info"
		}
		o.Level = level
	}
}

func WithFormat(format string) Option {
	return func(o *options) {
		if format == "" {
			format = "json"
		}
		o.Format = format
	}
}

func WithOutput(output string) Option {
	return func(o *options) {
		if output == "" {
			output = "stderr"
		}
		o.Output = output
	}
}

func WithTracing() Option {
	return func(o *options) { o.Tracing = true }
}
