package ports

import "github.com/google/uuid"

type IngestorConsumer[T any, Enum comparable] interface {
	ID() uuid.UUID
	State() Enum
	Submit(rows ...T)
}

type Ingestor[T any, Enum comparable] interface {
	IngestorConsumer[T, Enum]
	Start()
	Stop()
}
