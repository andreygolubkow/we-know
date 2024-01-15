package main

import (
	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
	"log"
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

var enqueuer = work.NewEnqueuer("application_namespace", redisPool)

func main() {
	// Enqueue a job named "send_email" with the specified parameters.
	_, err := enqueuer.Enqueue("send_welcome_email", work.Q{"email_address": "test@example.com", "user_id": 4})
	if err != nil {
		log.Fatal(err)
	}
}
