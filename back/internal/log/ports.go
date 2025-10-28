package log

import (
	"github.com/getsentry/sentry-go"
	"github.com/gofiber/fiber/v2"
)

// This interface allows the implemented logger to retrieve additional data,
// Apply it to a pointed Options structure and then process it.
type Option interface {
	Apply(*Options)
}

// Logger defines minimal logger implementation methods
// for Go projects on this repo, and must be the only logging dependency.
type Logger interface {
	//	Record a given message or error with optional contextual datas.
	Info(msg string, opts ...Option)
	Warn(err error, opts ...Option)
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
	With(opts ...Option) Logger
}

type Fiber interface {
	Logger
	//	this method must be able to clone the logger and retreive what it needs from a fiber.Ctx.
	WithFiberCtx(ctx *fiber.Ctx, opts ...Option) Fiber
}

// SentryLogger defines additional sentry-specific methods.
// The idea is to use it this way:
//
//	func todo(logger Logger) {
//		if logger, ok := logger.(SentryLogger); ok {
//			// use logger as SentryLogger
//		}
//	}
type Sentry interface {
	Fiber
	// like Logger.With, this method must be able to clone the logger
	//	so that the new instance has a sentry hub available for future logging.
	WithHub(hub *sentry.Hub, opts ...Option) Sentry
	//	just a gateway method to breadscumb sentry use
	AddBreadcrumb(*sentry.Breadcrumb, *sentry.BreadcrumbHint)
	ClearBreadcrumbs()
}
