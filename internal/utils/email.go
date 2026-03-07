package utils

import (
	"fmt"
	"log"
	"net/smtp"

	"github.com/dr15/internship-hub-api/config"
)

func SendResetPasswordEmail(toEmail, token string) error {
	conf := config.AppConfig

	// In development, if SMTP is not configured, just log the token
	if conf.SMTPHost == "localhost" || conf.SMTPUser == "" {
		fmt.Printf("\n--- DEVELOPMENT EMAIL MOCK ---\n")
		fmt.Printf("To: %s\n", toEmail)
		fmt.Printf("Subject: Reset Password\n")
		fmt.Printf("Link: http://localhost:5173/reset-password?token=%s\n", token)
		fmt.Printf("------------------------------\n\n")
		return nil
	}

	subject := "Subject: Reset Password - Internship Hub\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body := fmt.Sprintf(`
		<h3>Reset Password</h3>
		<p>Anda menerima email ini karena Anda (atau seseorang) meminta reset password untuk akun Anda.</p>
		<p>Silakan klik link di bawah ini untuk mereset password Anda:</p>
		<a href="http://localhost:5173/reset-password?token=%s">Reset Password</a>
		<p>Jika Anda tidak meminta ini, abaikan email ini.</p>
	`, token)

	msg := []byte(subject + mime + body)
	auth := smtp.PlainAuth("", conf.SMTPUser, conf.SMTPPass, conf.SMTPHost)

	err := smtp.SendMail(conf.SMTPHost+":"+conf.SMTPPort, auth, conf.SMTPSender, []string{toEmail}, msg)
	if err != nil {
		log.Printf("Failed to send email: %v", err)
		return err
	}

	return nil
}
