package models

import "time"

type SMSRequest struct {
	ID              string    `json:"id" cql:"id"`
	PhoneNumber     string    `json:"phone_number" cql:"phone_number"`
	Message         string    `json:"message" cql:"message"`
	Status          string    `json:"status" cql:"status"` // this can have "success", "failed", "pending values"
	FailureCode     string    `json:"failure_code" cql:"failure_code"`
	FailureComments string    `json:"failure_comments" cql:"failure_comments"`
	CreatedAt       time.Time `json:"created_at" cql:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" cql:"updated_at"`
}
