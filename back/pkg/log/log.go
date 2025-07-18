package log

import (
	"maps"

	"github.com/getsentry/sentry-go"
	"github.com/gofiber/fiber/v2"
)

type Options struct {
	Contexts Contexts
	Tags     Tags
	User     *User
}

// This interface allows the implemented logger to retrieve additional data,
// Apply it to a pointed Options structure and then process it.
type Option interface {
	Apply(*Options)
}

// LoggerInterface defines minimal logger implementation methods
// for Go projects on this repo, and must be the only logging dependency.
type LoggerInterface interface {
	//	Record a given message or error with optional contextual datas.
	Info(msg string, opts ...Option)
	Error(err error, opts ...Option)
	Fatal(err error, opts ...Option)
	// The aim of this method is to clone the logger and to be able to introduce
	// additional elements so that they are available for potential future logging.
	// usage:
	//
	//	// instantiates a copy of the logger with the associated datas
	//	sublogger := logger.With(Tags{"key", "value"})
	//	// all logs using the sublogger manage the additional datas added previously
	//	sublogger.Info("message")
	With(opts ...Option) LoggerInterface
}

type FiberLoggerInterface interface {
	LoggerInterface
	//	this method must be able to clone the logger and retreive what it needs from a fiber.Ctx.
	WithFiberCtx(ctx *fiber.Ctx, opts ...Option) FiberLoggerInterface
}

// SentryLoggerInterface defines additional sentry-specific methods.
// The idea is to use it this way:
//
//	func todo(logger LoggerInterface) {
//		if logger, ok := logger.(SentryLoggerInterface); ok {
//			// use logger as SentryLoggerInterface
//		}
//	}
type SentryLoggerInterface interface {
	FiberLoggerInterface
	// like LoggerInterface.With, this method must be able to clone the logger
	//	so that the new instance has a sentry hub available for future logging.
	WithHub(hub *sentry.Hub, opts ...Option) SentryLoggerInterface
	//	just a gateway method to breadscumb sentry use
	AddBreadcrumb(*sentry.Breadcrumb, *sentry.BreadcrumbHint)
	ClearBreadcrumbs()
}

type User struct {
	ID       string            `json:"id"`
	Email    string            `json:"email"`
	Username string            `json:"username"`
	Data     map[string]string `json:"data"`
}

func (u *User) Apply(opts *Options) {
	opts.User = u
}

type Contexts map[string]map[string]any

func (c Contexts) Apply(opts *Options) {
	if opts.Contexts == nil {
		opts.Contexts = c
	} else {
		maps.Copy(opts.Contexts, c)
	}
}

type Tags map[string]string

func (t Tags) Apply(opts *Options) {
	opts.Tags = t

	if opts.Tags == nil {
		opts.Tags = t
	} else {
		maps.Copy(opts.Tags, t)
	}
}
