package worker

type Consumer interface {
	Server() error // start server
	Stop()  // stop server
	StopChan() chan int
	Handle(h func(j *Job)) // handle job
}
