package kafka

type KafkaConfig struct {
	Brokers  []string
	Topic    string
	GroupID  string
	UseTLS   bool
	Username string
	Password string
}
