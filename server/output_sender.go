package server

// OutputSender streams Output from server to client
type OutputSender interface {
	Send(*Output) error
}

// OutputReceiver receives a stream of Output from a server
type OutputReceiver interface {
	Recv() (*Output, error)
}
