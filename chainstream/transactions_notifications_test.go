// Package chainstream_test contains tests for parsing and classifying transaction notifications.
//
// NOTE: The included transaction samples (e.g., Buy/Sell/Create) are solely for demonstration purposes.
// These are public examples and do not imply any affiliation with their origin, content, or platforms
// such as pump.fun. The authors of this test suite are not associated with the transactions shown.
package chainstream_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/gerasimovvladislav/zensol-go/chainstream"
)

func loadNotification(t *testing.T, file string) *chainstream.TransactionNotification {
	t.Helper()
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	var tx chainstream.TransactionNotification
	if err := json.Unmarshal(data, &tx); err != nil {
		t.Fatalf("failed to unmarshal tx: %v", err)
	}
	return &tx
}

func TestSlotMethod(t *testing.T) {
	tx := loadNotification(t, "testdata/sample_tx_buy.json")
	expected := uint64(330588464)

	if got := tx.Slot(); got != expected {
		t.Errorf("Slot() = %d, expected %d", got, expected)
	}
}

func TestSignatureMethod(t *testing.T) {
	tx := loadNotification(t, "testdata/sample_tx_buy.json")
	expected := "3w8agXbpQDUjrixUpojgs3nqVCvQ2cqNMaUc4te42rqtgdapHkKCADBukL8mJJMMrhsED59PqBPZtrPx8K1EdVWP"

	if got := tx.Signature(); got != expected {
		t.Errorf("Signature() = %q, expected %q", got, expected)
	}
}

func TestOwnerMethod(t *testing.T) {
	tx := loadNotification(t, "testdata/sample_tx_buy.json")
	expected := "53CkQzZiYAqwSdYRUX546ekKkNsKQCu9KTu9duvGZnhF"

	if got := tx.Owner(); got != expected {
		t.Errorf("Owner() = %q, expected %q", got, expected)
	}
}

func TestInstructionTypeMethod(t *testing.T) {
	tests := []struct {
		name     string
		jsonFile string
		expected string
	}{
		{"Buy Instruction", "testdata/sample_tx_buy.json", "Buy"},
		{"Sell Instruction", "testdata/sample_tx_sell.json", "Sell"},
		{"Create Log", "testdata/sample_tx_create.json", "Create"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data, err := os.ReadFile(tc.jsonFile)
			if err != nil {
				t.Fatalf("failed to read file: %v", err)
			}

			var tx chainstream.TransactionNotification
			if err := json.Unmarshal(data, &tx); err != nil {
				t.Fatalf("failed to unmarshal tx: %v", err)
			}

			if got := tx.InstructionType(); got != tc.expected {
				t.Errorf("InstructionType() = %q, expected %q", got, tc.expected)
			}
		})
	}
}

func TestTransactionNotificationParsing(t *testing.T) {
	tests := []struct {
		name     string
		jsonFile string
		expected string
	}{
		{"Buy Tx", "testdata/sample_tx_buy.json", "Buy"},
		{"Sell Tx", "testdata/sample_tx_sell.json", "Sell"},
		{"Create Tx", "testdata/sample_tx_create.json", "Create"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data, err := os.ReadFile(tc.jsonFile)
			if err != nil {
				t.Fatalf("failed to read file: %v", err)
			}

			var tx chainstream.TransactionNotification
			if err := json.Unmarshal(data, &tx); err != nil {
				t.Fatalf("failed to unmarshal tx: %v", err)
			}

			detected := tx.InstructionType()

			if detected != tc.expected {
				t.Errorf("expected instruction %q, got %q", tc.expected, detected)
			} else {
				t.Logf("✅ Detected instruction: %s", detected)
				t.Logf("🔑 Signature: %s", tx.Params.Result.Context.Signature)
				t.Logf("🧱 Slot: %d", tx.Params.Result.Value.Slot)
			}
		})
	}
}

func TestInvalidJSON(t *testing.T) {
	invalid := `{"jsonrpc": "2.0", "method": "transactionNotification", "params": {` // broken JSON
	var tx chainstream.TransactionNotification
	if err := json.Unmarshal([]byte(invalid), &tx); err == nil {
		t.Error("expected unmarshal error for invalid JSON")
	} else {
		t.Logf("✅ Correctly failed to unmarshal: %v", err)
	}
}

func TestMissingFields(t *testing.T) {
	minimal := `{"jsonrpc": "2.0", "method": "transactionNotification", "params": {"subscription": 1, "result": {}}}`
	var tx chainstream.TransactionNotification
	if err := json.Unmarshal([]byte(minimal), &tx); err != nil {
		t.Errorf("unexpected unmarshal error: %v", err)
	} else {
		t.Logf("🧪 Parsed with missing fields: %+v", tx)
	}
}
