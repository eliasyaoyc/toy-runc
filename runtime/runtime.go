package runtime

var (
	// ReallyCrash controls the behavior of HandleCrash and now defaults
	// true. It's still exposed so components can optionally set to false
	// to restore prior behavior.
	ReallyCrash = true

	// PanicHandlers is a list of functions which will be invoked when a panic happens.
	PanicHandlers = []func(interface{}){logPanic}
)

func logPanic(r interface{}) {
}
