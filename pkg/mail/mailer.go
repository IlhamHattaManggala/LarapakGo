package mail

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strconv"
	"strings"
)

// Message holds email sending parameters.
type Message struct {
	From     string   `json:"from"`
	FromName string   `json:"from_name"`
	To       []string `json:"to"`
	Subject  string   `json:"subject"`
	Body     string   `json:"body"`
	HTML     bool     `json:"html"`
}

// Config stores SMTP configuration.
type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	FromName string
}

// GetConfig reads SMTP configuration from environment variables.
func GetConfig() Config {
	host := os.Getenv("MAIL_HOST")
	portStr := os.Getenv("MAIL_PORT")
	port, _ := strconv.Atoi(portStr)
	if port <= 0 {
		port = 587
	}
	user := os.Getenv("MAIL_USERNAME")
	pass := os.Getenv("MAIL_PASSWORD")
	from := os.Getenv("MAIL_FROM")
	if from == "" {
		from = "no-reply@larapak.local"
	}
	fromName := os.Getenv("MAIL_FROM_NAME")
	if fromName == "" {
		fromName = "LarapakGo"
	}

	return Config{
		Host:     host,
		Port:     port,
		Username: user,
		Password: pass,
		From:     from,
		FromName: fromName,
	}
}

// Send dispatches an email using net/smtp. In dev without host configured, it logs cleanly.
func Send(msg Message) error {
	cfg := GetConfig()
	if msg.From == "" {
		msg.From = cfg.From
	}
	if msg.FromName == "" {
		msg.FromName = cfg.FromName
	}

	if len(msg.To) == 0 {
		return fmt.Errorf("email mbutuhna panampa (recipient 'to' ora kena kosong)")
	}

	// In local development without configured SMTP server, log message without failure
	if cfg.Host == "" || (cfg.Host == "localhost" && cfg.Username == "") {
		log.Printf("[MAIL LOG] Saka: %s <%s> | Marang: %s | Subyek: %s | Dawa Isi: %d bytes",
			msg.FromName, msg.From, strings.Join(msg.To, ", "), msg.Subject, len(msg.Body))
		return nil
	}

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	contentType := "text/plain; charset=UTF-8"
	if msg.HTML || strings.Contains(msg.Body, "<html") || strings.Contains(msg.Body, "</div>") || strings.Contains(msg.Body, "<p>") {
		contentType = "text/html; charset=UTF-8"
	}

	headers := fmt.Sprintf("From: %s <%s>\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: %s\r\n\r\n",
		msg.FromName, msg.From, strings.Join(msg.To, ", "), msg.Subject, contentType)

	bodyBytes := []byte(headers + msg.Body)

	var auth smtp.Auth
	if cfg.Username != "" && cfg.Password != "" {
		auth = smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	}

	// SSL/TLS direct port 465
	if cfg.Port == 465 {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         cfg.Host,
		}
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("gagal nyambung SMTPS: %w", err)
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, cfg.Host)
		if err != nil {
			return err
		}
		defer client.Quit()

		if auth != nil {
			if err = client.Auth(auth); err != nil {
				return err
			}
		}
		if err = client.Mail(msg.From); err != nil {
			return err
		}
		for _, to := range msg.To {
			if err = client.Rcpt(to); err != nil {
				return err
			}
		}
		w, err := client.Data()
		if err != nil {
			return err
		}
		_, err = w.Write(bodyBytes)
		if err != nil {
			return err
		}
		return w.Close()
	}

	return smtp.SendMail(addr, auth, msg.From, msg.To, bodyBytes)
}
