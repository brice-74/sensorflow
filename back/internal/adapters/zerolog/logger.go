package zerolog

import (
	"github.com/brice-74/sensorflow/internal/log"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

// Logger implements log.Logger interface and acts as a bridge
// between the interface and zerolog.Logger (https://github.com/rs/zerolog).
// The idea is to keep zerolog's stateless* nature, to preserve its efficiency
// in terms of speed and its low memory allocation.
//
// But these features have their drawbacks, such as the inability to edit
// data previously added to the options (because the content is already written)
// and the possibility of creating duplicate keys in the json.
//
// *stateless logger: A logger that doesn't accumulate context or additional
// data in memory between log calls, allowing optimized direct
// writing without memory overhead.
type Logger struct {
	zlog zerolog.Logger
}

var _ log.Fiber = (*Logger)(nil)

func NewLogger(zlog zerolog.Logger) *Logger {
	return &Logger{zlog: zlog}
}

// applyOptions adds data about options to the logger,
// a new instance zerolog.Logger is returned.
func applyOptions(zlog zerolog.Logger, opts ...log.Option) zerolog.Logger {
	if len(opts) == 0 {
		return zlog
	}

	logctx := zlog.With()
	o := log.Options{}
	for _, opt := range opts {
		opt.Apply(&o)
	}

	for k, v := range o.Contexts {
		logctx = logctx.Any("ctx_"+k, v)
	}

	for k, v := range o.Tags {
		logctx = logctx.Str("tag_"+k, v)
	}

	if o.User != nil {
		logctx = logctx.Any("user", o.User)
	}

	return logctx.Logger()
}

func (l *Logger) Info(msg string, opts ...log.Option) {
	zlog := applyOptions(l.zlog, opts...)
	zlog.Info().Msg(msg)
}

func (l *Logger) Warn(err error, opts ...log.Option) {
	zlog := applyOptions(l.zlog, opts...)
	zlog.Warn().Err(err).Msg("")
}

func (l *Logger) Error(err error, opts ...log.Option) {
	zlog := applyOptions(l.zlog, opts...)
	zlog.Error().Err(err).Msg("")
}

// Fatal method avoid os.Exist.
// We prefer to manage app exits on our side
func (l *Logger) Fatal(err error, opts ...log.Option) {
	zlog := applyOptions(l.zlog, opts...)
	zlog.WithLevel(zerolog.FatalLevel).Err(err).Msg("")
}

// With create a new Logger instance with the associated options
func (l *Logger) With(opts ...log.Option) log.Logger {
	return &Logger{
		zlog: applyOptions(l.zlog, opts...),
	}
}

func (l *Logger) WithFiberCtx(ctx *fiber.Ctx, opts ...log.Option) log.Fiber {
	opts = append(opts, log.Contexts{
		"request_": {
			"method":  ctx.Method(),
			"path":    ctx.Path(),
			"headers": ctx.GetReqHeaders(),
			"body":    string(ctx.Body()),
		},
	})

	return &Logger{
		zlog: applyOptions(l.zlog, opts...),
	}
}
