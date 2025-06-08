package dao

import (
	"context"
	"sync"
	"time"

	"github.com/padam-meesho/NotificationService/config"
	"github.com/padam-meesho/NotificationService/internal/models"
	"github.com/padam-meesho/NotificationService/internal/utils"
	"github.com/scylladb/gocqlx/v2"
	"github.com/scylladb/gocqlx/v2/qb"
)

type ScyllaDbDao interface {
	// list all the methods being implemented
	InsertSMSRequest(ctx context.Context, sms models.AddSmsEntryInDb) error
	GetSMSDetailsFromDB(ctx context.Context, requestId string) (*models.SMSRequest, error)
	UpdateSMSDetailsInDB(ctx context.Context, smsDetails *models.SMSRequest) error
}

type ScyllaDbDaoImpl struct {
	// this will have a scylladb client/session
	scyllaSession *gocqlx.Session
}

var (
	scyallaOnce     sync.Once
	scyllaDbSession *ScyllaDbDaoImpl
)

// this shall be a global standalone function.
func NewScyllaSessionDao() *ScyllaDbDaoImpl {
	scyallaOnce.Do(func() {
		scyllaDbSession = &ScyllaDbDaoImpl{
			scyllaSession: config.GetScyllaSession().ScyllaSession,
		}
	})
	return scyllaDbSession
}

// this one shall also be a global standalone function
func GetScyllaSession() *ScyllaDbDaoImpl {
	return scyllaDbSession
}

// Example: Insert into sms_requests using gocqlx/qb
func (session ScyllaDbDaoImpl) InsertSMSRequest(ctx context.Context, sms models.AddSmsEntryInDb) error {
	logger := utils.DatabaseLogger(ctx, "insert", "sms_requests", sms.RequestID)

	logger.Info().
		Str("phone_number", sms.PhoneNumber).
		Str("message_preview", sms.Message[:min(len(sms.Message), 50)]).
		Msg("Attempting to insert SMS request")

	query := qb.Insert("sms_requests").
		Columns("id",
			"phone_number",
			"message",
			"status",
			"failure_code",
			"failure_comments",
			"created_at",
			"updated_at").
		QueryContext(ctx, *session.scyllaSession)

	insertMap := qb.M{
		"id":               sms.RequestID,
		"phone_number":     sms.PhoneNumber,
		"message":          sms.Message,
		"status":           "Pending", // replace with an enum value
		"failure_code":     "",
		"failure_comments": "",
		"created_at":       time.Now(),
		"updated_at":       time.Now(),
	}

	err := query.BindMap(insertMap).ExecRelease()
	if err != nil {
		logger.Error().
			Err(err).
			Str("phone_number", sms.PhoneNumber).
			Msg("Failed to insert SMS request into database")
		return err
	}

	logger.Info().
		Str("phone_number", sms.PhoneNumber).
		Msg("Successfully inserted SMS request into database")
	return nil
}

// okay now we have to create a scylla entry in the keyspace and the table specified, the request must be of the valid DTO, and then it should update it.
func (session ScyllaDbDaoImpl) GetSMSDetailsFromDB(ctx context.Context, requestId string) (*models.SMSRequest, error) {
	logger := utils.DatabaseLogger(ctx, "select", "sms_requests", requestId)

	logger.Info().Msg("Attempting to retrieve SMS request from database")

	var data models.SMSRequest
	query := qb.Select("sms_requests").
		Columns("id", "phone_number", "message", "status", "failure_code", "failure_comments", "created_at", "updated_at").
		Where(qb.Eq("id")).
		QueryContext(ctx, *session.scyllaSession)

	err := query.BindMap(qb.M{
		"id": requestId,
	}).GetRelease(&data)

	if err != nil {
		logger.Error().
			Err(err).
			Msg("Failed to retrieve SMS request from database")
		return nil, err
	}

	logger.Info().
		Str("phone_number", data.PhoneNumber).
		Str("status", data.Status).
		Msg("Successfully retrieved SMS request from database")
	return &data, nil
}

func (session ScyllaDbDaoImpl) UpdateSMSDetailsInDB(ctx context.Context, smsDetails *models.SMSRequest) error {
	logger := utils.DatabaseLogger(ctx, "update", "sms_requests", smsDetails.ID)

	logger.Info().
		Str("new_status", smsDetails.Status).
		Msg("Attempting to update SMS request in database")

	// Build the update query
	query := qb.Update("sms_requests").
		Set(
			"status",
			"failure_code",
			"failure_comments",
			"updated_at",
		).
		Where(qb.Eq("id")).
		QueryContext(ctx, *session.scyllaSession)

	// this is a mechanism to fill in some default values into the db.
	status := smsDetails.Status
	if status == "" {
		status = "Success"
	}
	failureCode := smsDetails.FailureCode
	if failureCode == "" {
		failureCode = "null"
	}
	failureComments := smsDetails.FailureComments
	if failureComments == "" {
		failureComments = "null"
	}
	updateMap := qb.M{
		"id":               smsDetails.ID,
		"status":           status,
		"failure_code":     failureCode,
		"failure_comments": failureComments,
		"updated_at":       time.Now(),
	}

	err := query.BindMap(updateMap).ExecRelease()
	if err != nil {
		logger.Error().
			Err(err).
			Str("new_status", status).
			Str("failure_code", failureCode).
			Msg("Failed to update SMS request in database")
		return err
	}

	logger.Info().
		Str("new_status", status).
		Msg("Successfully updated SMS request in database")
	return nil
}

// Helper function for min operation
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
