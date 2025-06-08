package repo

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/padam-meesho/NotificationService/dao"
	"github.com/padam-meesho/NotificationService/internal/models"
	"github.com/padam-meesho/NotificationService/internal/utils"
)

// okay we need to have a struct that has the DAOs which we shall need to implement the changes.
// we need to have an interface which has all the methods we need to define in our service.
// Each method shall be defined as a struct method it is a part of.

type NotificationServiceMethods interface {
	InitNotificationService(redisDao dao.RedisDaoImpl, scyllaDao dao.ScyllaDbDaoImpl)
	SendSMSService(ctx context.Context, req models.SendSms) (string, error)
	HandleKafkaMessages(ctx context.Context, requestId string) error
	SendMessage(req *models.SMSRequest)
	GetSMSService(ctx context.Context, reqID string) (any, error)
	GetBlacklistService(ctx context.Context) ([]string, error)
	AddToBlacklistService(ctx context.Context, req models.AddToBlacklist) error
	RemoveFromBlacklistService(ctx context.Context, number string) (bool, error)
}

type NotificationServiceMethodsImpl struct {
	// here we have to have all the DAO clients.
	redisDao  dao.RedisDaoImpl
	scyllaDao dao.ScyllaDbDaoImpl
}

var (
	notificationServiceInstance *NotificationServiceMethodsImpl
)

// InitNotificationService initializes the notification service with the required DAOs
func InitNotificationService(redisDao dao.RedisDaoImpl, scyllaDao dao.ScyllaDbDaoImpl) {
	notificationServiceInstance = &NotificationServiceMethodsImpl{
		redisDao:  redisDao,
		scyllaDao: scyllaDao,
	}
}

func GetNotificationServiceInstance() *NotificationServiceMethodsImpl {
	return notificationServiceInstance
}

// now we have to define the service functions which shall have use the repo/dao layer.
func (notificationServiceInstance *NotificationServiceMethodsImpl) SendSMSService(ctx context.Context, req models.SendSms) (string, error) {
	logger := utils.RequestLogger(ctx, "service", "send_sms")

	// here we have to hit the db and create the db entry and then publish to the producer too.
	// create a blank entry for a new SMS
	// create a payload and then throw into kafka
	requestID := uuid.New().String()

	logger.Info().
		Str("request_id", requestID).
		Str("phone_number", req.PhoneNumber).
		Msg("Processing SMS send request")

	incomingReq := models.AddSmsEntryInDb{
		RequestID:   requestID,
		PhoneNumber: req.PhoneNumber,
		Message:     req.Message,
	}
	err := notificationServiceInstance.scyllaDao.InsertSMSRequest(ctx, incomingReq)
	if err != nil {
		logger.Error().
			Err(err).
			Str("request_id", requestID).
			Str("phone_number", req.PhoneNumber).
			Msg("Failed to insert SMS request into database")
		return "", err
	}

	logger.Info().
		Str("request_id", requestID).
		Str("phone_number", req.PhoneNumber).
		Msg("Successfully created SMS request")
	return requestID, nil
}

func (notificationServiceInstance *NotificationServiceMethodsImpl) HandleKafkaMessages(ctx context.Context, requestId string) error {
	logger := utils.RequestLogger(ctx, "service", "handle_kafka_message")

	logger.Info().
		Str("request_id", requestId).
		Msg("Processing Kafka message for SMS request")

	// db call
	smsDetails, err := notificationServiceInstance.scyllaDao.GetSMSDetailsFromDB(ctx, requestId)
	if err != nil {
		logger.Error().
			Err(err).
			Str("request_id", requestId).
			Msg("Failed to retrieve SMS details from database")
		return nil
	}

	isPresent, err := notificationServiceInstance.redisDao.CheckNumberInBlacklistedSet(ctx, smsDetails.PhoneNumber)
	if err != nil {
		logger.Error().
			Err(err).
			Str("request_id", requestId).
			Str("phone_number", smsDetails.PhoneNumber).
			Msg("Failed to check blacklist status")
		return nil
	}

	if isPresent {
		logger.Warn().
			Str("request_id", requestId).
			Str("phone_number", smsDetails.PhoneNumber).
			Msg("SMS blocked - phone number is blacklisted")
		return nil
	}

	// if not present
	logger.Info().
		Str("request_id", requestId).
		Str("phone_number", smsDetails.PhoneNumber).
		Msg("Sending SMS to external service")
	notificationServiceInstance.SendMessage(smsDetails)

	err = notificationServiceInstance.scyllaDao.UpdateSMSDetailsInDB(ctx, smsDetails)
	if err != nil {
		logger.Error().
			Err(err).
			Str("request_id", requestId).
			Msg("Failed to update SMS status in database")
		return err
	}

	logger.Info().
		Str("request_id", requestId).
		Str("phone_number", smsDetails.PhoneNumber).
		Msg("Successfully processed SMS request")
	return nil
}

