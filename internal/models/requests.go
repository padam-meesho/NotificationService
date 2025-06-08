package models

type SendSms struct {
	PhoneNumber string `json:"phone_number"`
	Message     string `json:"message"`
}

type AddToBlacklist struct {
	PhoneNumbers string `json:"phone_numbers"`
}

type AddSmsEntryInDb struct {
	RequestID   string `json:"request_id"`
	PhoneNumber string `json:"phone_number"`
	Message     string `json:"message"`
}

type GetSmsDetailsFromDbRequest struct {
	RequestId string `json:"request_id"`
}

type UpdateSmsDetailsInDbRequest struct {
	ID              string `json:"id"`
	PhoneNumber     string `json:"phone_number"`
	Status          string `json:"status"` // this can have "success", "failed", "pending values"
	FailureCode     string `json:"failure_code,omitempty"`
	FailureComments string `json:"failure_comments,omitempty"`
}
