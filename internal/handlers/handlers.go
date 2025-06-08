package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/padam-meesho/NotificationService/internal/models"
	"github.com/padam-meesho/NotificationService/internal/repo"
	"github.com/padam-meesho/NotificationService/internal/utils"
	"github.com/padam-meesho/NotificationService/kafka"
)

// we generate a traceID for all the different requests we get.
func SendSmsController(c *gin.Context) {
	// now here we have to define the service function that does the working internally.
	// we have to parse the request body into a struct and validate if it it of the correct structure or not.
	// c.ShouldBindJSON(the object of the request body type)
	logger := utils.LogWithContext(c.Request.Context())
	logger.Info().Msgf("SendSmsController called")
	var req models.SendSms
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(400, gin.H{
			"message": "Invalid Request Body",
		})
		return
	}
	serviceInstance := repo.GetNotificationServiceInstance()
	// now since the validation is done, call the service.
	reqId, err := serviceInstance.SendSMSService(c, req)
	if err != nil {
		c.JSON(400, gin.H{"ERROR": err})
		return
	}

	// Create SMS payload
	smsPayload := models.SendSmsPayload{
		MessageId: reqId,
	}

	// Marshal the SMS payload
	smsPayloadBytes, err := json.Marshal(smsPayload)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to marshal SMS payload")
		c.JSON(500, gin.H{"error": "Failed to process request"})
		return
	}

	// Create Kafka payload
	payload := models.KafkaPayload{
		Type: "SMS_REQUEST",
		Data: smsPayloadBytes,
	}

	// Send to Kafka producer
	kafkaInstance := kafka.GetKafkaDao()
	if kafkaInstance == nil {
		logger.Error().Msg("Kafka DAO not initialized")
		c.JSON(500, gin.H{"error": "Failed to process request"})
		return
	}
	err = kafkaInstance.Produce(payload)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to produce Kafka message")
		c.JSON(500, gin.H{"error": "Failed to process request"})
		return
	}

	c.JSON(200, gin.H{"request_id": reqId, "message": "message sent successfully!"})
}

func GetSmsController(c *gin.Context) {
	logger := utils.LogWithContext(c.Request.Context())
	requestID := c.Param("request_id")
	logger.Info().
		Str("request_id", requestID).
		Msg("Processing SMS request")
	request_id := c.Param("request_id") // these are the path variables used in our route path.
	// okay so understand this,
	// gin is different than fiber so we have to
	// get a response and a err from our service
	// and then set the response directly in the
	// context of our gin.Context, which was
	// traditionally different in fiber, where we
	// have the option to create an error object
	// and then return it from the handler.
	serviceInstance := repo.GetNotificationServiceInstance()
	resp, err := serviceInstance.GetSMSService(c, request_id)
	if err != nil {
		c.JSON(400, gin.H{"ERROR": err})
		return
	}
	c.JSON(200, gin.H{"request_id": requestID, "message_details": resp})
}

func GetBlacklistController(c *gin.Context) {
	logger := utils.LogWithContext(c.Request.Context())
	logger.Info().Msgf("")
	serviceInstance := repo.GetNotificationServiceInstance()
	resp, err := serviceInstance.GetBlacklistService(c)
	if err != nil {
		c.JSON(400, gin.H{"ERROR": err})
		return
	}
	c.JSON(200, gin.H{"blacklisted_numbers": resp})
}

func AddToBlacklistController(c *gin.Context) {
	logger := utils.LogWithContext(c.Request.Context())
	logger.Info().Msgf("")
	var req models.AddToBlacklist
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(400, gin.H{
			"message": "Invalid Request Body",
		})
		return
	}
	serviceInstance := repo.GetNotificationServiceInstance()
	// now since the validation is done, call the service.
	err = serviceInstance.AddToBlacklistService(c, req)
	if err != nil {
		c.JSON(400, gin.H{"ERROR": err})
		return
	}
	c.JSON(200, gin.H{"Message": fmt.Sprintf("%s successfully blacklisted", req.PhoneNumbers)})
}

func RemoveFromBlacklistController(c *gin.Context) {
	logger := utils.LogWithContext(c.Request.Context())
	logger.Info().Msgf("")
	// use of path variable here too.
	number := c.Param("number")
	// number, _ = strconv.Atoi(number)
	serviceInstance := repo.GetNotificationServiceInstance()
	resp, err := serviceInstance.RemoveFromBlacklistService(c, number)
	if err != nil {
		c.JSON(400, err)
		return
	}
	if resp {
		c.JSON(200, gin.H{"Message": "number successfully removed!"})
		return
	}
	c.JSON(200, gin.H{"Message": "number not present in blacklist!"})
}
