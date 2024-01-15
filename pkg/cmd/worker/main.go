package main

import (
	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
	"os"
	"os/signal"
	"we-know/pkg/worker/jobs"
)

// Make a redis pool
var redisPool = &redis.Pool{
	MaxActive: 5,
	MaxIdle:   5,
	Wait:      true,
	Dial: func() (redis.Conn, error) {
		return redis.Dial("tcp", ":6379")
	},
}

func main() {
	// Make a new pool. Arguments:
	// Context{} is a struct that will be the context for the request.
	// 10 is the max concurrency
	// "application_namespace" is the Redis namespace
	// redisPool is a Redis pool
	pool := work.NewWorkerPool(jobs.Context{}, 10, "application_namespace", redisPool)

	// Add middleware that will be executed for each job
	pool.Middleware((*jobs.Context).Log)
	pool.Middleware((*jobs.Context).FindCustomer)

	// Map the name of jobs to handlers functions
	pool.Job("send_welcome_email", (*jobs.Context).SendWelcomeEmail)

	// Start processing jobs
	pool.Start()

	// Wait for a signal to quit:
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, os.Kill)
	<-signalChan

	// Stop the pool
	pool.Stop()
}
