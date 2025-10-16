package sentry

import (
	"time"

	"github.com/brice-74/sensorflow/internal/log"
	"github.com/getsentry/sentry-go"
	sentryfiber "github.com/getsentry/sentry-go/fiber"
	"github.com/gofiber/fiber/v2"
)

// Logger implements log.Sentry interface and ensures that a sentry hub
// is always available. The hub is used to capture events and errors, and to
// manage scopes (contexts, users, etc.). A hub in Logger acts as a container
// for scope information (context) for each group of logs (such as a thread
// or an HTTP request).
type Logger struct {
	hub *sentry.Hub
}

var _ log.Logger = (*Logger)(nil)

// NewLogger creates a new instance of Logger with a specified
// Logger Client Options. So a Logger can contain only on sentry Client.
// This function forward the sentry error from the client creation.
func NewLogger(options sentry.ClientOptions) (*Logger, error) {
	client, err := sentry.NewClient(options)
	if err != nil {
		return nil, err
	}
	return &Logger{hub: sentry.NewHub(client, sentry.NewScope())}, nil
}

// Flush waits until the underlying Client Transport sends any buffered events to the sentry server
// It must be use just one time in the carrying goroutine
func (l *Logger) Flush(timeout time.Duration) bool {
	return l.hub.Client().Flush(timeout)
}

func (l *Logger) GetClient() *sentry.Client {
	return l.hub.Client()
}

// applyOptions applies various options (contexts, tags, user) to the
// current scope of the Logger hub.
func (l *Logger) applyOptions(opts ...log.Option) {
	o := log.Options{}
	for _, opt := range opts {
		opt.Apply(&o)
	}

	scope := l.hub.Scope()

	if o.User != nil {
		scope.SetUser(sentry.User{
			ID:       o.User.ID,
			Username: o.User.Username,
			Email:    o.User.Email,
			Data:     o.User.Data,
		})
	}

	scope.SetTags(o.Tags)
	scope.SetContexts(o.Contexts)
}

func (l *Logger) Info(msg string, opts ...log.Option) {
	l.applyOptions(opts...)
	l.hub.Scope().SetLevel(sentry.LevelInfo)
	l.hub.CaptureMessage(msg)
}

func (l *Logger) Error(err error, opts ...log.Option) {
	l.applyOptions(opts...)
	l.hub.Scope().SetLevel(sentry.LevelError)
	l.hub.CaptureException(err)
}

func (l *Logger) Fatal(err error, opts ...log.Option) {
	l.applyOptions(opts...)
	l.hub.Scope().SetLevel(sentry.LevelFatal)
	l.hub.CaptureException(err)
}

// clone creates a new instance of Logger with a new hub (or clones the current one)
// and applies the provided options. This ensures that each cloned logger can have its own
// independent scope, without interfering with the parent logger's hub.
func (l *Logger) clone(hub *sentry.Hub, opts ...log.Option) *Logger {
	if hub == nil {
		hub = l.hub.Clone()
	}
	clone := &Logger{
		hub: hub,
	}
	clone.applyOptions(opts...)
	return clone
}

// With create a new Logger containing cloned hub from the parent with the associated options
func (l *Logger) With(opts ...log.Option) log.Logger {
	return l.clone(nil, opts...)
}

// WithHub acts like the With method, but you can specify your hub.
func (l *Logger) WithHub(hub *sentry.Hub, opts ...log.Option) log.Sentry {
	return l.clone(hub, opts...)
}

// WithFiberCtx acts like the With method, but you can specify a fiber Ctx to retreive hub
// and some contextual request datas, see github.com/getsentry/sentry-go/fiber.
func (l *Logger) WithFiberCtx(ctx *fiber.Ctx, opts ...log.Option) log.Fiber {
	return l.clone(sentryfiber.GetHubFromContext(ctx), opts...)
}

// AddBreadcrumb adds a breadcrumb to the current Logger scope.
func (l *Logger) AddBreadcrumb(breadcrumb *sentry.Breadcrumb, hint *sentry.BreadcrumbHint) {
	l.hub.AddBreadcrumb(breadcrumb, hint)
}

// ClearBreadcrumbs removes all breadcrumbs from the current scope.
func (l *Logger) ClearBreadcrumbs() {
	l.hub.Scope().ClearBreadcrumbs()
}
