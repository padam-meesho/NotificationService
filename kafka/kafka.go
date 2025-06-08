package kafka

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/padam-meesho/NotificationService/config"
	"github.com/padam-meesho/NotificationService/internal/models"
	"github.com/padam-meesho/NotificationService/internal/repo"
	"github.com/padam-meesho/NotificationService/internal/utils"
)

type KafkaDao interface {
	Produce(payload models.KafkaPayload) error
	Consume()
}

type KafkaDaoImpl struct {
	producer *kafka.Producer
	consumer *kafka.Consumer
}

var (
	kafkaInstance    *KafkaDaoImpl
	kafkaDaoOnce     sync.Once
	KAFKA_TOPIC_NAME string = "notification.send_sms"
)

func NewKafkaDao(appConfig *models.AppConfig) *KafkaDaoImpl {
	logger := utils.ComponentLogger("kafka")

	kafkaDaoOnce.Do(func() {
		kafkaConfig := config.GetKafkaClient()
		if kafkaConfig == nil {
			logger.Fatal().Msg("Kafka client not initialized during DAO creation")
		}
		kafkaInstance = &KafkaDaoImpl{
			producer: kafkaConfig.KafkaProducer,
			consumer: kafkaConfig.KafkaConsumer,
		}
		logger.Info().Msg("Kafka DAO initialized successfully")
	})
	return kafkaInstance
}

func GetKafkaDao() *KafkaDaoImpl {
	return kafkaInstance
}

func (p *KafkaDaoImpl) Produce(payload models.KafkaPayload) error {
	logger := utils.KafkaLogger("produce", KAFKA_TOPIC_NAME)

	marshalledPayload, err := json.Marshal(&payload)
	if err != nil {
		logger.Error().
			Err(err).
			Str("payload_type", payload.Type).
			Msg("Failed to marshal Kafka payload")
		return err
	}

	logger.Info().
		Str("payload_type", payload.Type).
		Int("payload_size", len(marshalledPayload)).
		Msg("Attempting to produce message")

	err = p.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &KAFKA_TOPIC_NAME,
			Partition: kafka.PartitionAny,
		},
		Value: marshalledPayload,
	}, nil)

	if err != nil {
		logger.Error().
			Err(err).
			Str("payload_type", payload.Type).
			Msg("Failed to produce message to Kafka")
		return err
	}

	logger.Info().
		Str("payload_type", payload.Type).
		Msg("Message successfully sent to Kafka")
	return nil
}

func (c *KafkaDaoImpl) Consume() {
	logger := utils.KafkaLogger("consume", KAFKA_TOPIC_NAME)
	defer c.consumer.Close()

	c.consumer.SubscribeTopics([]string{KAFKA_TOPIC_NAME}, nil)
	serviceInstance := repo.GetNotificationServiceInstance()

	logger.Info().Msg("Kafka consumer started and subscribed to topic")

	for {
		msg, err := c.consumer.ReadMessage(-1)
		if err != nil {
			logger.Error().
				Err(err).
				Msg("Error reading message from Kafka")
			continue
		}

		logger.Info().
			Int32("partition", msg.TopicPartition.Partition).
			Int64("offset", int64(msg.TopicPartition.Offset)).
			Int("message_size", len(msg.Value)).
			Msg("Received message from Kafka")

		var payload models.KafkaPayload
		err = json.Unmarshal(msg.Value, &payload)
		if err != nil {
			logger.Error().
				Err(err).
				Str("raw_message", string(msg.Value)).
				Msg("Failed to unmarshal Kafka payload")
			continue
		}

		logger.Info().
			Str("message_type", payload.Type).
			Msg("Successfully parsed Kafka message")

		switch payload.Type {
		case "SMS_REQUEST":
			var sendSMSPayload models.SendSmsPayload
			err := json.Unmarshal(payload.Data, &sendSMSPayload)
			if err != nil {
				logger.Error().
					Err(err).
					Str("raw_data", string(payload.Data)).
					Msg("Failed to unmarshal SMS payload")
				continue
			}

			logger.Info().
				Str("message_id", sendSMSPayload.MessageId).
				Msg("Processing SMS request from Kafka")

			ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
			err = serviceInstance.HandleKafkaMessages(ctx, sendSMSPayload.MessageId)
			if err != nil {
				logger.Error().
					Err(err).
					Str("message_id", sendSMSPayload.MessageId).
					Msg("Failed to process SMS request")
			} else {
				logger.Info().
					Str("message_id", sendSMSPayload.MessageId).
					Msg("Successfully processed SMS request")
			}
		default:
			logger.Warn().
				Str("message_type", payload.Type).
				Msg("Received unknown message type")
		}
	}
}