func (notificationServiceInstance *NotificationServiceMethodsImpl) SendMessage(req *models.SMSRequest) {
	logger := utils.ComponentLogger("sms_gateway")

	logger.Info().
		Str("request_id", req.ID).
		Str("phone_number", req.PhoneNumber).
		Msg("Sending SMS via external gateway")

	fmt.Printf("RequestID: %v | PhoneNumber: %v | Message: %s\n", req.ID, req.PhoneNumber, req.Message)
}

// we have to define a model for this.
// we have to fix a database schema fr
func (notificationServiceInstance *NotificationServiceMethodsImpl) GetSMSService(ctx context.Context, reqID string) (any, error) {
	logger := utils.RequestLogger(ctx, "service", "get_sms")

	logger.Info().
		Str("request_id", reqID).
		Msg("Retrieving SMS request details")

	// we have to hit db and fetch the sms details by request ID.
	smsDetails, err := notificationServiceInstance.scyllaDao.GetSMSDetailsFromDB(ctx, reqID)
	if err != nil {
		logger.Error().
			Err(err).
			Str("request_id", reqID).
			Msg("Failed to retrieve SMS details from database")
		return nil, fmt.Errorf("failed to retrieve SMS details for request ID %s", reqID)
	}

	logger.Info().
		Str("request_id", reqID).
		Str("status", smsDetails.Status).
		Msg("Successfully retrieved SMS details")
	return smsDetails, nil
}

func (notificationServiceInstance *NotificationServiceMethodsImpl) GetBlacklistService(ctx context.Context) ([]string, error) {
	logger := utils.RequestLogger(ctx, "service", "get_blacklist")

	logger.Info().Msg("Retrieving blacklisted numbers")

	// we have to hit the db and fetch all the blacklisted numbers
	blacklistedNumbers, err := notificationServiceInstance.redisDao.GetAllBlacklistedNumbers(ctx)
	if err != nil {
		logger.Error().
			Err(err).
			Msg("Failed to retrieve blacklisted numbers")
		return nil, fmt.Errorf("failed to retrieve blacklisted numbers")
	}

	logger.Info().
		Int("count", len(blacklistedNumbers)).
		Msg("Successfully retrieved blacklisted numbers")
	return blacklistedNumbers, nil
}

func (notificationServiceInstance *NotificationServiceMethodsImpl) AddToBlacklistService(ctx context.Context, req models.AddToBlacklist) error {
	logger := utils.RequestLogger(ctx, "service", "add_to_blacklist")

	logger.Info().
		Str("phone_number", req.PhoneNumbers).
		Msg("Adding number to blacklist")

	err := notificationServiceInstance.redisDao.AddNumberToBlacklistedSet(ctx, req.PhoneNumbers)
	if err != nil {
		logger.Error().
			Err(err).
			Str("phone_number", req.PhoneNumbers).
			Msg("Failed to add number to blacklist")
		return fmt.Errorf("failed to add number %s to blacklist", req.PhoneNumbers)
	}

	logger.Info().
		Str("phone_number", req.PhoneNumbers).
		Msg("Successfully added number to blacklist")
	return nil
}

func (notificationServiceInstance *NotificationServiceMethodsImpl) RemoveFromBlacklistService(ctx context.Context, number string) (bool, error) {
	logger := utils.RequestLogger(ctx, "service", "remove_from_blacklist")

	logger.Info().
		Str("phone_number", number).
		Msg("Removing number from blacklist")

	removedCount, err := notificationServiceInstance.redisDao.RemoveFromBlacklistedSet(ctx, number)
	if err != nil {
		logger.Error().
			Err(err).
			Str("phone_number", number).
			Msg("Failed to remove number from blacklist")
		return false, fmt.Errorf("failed to remove number %s from blacklist", number)
	}

	success := removedCount > 0
	logger.Info().
		Str("phone_number", number).
		Bool("removed", success).
		Int64("removed_count", removedCount).
		Msg("Blacklist removal completed")

	return success, nil
}
