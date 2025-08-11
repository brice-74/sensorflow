package kit

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/brice-74/sensorflow/internal/config"
	"github.com/brice-74/sensorflow/pkg/log"
	"github.com/getsentry/sentry-go"
	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

func Zlog(fileLvl, consoleLvl zerolog.Level) log.FiberLoggerInterface {
	return log.NewZerolog(zerolog.New(
		zerolog.MultiLevelWriter([]io.Writer{
			// console writer
			&zerolog.FilteredLevelWriter{
				Writer: zerolog.LevelWriterAdapter{
					Writer: zerolog.ConsoleWriter{
						Out:        os.Stdout,
						TimeFormat: "2006-01-02 15:04:05",
					},
				},
				Level: consoleLvl,
			},
			// file writer
			&zerolog.FilteredLevelWriter{
				Writer: zerolog.LevelWriterAdapter{
					Writer: &lumberjack.Logger{
						Filename:   "logs/app.log",
						MaxSize:    20,
						MaxBackups: 5,
						MaxAge:     30,
					},
				},
				Level: fileLvl,
			},
		}...),
	).With().Timestamp().Logger())
}

func Sentry(conf config.Sentry) (log.SentryLoggerInterface, func() error, error) {
	sentrylog, err := log.NewSentry(sentry.ClientOptions{
		Dsn:              conf.DSN,
		Debug:            false,
		AttachStacktrace: true,
		EnableTracing:    true,
		SampleRate:       conf.SampleRate,
		TracesSampleRate: conf.TracesSampleRate,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("sentry logger init failed, err: %s", err)
	}

	flush := func() error {
		if !sentrylog.Flush(time.Second * 15) {
			return errors.New("sentry flush timeout! some logs are lost")
		}

		return nil
	}

	return sentrylog, flush, nil
}
