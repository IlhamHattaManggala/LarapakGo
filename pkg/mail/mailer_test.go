package mail

import (
	"testing"
)

func TestSendInDevMode(t *testing.T) {
	msg := Message{
		To:      []string{"test@ngapak.org"},
		Subject: "Sugeng Rawuh",
		Body:    "Halo saka LarapakGo!",
	}

	err := Send(msg)
	if err != nil {
		t.Fatalf("expected nil error in dev fallback, got: %v", err)
	}
}

func TestSendMissingRecipient(t *testing.T) {
	msg := Message{
		Subject: "Ora Ana Panampa",
		Body:    "Test",
	}

	err := Send(msg)
	if err == nil {
		t.Fatalf("expected error when recipient list is empty")
	}
}
