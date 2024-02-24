package main

import (
	"fmt"
	"sync"

	"github.com/gofiber/fiber/v2"
)

type Message struct {
	Data string `json:"data"`
}

type PubSub struct {
	sub []chan Message
	mu  sync.Mutex
}

func (ps *PubSub) Subscribe() chan Message {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ch := make(chan Message, 1)
	ps.sub = append(ps.sub, ch)

	return ch
}

func (ps *PubSub) Publish(msg *Message) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	for _, sub := range ps.sub {
		sub <- *msg
	}
}

func (ps *PubSub) UnSubScribe(ch chan Message) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	for i, sub := range ps.sub {
		if sub == ch {
			ps.sub = append(ps.sub[:i], ps.sub[i+1:]...)
			close(sub)
			return
		}
	}
}

func main() {
	app := fiber.New()

	pubsub := &PubSub{}

	app.Post("/publisher", func(c *fiber.Ctx) error {
		message := new(Message)
		if err := c.BodyParser(message); err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		pubsub.Publish(message)
		return c.JSON(&fiber.Map{
			"Status":  "OK",
			"Message": "Message published",
		})
	})

	sub := pubsub.Subscribe()
	go func() {
		for msg := range sub {
			fmt.Println("Received message:", msg)
		}
	}()

	app.Listen(":8080")
}
