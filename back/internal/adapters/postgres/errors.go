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
		return errors.NewError(errors.ErrNotFound, err, nil)
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
		return errors.NewError(errors.ErrNotFound, nil, nil)
	}

	if expected > 0 && int(rows) != expected {
		return errors.NewError(errors.ErrUnexpectedRows,
			fmt.Errorf("expected %d rows affected, got %d", expected, rows),
			map[string]any{"expected": expected, "got": rows})
	}

	return nil
}

// mapDriverError maps common driver errors to Error codes.
func mapDriverError(err error) error {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return errors.NewError(errors.ErrTimeout, err, nil)
	case errors.Is(err, context.Canceled):
		return errors.NewError(errors.ErrCanceled, err, nil)
	}

	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch pqErr.Code {
		case "23505":
			return errors.NewError(errors.ErrAlreadyExists, err, nil)
		case "23503":
			return errors.NewError(errors.ErrInvalidReference, err, nil)
		case "23502", "23514":
			return errors.NewError(errors.ErrInvalidInput, err, nil)
		}
	}

	return err
}
