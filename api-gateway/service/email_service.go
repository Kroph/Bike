package service

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"time"
)

type EmailService interface {
	SendOrderConfirmation(to string, orderID string, orderDetails map[string]interface{}) error
	SendEmailVerification(to string, username string, verificationToken string) error
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
	subject := fmt.Sprintf("Order Confirmation #%s", orderID)

	// Add the order ID to the details map
	orderDetails["OrderID"] = orderID
	orderDetails["Date"] = time.Now().Format("January 2, 2006")

	// Parse the HTML email template
	t, err := template.New("order_confirmation").Parse(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Order Confirmation</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { text-align: center; padding: 10px; background-color: #f8f9fa; }
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
            <h1>Order Confirmation</h1>
        </div>
        
        <p>Dear Customer,</p>
        
        <p>Thank you for your order. We're pleased to confirm that we've received your order and it's being processed.</p>
        
        <h2>Order Details</h2>
        <p><strong>Order Number:</strong> {{.OrderID}}<br>
        <strong>Order Date:</strong> {{.Date}}</p>
        
        <table>
            <tr>
                <th>Product</th>
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
        
        <p>We'll send you another email when your order ships.</p>
        
        <p>Thank you for shopping with us!</p>
        
        <div class="footer">
            <p>This is an automated email, please do not reply to this message.</p>
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
	subject := "Verify Your Email Address"

	// Create a template data map
	data := map[string]string{
		"Username":         username,
		"VerificationLink": fmt.Sprintf("https://yourdomain.com/verify-email?token=%s", verificationToken),
	}

	// Parse the HTML email template
	t, err := template.New("email_verification").Parse(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Email Verification</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { text-align: center; padding: 10px; background-color: #f8f9fa; }
        .footer { text-align: center; margin-top: 30px; font-size: 12px; color: #6c757d; }
        .button { display: inline-block; padding: 10px 20px; background-color: #007bff; color: white; 
                  text-decoration: none; border-radius: 4px; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Verify Your Email Address</h1>
        </div>
        
        <p>Hello {{.Username}},</p>
        
        <p>Thank you for registering. Please verify your email address by clicking the button below:</p>
        
        <p style="text-align: center;">
            <a href="{{.VerificationLink}}" class="button">Verify Email</a>
        </p>
        
        <p>If the button doesn't work, you can also copy and paste the following link into your browser:</p>
        
        <p style="word-break: break-all;">{{.VerificationLink}}</p>
        
        <p>This link is valid for 24 hours. If you didn't request this email, please ignore it.</p>
        
        <div class="footer">
            <p>This is an automated email, please do not reply to this message.</p>
        </div>
    </div>
</body>
</html>
`)
	if err != nil {
		return fmt.Errorf("failed to parse email template: %v", err)
	}

	var body bytes.Buffer
	if err := t.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute email template: %v", err)
	}

	return s.sendEmail(to, subject, body.String())
}

func (s *emailService) sendEmail(to, subject, htmlBody string) error {
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

	// Gmail requires TLS
	// Create TLS configuration
	tlsConfig := &tls.Config{
		ServerName: s.host,
	}

	// Connect to the SMTP server
	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %v", err)
	}
	defer conn.Close()

	// Create a new SMTP client
	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %v", err)
	}
	defer client.Close()

	// Authenticate
	if err = client.Auth(s.auth); err != nil {
		return fmt.Errorf("authentication failed: %v", err)
	}

	// Set the sender and recipient
	if err = client.Mail(s.from); err != nil {
		return fmt.Errorf("failed to set sender: %v", err)
	}

	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("failed to set recipient: %v", err)
	}

	// Send the email body
	wc, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to start data command: %v", err)
	}

	_, err = fmt.Fprintf(wc, message)
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	err = wc.Close()
	if err != nil {
		return fmt.Errorf("failed to close connection: %v", err)
	}

	log.Printf("Email sent to %s", to)
	return nil
}
