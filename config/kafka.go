package config

import (
	"sync"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/padam-meesho/NotificationService/internal/models"
	"github.com/padam-meesho/NotificationService/internal/utils"
)

type KafkaClientImpl struct {
	KafkaProducer *kafka.Producer
	KafkaConsumer *kafka.Consumer
}

var (
	kafkaOnce     sync.Once
	kafkaInstance *KafkaClientImpl
)

func NewKafkaClient(appConfig *models.AppConfig) *KafkaClientImpl {
	logger := utils.ComponentLogger("kafka")

	kafkaOnce.Do(func() {
		logger.Info().
			Str("bootstrap_servers", appConfig.Kafka.BootStrapServers).
			Str("group_id", appConfig.Kafka.GroupId).
			Msg("Initializing Kafka client")

		kafkaInstance = &KafkaClientImpl{
			KafkaProducer: InitKafkaProducer(appConfig),
			KafkaConsumer: InitKafkaConsumer(appConfig),
		}

		if kafkaInstance.KafkaProducer == nil || kafkaInstance.KafkaConsumer == nil {
			logger.Fatal().Msg("Failed to initialize Kafka producer or consumer")
		}

		logger.Info().Msg("Successfully initialized Kafka client")
	})
	return kafkaInstance
}

func GetKafkaClient() *KafkaClientImpl {
	return kafkaInstance
}

func InitKafkaProducer(appConfig *models.AppConfig) *kafka.Producer {
	logger := utils.ComponentLogger("kafka")

	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": appConfig.Kafka.BootStrapServers,
	})
	if err != nil {
		logger.Error().
			Err(err).
			Str("bootstrap_servers", appConfig.Kafka.BootStrapServers).
			Msg("Failed to create Kafka producer")
		return nil
	}

	logger.Info().
		Str("bootstrap_servers", appConfig.Kafka.BootStrapServers).
		Msg("Successfully created Kafka producer")
	return p
}

func InitKafkaConsumer(appConfig *models.AppConfig) *kafka.Consumer {
	logger := utils.ComponentLogger("kafka")

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": appConfig.Kafka.BootStrapServers,
		"group.id":          appConfig.Kafka.GroupId,
		"auto.offset.reset": appConfig.Kafka.AutoOffsetReset,
	})
	if err != nil {
		logger.Error().
			Err(err).
			Str("bootstrap_servers", appConfig.Kafka.BootStrapServers).
			Str("group_id", appConfig.Kafka.GroupId).
			Msg("Failed to create Kafka consumer")
		return nil
	}

	logger.Info().
		Str("bootstrap_servers", appConfig.Kafka.BootStrapServers).
		Str("group_id", appConfig.Kafka.GroupId).
		Msg("Successfully created Kafka consumer")
	return c
}
