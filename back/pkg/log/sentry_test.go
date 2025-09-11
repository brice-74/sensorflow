//go:build unit

package log_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/brice-74/sensorflow/pkg/log"
	"github.com/getsentry/sentry-go"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

type mockTransport struct {
	events []*sentry.Event
}

func (m *mockTransport) Close()                                  {}
func (m *mockTransport) Configure(_ sentry.ClientOptions)        {}
func (m *mockTransport) Flush(_ time.Duration) bool              { return true }
func (m *mockTransport) FlushWithContext(_ context.Context) bool { return true }
func (m *mockTransport) SendEvent(event *sentry.Event) {
	m.events = append(m.events, event)
}
func (m *mockTransport) Reset() {
	m.events = m.events[:0]
}

func TestSentryLogger(t *testing.T) {
	transport := &mockTransport{}
	logger, err := log.NewSentry(sentry.ClientOptions{Transport: transport})
	require.NoErrorf(t, err, "create sentry client error: %s", err)

	t.Run("Info", func(t *testing.T) {
		logger.Info("info message")

		require.Len(t, transport.events, 1, "event not recorded")
		require.Equal(t, "info message", transport.events[0].Message)
		require.Equal(t, sentry.LevelInfo, transport.events[0].Level)
		transport.Reset()
	})

	t.Run("Error", func(t *testing.T) {
		err := errors.New("error message")
		logger.Error(err)

		require.Len(t, transport.events, 1, "event not recorded")
		require.Len(t, transport.events[0].Exception, 1, "exception not recorded")
		require.Equal(t, err.Error(), transport.events[0].Exception[0].Value)
		require.Equal(t, sentry.LevelError, transport.events[0].Level)
		transport.Reset()
	})

	t.Run("Fatal", func(t *testing.T) {
		err := errors.New("error fatal")
		logger.Fatal(err)

		require.Len(t, transport.events, 1, "event not recorded")
		require.Len(t, transport.events[0].Exception, 1, "exception not recorded")
		require.Equal(t, err.Error(), transport.events[0].Exception[0].Value)
		require.Equal(t, sentry.LevelFatal, transport.events[0].Level)
		transport.Reset()
	})

	t.Run("With", func(t *testing.T) {
		sub := logger.With(
			&log.User{
				ID:       "u123",
				Email:    "user@example.com",
				Username: "testuser",
			},
			log.Tags{
				"env": "production",
			},
			log.Contexts{
				"request": {"id": "123"},
			},
		)
		sub.Info("message with options")

		require.Len(t, transport.events, 1, "event not recorded")
		require.Equal(t, "message with options", transport.events[0].Message)
		require.Equal(t, sentry.LevelInfo, transport.events[0].Level)
		require.Equal(t, map[string]string{"env": "production"}, transport.events[0].Tags)
		require.Equal(t, sentry.User{ID: "u123", Email: "user@example.com", Username: "testuser"}, transport.events[0].User)
		require.Equal(t, map[string]interface{}{"id": "123"}, transport.events[0].Contexts["request"])
		transport.Reset()
	})

	t.Run("WithOptionMergeCaptureOptions", func(t *testing.T) {
		sub := logger.With(
			log.Tags{
				"env": "production",
			},
		)
		sub.Info("sublogger message",
			log.Tags{
				"other": "tag",
			},
			log.Contexts{
				"request": {"id": "123"},
			},
		)

		require.Len(t, transport.events, 1, "event not recorded")
		require.Equal(t, "sublogger message", transport.events[0].Message)
		require.Equal(t, sentry.LevelInfo, transport.events[0].Level)
		require.Equal(t, map[string]string{"env": "production", "other": "tag"}, transport.events[0].Tags)
		require.Equal(t, map[string]interface{}{"id": "123"}, transport.events[0].Contexts["request"])
		transport.Reset()
	})

	t.Run("WithOptionInheritance", func(t *testing.T) {
		sub := logger.With(
			log.Tags{
				"parent": "tag",
			},
		)

		sub2 := sub.With(
			log.Tags{
				"child": "tag",
			},
		)

		sub2.Info("")

		require.Len(t, transport.events, 1, "event not recorded")
		require.Equal(t, map[string]string{"parent": "tag", "child": "tag"}, transport.events[0].Tags)
		transport.Reset()
	})

	t.Run("WithHub", func(t *testing.T) {
		hub := sentry.NewHub(logger.GetClient(), sentry.NewScope())
		hub.Scope().SetTag("from", "hub")

		sub := logger.WithHub(
			hub,
			log.Tags{
				"tag1": "tag",
			},
		)
		sub.Info("")

		require.Len(t, transport.events, 1, "event not recorded")
		require.Equal(t, map[string]string{"from": "hub", "tag1": "tag"}, transport.events[0].Tags)
		transport.Reset()
	})

	t.Run("WithFiberCtx", func(t *testing.T) {
		hub := sentry.NewHub(logger.GetClient(), sentry.NewScope())
		hub.Scope().SetTag("from", "hub")

		app := fiber.New()
		ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
		defer app.ReleaseCtx(ctx)
		ctx.Locals("sentry", hub)

		sub := logger.WithFiberCtx(
			ctx,
			log.Tags{
				"tag1": "tag",
			},
		)
		sub.Info("")

		require.Len(t, transport.events, 1, "event not recorded")
		require.Equal(t, map[string]string{"from": "hub", "tag1": "tag"}, transport.events[0].Tags)
		transport.Reset()
	})

	t.Run("Add&ClearBreadcrumbs", func(t *testing.T) {
		logger.AddBreadcrumb(&sentry.Breadcrumb{
			Message: "breadcrumb1",
		}, nil)

		logger.AddBreadcrumb(&sentry.Breadcrumb{
			Message: "breadcrumb2",
		}, nil)

		logger.Info("")

		if len(transport.events) == 0 {
			t.Fatalf("event not recorded")
		}

		require.Len(t, transport.events, 1, "event not recorded")
		require.Len(t, transport.events[0].Breadcrumbs, 2, "breadcrumb not recorded")
		require.Equal(t, "breadcrumb1", transport.events[0].Breadcrumbs[0].Message)
		require.Equal(t, "breadcrumb2", transport.events[0].Breadcrumbs[1].Message)
		transport.Reset()

		logger.ClearBreadcrumbs()
		logger.Info("")

		require.Len(t, transport.events, 1, "event not recorded")
		require.Equal(t, 0, len(transport.events[0].Breadcrumbs))
		transport.Reset()
	})

	t.Run("BreadcrumbsInheritance", func(t *testing.T) {
		logger.AddBreadcrumb(&sentry.Breadcrumb{
			Message: "breadcrumb1",
		}, nil)

		sub := logger.With()
		sub.Info("")

		require.Len(t, transport.events, 1, "event not recorded")
		require.Equal(t, 1, len(transport.events[0].Breadcrumbs))

		logger.ClearBreadcrumbs()
		transport.Reset()
	})

	t.Run("ChildDoNotAffectParent", func(t *testing.T) {
		sub := logger.With(
			log.Tags{
				"env": "production",
			},
		).(log.SentryLoggerInterface)

		sub.AddBreadcrumb(&sentry.Breadcrumb{
			Message: "breadcrumb1",
		}, nil)

		logger.Info("parent message")

		require.Len(t, transport.events, 1, "event not recorded")
		require.NotEqual(t, map[string]string{"env": "production"}, transport.events[0].Tags)
		require.Equal(t, 0, len(transport.events[0].Breadcrumbs))
		transport.Reset()
	})

	t.Run("ParentDoNotAffectChild", func(t *testing.T) {
		sub := logger.With().(log.SentryLoggerInterface)
		sub.AddBreadcrumb(&sentry.Breadcrumb{
			Message: "createdBySub1",
		}, nil)

		sub2 := sub.With().(log.SentryLoggerInterface)
		sub2.AddBreadcrumb(&sentry.Breadcrumb{
			Message: "createdBySub2",
		}, nil)

		sub.ClearBreadcrumbs()
		sub2.Info("")

		require.Len(t, transport.events, 1, "event not recorded")
		require.Equal(t, 2, len(transport.events[0].Breadcrumbs))

		transport.Reset()
	})
}
