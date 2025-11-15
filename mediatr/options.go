package mediatr

import "github.com/oesand/octo/backoff"

type Option func(manager *Manager)

// WithBackOff enables backoff support for Send and Publish (also UnmarshallAndPublish) methods
func WithBackOff(options ...backoff.BackOffOption) Option {
	return func(manager *Manager) {
		manager.useBackOff.Store(&backoffConf{
			options: options,
		})
	}
}

type backoffConf struct {
	options []backoff.BackOffOption
}
