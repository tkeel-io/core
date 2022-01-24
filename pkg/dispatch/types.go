package dispatch

type Dispatcher interface {
	Run() error
	Stop() error
}
