package worker

import "time"

type Producer interface {
	Publish(jobs ...*Job) error
	DeferredPublish( t time.Duration, jobs ...*Job) error
}
