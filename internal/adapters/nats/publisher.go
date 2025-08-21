package nats

import (
	"context"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"

	"service-mqtt/internal/config"
	"service-mqtt/internal/core/port"
)

var _ port.Publisher = (*Publisher)(nil)

type Publisher struct {
	nc      *nats.Conn
	js      nats.JetStreamContext
	timeout time.Duration
	log     zerolog.Logger
}

func New(cfg config.Config, logger zerolog.Logger) (*Publisher, error) {
	nc, err := nats.Connect(cfg.NATSURL,
		nats.Name("service-mqtt-publisher"),
	)
	if err != nil {
		return nil, err
	}

	var js nats.JetStreamContext
	if cfg.EnableJetStream {
		if js, err = nc.JetStream(); err != nil {
			_ = nc.Drain()
			return nil, err
		}
	}

	return &Publisher{
		nc:      nc,
		js:      js,
		timeout: cfg.PublishTimeout,
		log:     logger.With().Str("adapter", "nats").Logger(),
	}, nil
}

func (p *Publisher) Publish(ctx context.Context, subject string, payload []byte) error {
	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	if p.js == nil {
		if err := p.nc.Publish(subject, payload); err != nil {
			return err
		}
		return p.nc.FlushWithContext(ctx)
	}

	_, err := p.js.PublishMsg(&nats.Msg{
		Subject: subject,
		Header:  nats.Header{"Content-Type": []string{"application/json"}},
		Data:    payload,
	}, nats.Context(ctx))
	return err
}

func (p *Publisher) Close() error {
	if err := p.nc.Drain(); err != nil {
		return err
	}
	p.nc.Close()
	return nil
}
