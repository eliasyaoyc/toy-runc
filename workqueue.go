package go_workqueue

type Plugin interface {
}

type Queue interface {
	Add(elem interface{})
	Get() (elem interface{}, shutdown bool)
	Done(elem interface{})
	Len() int
	ShutDown()
	ShutDownWithDrain()
	ShuttingDown() bool
}

func New() *WorkQueue {
	return NewWithName("")
}

func NewWithName(name string) *WorkQueue {
	return newWorkQueue()
}

type element interface{}

type WorkQueue struct {
	queue []element
}

func newWorkQueue() *WorkQueue {

}
