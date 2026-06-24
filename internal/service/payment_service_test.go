package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
)

func TestPaymentServiceVerifySignature(t *testing.T) {
	const secret = "webhook-secret"
	body := []byte(`{"ref_id":"ref-123"}`)

	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	service := &PaymentService{webhookSecret: secret}
	if !service.VerifySignature(body, signature) {
		t.Fatal("expected valid signature to be accepted")
	}
	if service.VerifySignature([]byte(`{"changed":true}`), signature) {
		t.Fatal("expected signature for a different body to be rejected")
	}
}

func TestNewPaymentTransactionCode(t *testing.T) {
	first, err := newPaymentTransactionCode()
	if err != nil {
		t.Fatalf("generate first transaction code: %v", err)
	}
	second, err := newPaymentTransactionCode()
	if err != nil {
		t.Fatalf("generate second transaction code: %v", err)
	}

	if !strings.HasPrefix(first, "PAY-") {
		t.Fatalf("expected PAY- prefix, got %q", first)
	}
	if len(first) != len("PAY-")+24 {
		t.Fatalf("expected 24 hexadecimal characters, got %q", first)
	}
	if first == second {
		t.Fatalf("expected unique transaction codes, both were %q", first)
	}
}
