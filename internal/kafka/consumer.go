package kafka

import (
	"crypto/tls"

	kafkago "github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
	"go.uber.org/zap"
)

func NewReader(cfg *KafkaConfig, log *zap.Logger) *kafkago.Reader {
	if len(cfg.Brokers) == 0 {
		log.Info("kafka brokers not configured, consumer disabled")
		return nil
	}

	dialer := &kafkago.Dialer{}
	if cfg.UseTLS {
		mechanism, err := scram.Mechanism(scram.SHA256, cfg.Username, cfg.Password)
		if err != nil {
			log.Error("failed to create SASL mechanism for consumer", zap.Error(err))
			return nil
		}
		dialer.TLS = &tls.Config{}
		dialer.SASLMechanism = mechanism
	}

	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers: cfg.Brokers,
		Topic:   cfg.Topic,
		GroupID: cfg.GroupID,
		Dialer:  dialer,
	})

	log.Info("kafka consumer initialized", zap.Strings("brokers", cfg.Brokers))
	return reader
}
