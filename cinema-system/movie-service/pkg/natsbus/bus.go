package natsbus

import (
	"encoding/json"

	"github.com/nats-io/nats.go"
)

type Publisher struct {
	conn *nats.Conn
}

func NewPublisher(url string) (*Publisher, error) {
	conn, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}
	return &Publisher{conn: conn}, nil
}

func (p *Publisher) Close() {
	if p != nil && p.conn != nil {
		p.conn.Close()
	}
}

func (p *Publisher) Publish(subject string, payload any) error {
	if p == nil || p.conn == nil {
		return nil
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return p.conn.Publish(subject, data)
}
