package app

import (
	"github.com/padam-meesho/NotificationService/config"
	"github.com/padam-meesho/NotificationService/dao"
	"github.com/padam-meesho/NotificationService/internal/models"
	services "github.com/padam-meesho/NotificationService/internal/repo"
	"github.com/padam-meesho/NotificationService/internal/utils"
	"github.com/padam-meesho/NotificationService/kafka"
)

// this shall be like the constructor function for the NewApp, which shall be called in the main.go.

func NewApp(appConfig models.AppConfig) {
	logger := utils.ComponentLogger("app")

	logger.Info().Msg("Starting application initialization")

	// Initialize ScyllaDB client
	logger.Info().Msg("Initializing ScyllaDB")
	config.InitScyllaSession(&appConfig)

	// Initialize Redis client
	logger.Info().Msg("Initializing Redis")
	config.NewRedisCache(&appConfig)

	// Initialize Kafka client
	logger.Info().Msg("Initializing Kafka client")
	kafkaClient := config.NewKafkaClient(&appConfig)
	if kafkaClient == nil {
		logger.Fatal().Msg("Failed to initialize Kafka client")
	}

	// Initialize Kafka DAO
	logger.Info().Msg("Initializing Kafka DAO")
	kafkaDao := kafka.NewKafkaDao(&appConfig)

	// Start Kafka consumer in a separate goroutine
	go func() {
		logger.Info().Msg("Starting Kafka consumer in background")
		kafkaDao.Consume()
	}()

	// Initialize service instance with DAOs
	logger.Info().Msg("Initializing notification service")
	services.InitNotificationService(
		*dao.NewRedisDao(),
		*dao.NewScyllaSessionDao(),
	)

	logger.Info().Msg("Application initialization completed successfully")
}
