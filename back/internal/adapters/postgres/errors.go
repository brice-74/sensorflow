package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/brice-74/sensorflow/pkg/errors"
	"github.com/lib/pq"
)

// handleSelectError maps common select errors to your Error codes.
func handleSelectError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return errors.NewWrappedErr(errors.CodeNotFound, err, nil)
	}
	return mapDriverError(err)
}

// expected set less than 0 skips checks related to rows affected
func handleResultError(res sql.Result, err error, expected int) error {
	if err != nil {
		return mapDriverError(err)
	}
	if expected < 0 {
		return nil
	}

	rows, e := res.RowsAffected()
	if e != nil {
		return e
	}

	if rows == 0 && expected > 0 {
		return errors.NewWrappedErr(errors.CodeNotFound, nil, nil)
	}

	if expected > 0 && int(rows) != expected {
		return errors.NewWrappedErr(errors.CodeUnexpectedRows,
			fmt.Errorf("expected %d rows affected, got %d", expected, rows),
			map[string]any{"expected": expected, "got": rows})
	}

	return nil
}

// mapDriverError maps common driver errors to Error codes.
func mapDriverError(err error) error {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return errors.NewWrappedErr(errors.CodeTimeout, err, nil)
	case errors.Is(err, context.Canceled):
		return errors.NewWrappedErr(errors.CodeCanceled, err, nil)
	}

	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch pqErr.Code {
		case "23505":
			return errors.NewWrappedErr(errors.CodeAlreadyExists, err, nil)
		case "23503":
			return errors.NewWrappedErr(errors.CodeInvalidReference, err, nil)
		case "23502", "23514":
			return errors.NewWrappedErr(errors.CodeInvalidInput, err, nil)
		}
	}

	return err
}
