package ports

type Task interface {
	Do()
}

type AsyncSubmitter[T Task] interface {
	Submit(T) error
}
