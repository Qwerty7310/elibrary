package service

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

type PrintTask struct {
	Str1    string `json:"str1"`
	Str2    string `json:"str2"`
	Barcode string `json:"barcode"`
}

type PrintQueue struct {
	url   string
	queue string
}

func NewPrintQueue(url, queue string) *PrintQueue {
	return &PrintQueue{url: url, queue: queue}
}

func (p *PrintQueue) Send(ctx context.Context, task PrintTask) error {
	conn, err := amqp.Dial(p.url)
	if err != nil {
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		p.queue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	body, err := json.Marshal(task)
	if err != nil {
		return err
	}

	return ch.PublishWithContext(
		ctx,
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}
