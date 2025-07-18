package log

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

// ZeroLogger implements LoggerInterface and acts as a bridge
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
type Zerolog struct {
	zlog       zerolog.Logger
	stackTrace bool
}

func NewZerolog(zlog zerolog.Logger) *Zerolog {
	return &Zerolog{zlog: zlog}
}

// applyOptions adds data about options to the logger,
// a new instance zerolog.Logger is returned.
func applyOptions(zlog zerolog.Logger, opts ...Option) zerolog.Logger {
	if len(opts) == 0 {
		return zlog
	}

	logctx := zlog.With()
	o := Options{}
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

func (l *Zerolog) Info(msg string, opts ...Option) {
	zlog := applyOptions(l.zlog, opts...)
	zlog.Info().Msg(msg)
}

func (l *Zerolog) Error(err error, opts ...Option) {
	zlog := applyOptions(l.zlog, opts...)
	zlog.Error().Err(err).Msg("")
}

// Fatal method avoid os.Exist.
// We prefer to manage app exits on our side
func (l *Zerolog) Fatal(err error, opts ...Option) {
	zlog := applyOptions(l.zlog, opts...)
	zlog.WithLevel(zerolog.FatalLevel).Err(err).Msg("")
}

// With create a new ZeroLogger instance with the associated options
func (l *Zerolog) With(opts ...Option) LoggerInterface {
	return &Zerolog{
		zlog: applyOptions(l.zlog, opts...),
	}
}

func (l *Zerolog) WithFiberCtx(ctx *fiber.Ctx, opts ...Option) FiberLoggerInterface {
	opts = append(opts, Contexts{
		"request_": {
			"method":  ctx.Method(),
			"path":    ctx.Path(),
			"headers": ctx.GetReqHeaders(),
			"body":    string(ctx.Body()),
		},
	})

	return &Zerolog{
		zlog: applyOptions(l.zlog, opts...),
	}
}
