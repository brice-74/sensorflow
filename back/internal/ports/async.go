package ports

type Task = func()

type AsyncSubmitter interface {
	Submit(Task) error
}
