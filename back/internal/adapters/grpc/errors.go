package grpc

import (
	"github.com/brice-74/sensorflow/pkg/errors"
	"github.com/brice-74/sensorflow/pkg/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ErrorHandler struct {
	log log.LoggerInterface
}

func NewErrorHandler(logger log.LoggerInterface) *ErrorHandler {
	return &ErrorHandler{log: logger}
}

func (h *ErrorHandler) Handle(err error, logCtx log.Contexts) error {
	if err == nil {
		return nil
	}

	var e *errors.Error
	if errors.As(err, &e) {
		switch e.Code {
		case errors.ErrNotFound:
			return status.Error(codes.NotFound, "resource not found")
		case errors.ErrAlreadyExists:
			return status.Error(codes.AlreadyExists, "resource already exists")
		case errors.ErrInvalidInput:
			return status.Error(codes.InvalidArgument, "invalid input")
		case errors.ErrInvalidReference:
			return status.Error(codes.FailedPrecondition, "invalid reference")
		case errors.ErrTimeout:
			h.log.Error(err, logCtx)
			return status.Error(codes.DeadlineExceeded, "request timeout")
		case errors.ErrCanceled:
			h.log.Error(err, logCtx)
			return status.Error(codes.Canceled, "request canceled")
		}
	}

	h.log.Error(err, logCtx)
	return status.Error(codes.Internal, "internal error")
}
