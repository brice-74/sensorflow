package log

import (
	"fmt"
	"runtime/debug"
)

func HandlePanic(logger Logger, a any) {
	if sentryLogger, ok := logger.(Sentry); ok {
		sentryLogger.ReportPanic(a)
		return
	}

	logger.Fatal(fmt.Errorf("panic recovered: %v", a),
		Contexts{"panic": {"stacktrace": string(debug.Stack())}})
}

func PanicHandler(logger Logger) func(a any) {
	return func(a any) {
		HandlePanic(logger, a)
	}
}
