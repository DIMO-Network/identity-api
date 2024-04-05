package kafka

import (
	"context"

	"github.com/Shopify/sarama"
	"github.com/goccy/go-json"
	"github.com/rs/zerolog"
)

type Config struct {
	Brokers []string
	Topic   string
	Group   string
}

type wrap[A any] struct {
	handler func(context.Context, A) error
	logger  *zerolog.Logger
}

func (w *wrap[A]) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (w *wrap[A]) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (w *wrap[A]) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case msg := <-claim.Messages():
			var a A
			if err := json.Unmarshal(msg.Value, &a); err != nil {
				w.logger.Err(err).Msg("Failed unmarshaling message.")
			} else if err := w.handler(session.Context(), a); err != nil {
				w.logger.Err(err).Msg("Error processing message.")
			}
			session.MarkMessage(msg, "")
		case <-session.Context().Done():
			return nil
		}
	}
}

func Consume[A any](ctx context.Context, config Config, handler func(context.Context, A) error, logger *zerolog.Logger) error {
	g, err := sarama.NewConsumerGroup(config.Brokers, config.Group, nil)
	if err != nil {
		return err
	}

	w := wrap[A]{handler: handler, logger: logger}

	go func() {
		for {
			if err := g.Consume(ctx, []string{config.Topic}, &w); err != nil {
				logger.Err(err).Msg("Consumer group session ended with an error.")
			}
			if ctx.Err() != nil {
				logger.Info().Msg("Context canceled, shutting down.")
				if err := g.Close(); err != nil {
					logger.Err(err).Msg("Error closing consumer group.")
				}
				return
			}
		}
	}()

	return nil
}
