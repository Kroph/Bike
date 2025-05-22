package service

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/smtp"
	"time"
)

type EmailService interface {
	SendOrderConfirmation(to string, orderID string, orderDetails map[string]interface{}) error
	SendEmailVerification(to string, username string, verificationToken string) error
	SendEmailVerificationCode(to string, username string, code string) error
}

type emailService struct {
	from     string
	password string
	host     string
	port     string
	auth     smtp.Auth
}

func NewEmailService(from, password, host, port string) EmailService {
	auth := smtp.PlainAuth("", from, password, host)
	return &emailService{
		from:     from,
		password: password,
		host:     host,
		port:     port,
		auth:     auth,
	}
}

func (s *emailService) SendOrderConfirmation(to string, orderID string, orderDetails map[string]interface{}) error {
	subject := fmt.Sprintf("Bicycle Store - Order Confirmation #%s", orderID)

	// Add the order ID to the details map
	orderDetails["OrderID"] = orderID
	orderDetails["Date"] = time.Now().Format("January 2, 2006")

	// Parse the HTML email template
	t, err := template.New("order_confirmation").Parse(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Bicycle Store - Order Confirmation</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { text-align: center; padding: 10px; background-color: #4CAF50; color: white; }
        .footer { text-align: center; margin-top: 30px; font-size: 12px; color: #6c757d; }
        table { width: 100%; border-collapse: collapse; margin: 20px 0; }
        th, td { padding: 10px; text-align: left; border-bottom: 1px solid #ddd; }
        th { background-color: #f8f9fa; }
        .total { font-weight: bold; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Bicycle Store - Order Confirmation</h1>
        </div>
        
        <p>Dear Cyclist,</p>
        
        <p>Thank you for your order. We're pleased to confirm that we've received your order and it's being processed.</p>
        
        <h2>Order Details</h2>
        <p><strong>Order Number:</strong> {{.OrderID}}<br>
        <strong>Order Date:</strong> {{.Date}}</p>
        
        <table>
            <tr>
                <th>Bicycle</th>
                <th>Quantity</th>
                <th>Price</th>
                <th>Subtotal</th>
            </tr>
            {{range .Items}}
            <tr>
                <td>{{.Name}}</td>
                <td>{{.Quantity}}</td>
                <td>${{.Price}}</td>
                <td>${{.Subtotal}}</td>
            </tr>
            {{end}}
            <tr class="total">
                <td colspan="3">Total</td>
                <td>${{.Total}}</td>
            </tr>
        </table>
        
        <p>We'll send you another email when your bicycle is ready for delivery.</p>
        
        <p>Thank you for shopping with Bicycle Store!</p>
        
        <div class="footer">
            <p>This is an automated email, please do not reply to this message.</p>
            <p>Bicycle Store - Your Cycling Partner</p>
        </div>
    </div>
</body>
</html>
`)
	if err != nil {
		return fmt.Errorf("failed to parse email template: %v", err)
	}

	var body bytes.Buffer
	if err := t.Execute(&body, orderDetails); err != nil {
		return fmt.Errorf("failed to execute email template: %v", err)
	}

	return s.sendEmail(to, subject, body.String())
}

func (s *emailService) SendEmailVerification(to string, username string, verificationToken string) error {
	log.Printf("[EMAIL-SERVICE] Starting email verification send to: %s", to)
	log.Printf("[EMAIL-SERVICE] Using SMTP host: %s:%s", s.host, s.port)
	log.Printf("[EMAIL-SERVICE] From address: %s", s.from)

	subject := "Bicycle Store - Verify Your Email Address"

	data := map[string]string{
		"Username":         username,
		"VerificationLink": fmt.Sprintf("http://localhost:3000/verify-email?token=%s", verificationToken), // Update with your frontend URL
	}

	log.Printf("[EMAIL-SERVICE] Verification link: %s", data["VerificationLink"])

	t, err := template.New("email_verification").Parse(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Bicycle Store - Email Verification</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { text-align: center; padding: 10px; background-color: #4CAF50; color: white; }
        .footer { text-align: center; margin-top: 30px; font-size: 12px; color: #6c757d; }
        .button { display: inline-block; padding: 10px 20px; background-color: #4CAF50; color: white; 
                  text-decoration: none; border-radius: 4px; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Bicycle Store - Verify Your Email Address</h1>
        </div>
        
        <p>Hello {{.Username}},</p>
        
        <p>Thank you for registering with Bicycle Store. Please verify your email address by clicking the button below:</p>
        
        <p style="text-align: center;">
            <a href="{{.VerificationLink}}" class="button">Verify Email</a>
        </p>
        
        <p>If the button doesn't work, you can also copy and paste the following link into your browser:</p>
        
        <p style="word-break: break-all;">{{.VerificationLink}}</p>
        
        <p>This link is valid for 24 hours. If you didn't request this email, please ignore it.</p>
        
        <p>Happy cycling!</p>
        
        <div class="footer">
            <p>This is an automated email, please do not reply to this message.</p>
            <p>Bicycle Store - Your Cycling Partner</p>
        </div>
    </div>
</body>
</html>
`)
	if err != nil {
		log.Printf("[EMAIL-SERVICE] Failed to parse email template: %v", err)
		return fmt.Errorf("failed to parse email template: %v", err)
	}

	var body bytes.Buffer
	if err := t.Execute(&body, data); err != nil {
		log.Printf("[EMAIL-SERVICE] Failed to execute email template: %v", err)
		return fmt.Errorf("failed to execute email template: %v", err)
	}

	log.Printf("[EMAIL-SERVICE] Email template rendered successfully")

	err = s.sendEmail(to, subject, body.String())
	if err != nil {
		log.Printf("[EMAIL-SERVICE] Failed to send email: %v", err)
		return err
	}

	log.Printf("[EMAIL-SERVICE] Email sent successfully to %s", to)
	return nil
}

func (s *emailService) sendEmail(to, subject, htmlBody string) error {
	log.Printf("[EMAIL-SERVICE] Preparing to send email to: %s", to)

	// Set email headers
	headers := make(map[string]string)
	headers["From"] = s.from
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	// Construct message
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + htmlBody

	log.Printf("[EMAIL-SERVICE] Connecting to SMTP server: %s:%s", s.host, s.port)

	// Connect to SMTP server using plain TCP first (not TLS)
	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Printf("[EMAIL-SERVICE] Failed to connect to SMTP server: %v", err)
		return fmt.Errorf("failed to connect to SMTP server: %v", err)
	}
	defer conn.Close()

	log.Printf("[EMAIL-SERVICE] Successfully connected to SMTP server")

	// Create SMTP client
	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		log.Printf("[EMAIL-SERVICE] Failed to create SMTP client: %v", err)
		return fmt.Errorf("failed to create SMTP client: %v", err)
	}
	defer client.Close()

	log.Printf("[EMAIL-SERVICE] SMTP client created successfully")

	// Start TLS using STARTTLS
	tlsConfig := &tls.Config{
		ServerName: s.host,
	}

	if err = client.StartTLS(tlsConfig); err != nil {
		log.Printf("[EMAIL-SERVICE] Failed to start TLS: %v", err)
		return fmt.Errorf("failed to start TLS: %v", err)
	}

	log.Printf("[EMAIL-SERVICE] TLS started successfully")

	// Authenticate
	log.Printf("[EMAIL-SERVICE] Attempting authentication...")
	if err = client.Auth(s.auth); err != nil {
		log.Printf("[EMAIL-SERVICE] Authentication failed: %v", err)
		return fmt.Errorf("authentication failed: %v", err)
	}

	log.Printf("[EMAIL-SERVICE] Authentication successful")

	// Set the sender and recipient
	if err = client.Mail(s.from); err != nil {
		log.Printf("[EMAIL-SERVICE] Failed to set sender: %v", err)
		return fmt.Errorf("failed to set sender: %v", err)
	}

	if err = client.Rcpt(to); err != nil {
		log.Printf("[EMAIL-SERVICE] Failed to set recipient: %v", err)
		return fmt.Errorf("failed to set recipient: %v", err)
	}

	log.Printf("[EMAIL-SERVICE] Sender and recipient set successfully")

	// Send the email body
	wc, err := client.Data()
	if err != nil {
		log.Printf("[EMAIL-SERVICE] Failed to start data command: %v", err)
		return fmt.Errorf("failed to start data command: %v", err)
	}

	_, err = fmt.Fprintf(wc, message)
	if err != nil {
		log.Printf("[EMAIL-SERVICE] Failed to write email data: %v", err)
		return fmt.Errorf("failed to send email: %v", err)
	}

	err = wc.Close()
	if err != nil {
		log.Printf("[EMAIL-SERVICE] Failed to close email data writer: %v", err)
		return fmt.Errorf("failed to close connection: %v", err)
	}

	log.Printf("[EMAIL-SERVICE] Email sent successfully to %s", to)
	return nil
}

// Add this method to api-gateway/service/email_service.go

func (s *emailService) SendEmailVerificationCode(to string, username string, code string) error {
	log.Printf("[EMAIL-SERVICE] Starting email verification send to: %s with code: %s", to, code)

	subject := "Bicycle Store - Email Verification Code"

	// Create a template data map
	data := map[string]string{
		"Username": username,
		"Code":     code,
	}

	// Parse the HTML email template
	t, err := template.New("email_verification_code").Parse(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Bicycle Store</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { text-align: center; padding: 10px; background-color: #4CAF50; color: white; }
        .footer { text-align: center; margin-top: 30px; font-size: 12px; color: #6c757d; }
        .code-box { 
            background-color: #f8f9fa; 
            border: 2px dashed #4CAF50; 
            padding: 20px; 
            text-align: center; 
            margin: 20px 0;
            border-radius: 8px;
        }
        .verification-code { 
            font-size: 32px; 
            font-weight: bold; 
            color: #4CAF50; 
            letter-spacing: 8px;
            font-family: 'Courier New', monospace;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Bicycle Store - Email Verification</h1>
        </div>
        
        <p>Hello {{.Username}},</p>
        
        <div class="code-box">
            <p style="margin: 0; font-size: 18px; color: #666;">Your verification code is:</p>
            <div class="verification-code">{{.Code}}</div>
        </div>
    </div>
</body>
</html>
`)
	if err != nil {
		log.Printf("[EMAIL-SERVICE] Failed to parse email template: %v", err)
		return fmt.Errorf("failed to parse email template: %v", err)
	}

	var body bytes.Buffer
	if err := t.Execute(&body, data); err != nil {
		log.Printf("[EMAIL-SERVICE] Failed to execute email template: %v", err)
		return fmt.Errorf("failed to execute email template: %v", err)
	}

	log.Printf("[EMAIL-SERVICE] Email template rendered successfully")

	err = s.sendEmail(to, subject, body.String())
	if err != nil {
		log.Printf("[EMAIL-SERVICE] Failed to send email: %v", err)
		return err
	}

	log.Printf("[EMAIL-SERVICE] Verification code email sent successfully to %s", to)
	return nil
}
