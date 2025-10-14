package log

import "maps"

type Options struct {
	Contexts Contexts
	Tags     Tags
	User     *User
}

type User struct {
	ID       string            `json:"id"`
	Email    string            `json:"email"`
	Username string            `json:"username"`
	Data     map[string]string `json:"data"`
}

func (u *User) Apply(opts *Options) {
	opts.User = u
}

type Contexts map[string]map[string]any

func (c Contexts) Apply(opts *Options) {
	if opts.Contexts == nil {
		opts.Contexts = c
	} else {
		maps.Copy(opts.Contexts, c)
	}
}

type Tags map[string]string

func (t Tags) Apply(opts *Options) {
	opts.Tags = t

	if opts.Tags == nil {
		opts.Tags = t
	} else {
		maps.Copy(opts.Tags, t)
	}
}
