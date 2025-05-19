package service

import (
	"encoding/json"
	"log"
)

// MockEmailService implements EmailService but just logs the emails instead of sending them
type MockEmailService struct{}

func (s *MockEmailService) SendOrderConfirmation(to string, orderID string, orderDetails map[string]interface{}) error {
	detailsJSON, _ := json.Marshal(orderDetails)
	log.Printf("[MOCK EMAIL] Order confirmation for order %s to %s. Details: %s",
		orderID, to, string(detailsJSON))
	return nil
}

func (s *MockEmailService) SendEmailVerification(to string, username string, verificationToken string) error {
	log.Printf("[MOCK EMAIL] Email verification to %s (username: %s) with token: %s",
		to, username, verificationToken)
	return nil
}
