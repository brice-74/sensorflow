package redis

import (
	"fmt"

	"github.com/brice-74/sensorflow/pkg/errors"
	"github.com/brice-74/sensorflow/pkg/gobutil"
	"github.com/redis/go-redis/v9"
)

type cmd[T any] interface {
	Result() (T, error)
}

type Cmd[CMD cmd[T], T any] struct {
	Cmd      CMD
	markDown func()
}

func (cmd *Cmd[CMD, T]) Result() (T, error) {
	v, err := cmd.Cmd.Result()
	if err = HandleError(err, cmd.markDown); err != nil {
		var zero T
		return zero, err
	}
	return v, err
}

type MultiCmd[CMD cmd[T], T any] struct {
	Cmds     map[string]CMD
	markDown func()
}

func (cmd *MultiCmd[CMD, T]) Result() (map[string]T, error) {
	results := make(map[string]T, len(cmd.Cmds))
	var errs []error

	for key, c := range cmd.Cmds {
		v, err := c.Result()
		if err = HandleError(err, cmd.markDown); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", key, err))
		}
		results[key] = v
	}

	return results, errors.JoinWrap(errs...)
}

type StatusCmd = Cmd[*redis.StatusCmd, string]

type StringSliceCmd = Cmd[*redis.StringSliceCmd, []string]

type IntCmd = Cmd[*redis.IntCmd, int64]

type BoolCmd = Cmd[*redis.BoolCmd, bool]

type MultiBoolCmd = MultiCmd[*redis.BoolCmd, bool]

type StringCmdGob[T any] struct {
	Cmd      *redis.StringCmd
	markDown func()
}

func (cmd StringCmdGob[T]) Result() (*T, error) {
	str, err := cmd.Cmd.Result()
	if err = HandleError(err, cmd.markDown); err != nil {
		return nil, err
	}
	return gobutil.Decode[T]([]byte(str))
}

type SliceCmdGob[T any] struct {
	Cmd      *redis.SliceCmd
	markDown func()
}

func (cmd SliceCmdGob[T]) Result() ([]*T, error) {
	slice, err := cmd.Cmd.Result()
	if err = HandleError(err, cmd.markDown); err != nil {
		return nil, err
	}
	return gobutil.DecodeManyAnyStr[T](slice)
}
