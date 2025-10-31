package grpc

import (
	"github.com/brice-74/sensorflow/internal/log"
	"github.com/brice-74/sensorflow/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ErrorHandler struct {
	log log.Logger
}

func NewErrorHandler(logger log.Logger) *ErrorHandler {
	return &ErrorHandler{log: logger}
}

func (h *ErrorHandler) Handle(err error, logCtx log.Contexts) error {
	if err == nil {
		return nil
	}

	var e *errors.Error
	if errors.As(err, &e) {
		switch e.Code {
		case errors.CodeNotFound:
			return status.Error(codes.NotFound, "resource not found")
		case errors.CodeAlreadyExists:
			return status.Error(codes.AlreadyExists, "resource already exists")
		case errors.CodeInvalidInput:
			return status.Error(codes.InvalidArgument, "invalid input")
		case errors.CodeInvalidReference:
			return status.Error(codes.FailedPrecondition, "invalid reference")
		case errors.CodeTimeout:
			h.log.Error(err, logCtx)
			return status.Error(codes.DeadlineExceeded, "request timeout")
		case errors.CodeCanceled:
			h.log.Error(err, logCtx)
			return status.Error(codes.Canceled, "request canceled")
		}
	}

	h.log.Error(err, logCtx)
	return status.Error(codes.Internal, "internal error")
}
