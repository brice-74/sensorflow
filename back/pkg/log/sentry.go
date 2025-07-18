package log

import (
	"time"

	"github.com/getsentry/sentry-go"
	sentryfiber "github.com/getsentry/sentry-go/fiber"
	"github.com/gofiber/fiber/v2"
)

// Sentry implements SentryInterface and ensures that a Sentry hub
// is always available. The hub is used to capture events and errors, and to
// manage scopes (contexts, users, etc.). A hub in Sentry acts as a container
// for scope information (context) for each group of logs (such as a thread
// or an HTTP request).
type Sentry struct {
	hub *sentry.Hub
}

// NewSentry creates a new instance of Sentry with a specified
// Sentry Client Options. So a Sentry can contain only on sentry Client.
// This function forward the sentry error from the client creation.
func NewSentry(options sentry.ClientOptions) (*Sentry, error) {
	client, err := sentry.NewClient(options)
	if err != nil {
		return nil, err
	}
	return &Sentry{hub: sentry.NewHub(client, sentry.NewScope())}, nil
}

// Flush waits until the underlying Client Transport sends any buffered events to the Sentry server
// It must be use just one time in the carrying goroutine
func (l *Sentry) Flush(timeout time.Duration) bool {
	return l.hub.Client().Flush(timeout)
}

func (l *Sentry) GetClient() *sentry.Client {
	return l.hub.Client()
}

// applyOptions applies various options (contexts, tags, user) to the
// current scope of the Sentry hub.
func (l *Sentry) applyOptions(opts ...Option) {
	o := Options{}
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

func (l *Sentry) Info(msg string, opts ...Option) {
	l.applyOptions(opts...)
	l.hub.Scope().SetLevel(sentry.LevelInfo)
	l.hub.CaptureMessage(msg)
}

func (l *Sentry) Error(err error, opts ...Option) {
	l.applyOptions(opts...)
	l.hub.Scope().SetLevel(sentry.LevelError)
	l.hub.CaptureException(err)
}

func (l *Sentry) Fatal(err error, opts ...Option) {
	l.applyOptions(opts...)
	l.hub.Scope().SetLevel(sentry.LevelFatal)
	l.hub.CaptureException(err)
}

// clone creates a new instance of Sentry with a new hub (or clones the current one)
// and applies the provided options. This ensures that each cloned logger can have its own
// independent scope, without interfering with the parent logger's hub.
func (l *Sentry) clone(hub *sentry.Hub, opts ...Option) *Sentry {
	if hub == nil {
		hub = l.hub.Clone()
	}
	clone := &Sentry{
		hub: hub,
	}
	clone.applyOptions(opts...)
	return clone
}

// With create a new Sentry containing cloned hub from the parent with the associated options
func (l *Sentry) With(opts ...Option) LoggerInterface {
	return l.clone(nil, opts...)
}

// WithHub acts like the With method, but you can specify your hub.
func (l *Sentry) WithHub(hub *sentry.Hub, opts ...Option) SentryLoggerInterface {
	return l.clone(hub, opts...)
}

// WithFiberCtx acts like the With method, but you can specify a fiber Ctx to retreive hub
// and some contextual request datas, see github.com/getsentry/sentry-go/fiber.
func (l *Sentry) WithFiberCtx(ctx *fiber.Ctx, opts ...Option) FiberLoggerInterface {
	return l.clone(sentryfiber.GetHubFromContext(ctx), opts...)
}

// AddBreadcrumb adds a breadcrumb to the current Sentry scope.
func (l *Sentry) AddBreadcrumb(breadcrumb *sentry.Breadcrumb, hint *sentry.BreadcrumbHint) {
	l.hub.AddBreadcrumb(breadcrumb, hint)
}

// ClearBreadcrumbs removes all breadcrumbs from the current scope.
func (l *Sentry) ClearBreadcrumbs() {
	l.hub.Scope().ClearBreadcrumbs()
}
