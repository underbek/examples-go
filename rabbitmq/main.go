package main

import (
	"github.com/AndreyAndreevich/examples-go/rabbitmq/rabbit"
	"log"
)

func main() {
	consumer := rabbit.New(
		"amqp://guest:guest@localhost:5672/",
		"tasks",
		"delete",
		"delete",
		1,
	)

	handler := rabbit.Handler{}

	err := consumer.Handle(handler.Handle)
	log.Fatal(err)
}
