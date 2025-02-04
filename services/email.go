package services

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strconv"
	"text/template"
	"time"
	"zero/utils"

	"github.com/charmbracelet/log"
	"gopkg.in/gomail.v2"
)

type EmailService struct {
	DB *sql.DB
}

type EmailData struct {
	RecipientName string
	Code          string
}

func (ES *EmailService) GenerateVerificationCode() (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	code := make([]byte, 6)
	for i := range code {
		code[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(code), nil
}

func (EM *EmailService) AddVerificationEmailCode(user int, code string) error {
	expiration := time.Now().Add(15 * time.Minute) // expires in 15 minutes
	rows, err := EM.DB.Exec("INSERT INTO verification (user, code, expiration) VALUES (?, ?, ?)", user, code, expiration)
	if err != nil {
		utils.HandleSQLiteErrors(err)
		log.Error("failed to insert into the database", "err", err)
		return err
	}
	affacted, err := rows.RowsAffected()
	if affacted == 0 {
		log.Error("no affacted rows", "err", err)
		return err
	}
	if err != nil {
		log.Error("failed to add verification code to the database", "err", err)
		return err
	}
	return nil
}

func (ES *EmailService) SendVerificationEmail(email, recipient, code string) error {
	from := os.Getenv("SMTP_SERVER_EMAIL")
	smtpHost := os.Getenv("SMTP_SERVER_ADDRESS")
	port, err := strconv.Atoi(os.Getenv("SMTP_SERVER_PORT"))
	if err != nil {
		log.Error("failed to get smtp server port", "err", err)
		return err
	}
	smtpPort := port
	smtpUser := os.Getenv("SMTP_SERVER_USER")
	smtpPassword := os.Getenv("SMTP_SERVER_PASSWORD")
	templateFile, err := os.Open("email/verification.html")
	if err != nil {
		log.Error("failed to open email template", "err", err)
		return err
	}
	templateData, err := io.ReadAll(templateFile)
	if err != nil {
		log.Error("failed to read email template", "err", err)
		return err
	}
	defer templateFile.Close()
	t := template.New("emailTemplate")
	t, err = t.Parse(string(templateData))
	if err != nil {
		log.Error("failed to parse email template", "err", err)
		return err
	}
	data := EmailData{
		RecipientName: recipient,
		Code:          code,
	}
	var body bytes.Buffer
	err = t.Execute(&body, data)
	if err != nil {
		log.Error("failed to execute template", "err", err)
		return err
	}
	message := gomail.NewMessage()
	message.SetHeader("From", from)
	message.SetHeader("To", email)
	message.SetHeader("Subject", os.Getenv("APP_NAME")+" Email Verification")
	message.SetBody("text/html", body.String())
	dialer := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPassword)
	if err := dialer.DialAndSend(message); err != nil {
		log.Error("failed to send email", "err", err)
		return err
	}
	return nil
}

func (es *EmailService) CompareVerificationCode(user int, code string) error {
	var storedCode string
	var expiration time.Time
	err := es.DB.QueryRow("SELECT code, expiration FROM verification WHERE user = ? AND code = ?", user, code).
		Scan(&storedCode, &expiration)
	if err != nil {
		utils.HandleSQLiteErrors(err)
		return err
	}
	if time.Now().After(expiration) {
		return fmt.Errorf("verification code has expired")
	}
	_, err = es.DB.Exec("DELETE from verification WHERE user = ?", user)
	if err != nil {
		log.Error("failed to delete used verification code")
		return err
	}
	return nil
}

func (es *EmailService) MarkUserAsVerified(user int) error {
	_, err := es.DB.Exec("UPDATE users SET verified = 1 WHERE id = ?", user)
	if err != nil {
		log.Error("failed to update verification status")
		return err
	}
	return nil
}
