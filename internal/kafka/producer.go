package kafka

import (
	"crypto/tls"

	kafkago "github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
	"go.uber.org/zap"
)

func NewProducer(cfg *KafkaConfig, log *zap.Logger) *kafkago.Writer {
	if len(cfg.Brokers) == 0 {
		log.Info("kafka brokers not configured, producer disabled")
		return nil
	}

	transport := &kafkago.Transport{}

	if cfg.UseTLS {
		mechanism, err := scram.Mechanism(scram.SHA256, cfg.Username, cfg.Password)
		if err != nil {
			log.Error("failed to create SASL mechanism", zap.Error(err))
			return nil
		}
		transport.TLS = &tls.Config{}
		transport.SASL = mechanism
	}

	writer := &kafkago.Writer{
		Addr:         kafkago.TCP(cfg.Brokers...),
		Topic:        cfg.Topic,
		Balancer:     &kafkago.LeastBytes{},
		Transport:    transport,
		RequiredAcks: kafkago.RequireAll,
	}

	log.Info("kafka producer initialized", zap.Strings("brokers", cfg.Brokers))
	return writer
}
