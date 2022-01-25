package rabbit

import (
	"encoding/json"
	"fmt"
	"github.com/AndreyAndreevich/examples-go/rabbitmq/domain"
	"github.com/streadway/amqp"
)

type Handler struct{}

func (h *Handler) Handle(msg amqp.Delivery) {
	fmt.Println(msg.RoutingKey)
	var event domain.DeleteEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		fmt.Println(err)

		if err := msg.Reject(false); err != nil {
			panic(err)
		}
		return
	}

	fmt.Println(event)

	if err := msg.Ack(true); err != nil {
		panic(err)
	}
}
