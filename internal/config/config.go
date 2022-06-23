package config

import "github.com/go-ozzo/ozzo-validation"

type Config struct {
	MaxWorkers         int
	WorkerQueueSize    int
	WaitQueueSize      int
	ReaderBufferSize   int
	Debug              bool
	DatabaseConnection string
}

func (c Config) Validate() error {
	return validation.ValidateStruct(&c,

		validation.Field(&c.MaxWorkers, validation.Required, validation.Min(1)),
		validation.Field(&c.WorkerQueueSize, validation.Required, validation.Min(1)),
		validation.Field(&c.WaitQueueSize, validation.Required, validation.Min(1)),
		validation.Field(&c.ReaderBufferSize, validation.Required, validation.Min(1)),
		validation.Field(&c.DatabaseConnection, validation.Required, validation.Required),
	)
}
