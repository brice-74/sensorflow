package config

import "github.com/brice-74/sensorflow/pkg/config"

type Sentry struct {
	DSN              string
	SampleRate       float64
	TracesSampleRate float64
}

func (c *Sentry) Define(loader *config.Loader) {
	loader.String(&c.DSN, "sentry_dsn", "SENTRY_DSN", "", "sentry DSN").Required()
	loader.Float64(&c.SampleRate, "sentry_sample_rate", "SENTRY_SAMPLE_RATE", 0, "sentry sample rate")
	loader.Float64(&c.TracesSampleRate, "sentry_traces_sample_rate", "SENTRY_TRACES_SAMPLE_RATE", 0, "sentry traces sample rate")
}
