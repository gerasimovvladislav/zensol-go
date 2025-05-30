// Package chainstream defines the structures and subscription logic
// for handling transaction notifications from Syndica ChainStream WebSocket API.
package chainstream

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

// TransactionNotification represents a transaction update message.
type TransactionNotification struct {
	JSONRPC string                        `json:"jsonrpc"`
	Method  string                        `json:"method"`
	Params  TransactionNotificationParams `json:"params"`
}

// Slot returns the Solana slot in which the transaction was processed.
func (t *TransactionNotification) Slot() uint64 {
	return t.Params.Result.Value.Slot
}

// Signature returns the transaction signature from context.
func (t *TransactionNotification) Signature() string {
	return t.Params.Result.Context.Signature
}

// Owner returns the first account key in the transaction message (usually the transaction initiator).
func (t *TransactionNotification) Owner() string {
	if len(t.Params.Result.Value.Transaction.Message.AccountKeys) == 0 {
		return ""
	}
	return t.Params.Result.Value.Transaction.Message.AccountKeys[0]
}

// TransactionNotificationParams contains subscription ID and payload.
type TransactionNotificationParams struct {
	Subscription int64                       `json:"subscription"`
	Result       TransactionNotificationData `json:"result"`
}

// TransactionNotificationData holds the context and transaction data.
type TransactionNotificationData struct {
	Context ContextMetadata  `json:"context"`
	Value   TransactionValue `json:"value"`
}

// TransactionValue includes block metadata and full transaction info.
type TransactionValue struct {
	BlockTime   *int64             `json:"blockTime,omitempty"`
	Slot        uint64             `json:"slot"`
	Transaction EncodedTransaction `json:"transaction"`
	Meta        TransactionMeta    `json:"meta"`
}

// EncodedTransaction holds message and signature information.
type EncodedTransaction struct {
	Message     TransactionMessage `json:"message"`
	MessageHash string             `json:"messageHash"`
	Signatures  []string           `json:"signatures"`
}

// TransactionMessage describes a parsed transaction message.
type TransactionMessage struct {
	AccountKeys         []string              `json:"accountKeys"`
	AddressTableLookups []AddressTableLookup  `json:"addressTableLookups"`
	Header              MessageHeader         `json:"header"`
	Instructions        []CompiledInstruction `json:"instructions"`
	RecentBlockhash     string                `json:"recentBlockhash"`
}

// MessageHeader defines account signing/readonly metadata.
type MessageHeader struct {
	NumReadonlySigned   int `json:"numReadonlySignedAccounts"`
	NumReadonlyUnsigned int `json:"numReadonlyUnsignedAccounts"`
	NumSignatures       int `json:"numRequiredSignatures"`
}

// AddressTableLookup maps short account references to actual pubkeys.
type AddressTableLookup struct {
	AccountKey      string `json:"accountKey"`
	WritableIndexes []int  `json:"writableIndexes"`
	ReadonlyIndexes []int  `json:"readonlyIndexes"`
}

// CompiledInstruction describes a single instruction in compiled form.
type CompiledInstruction struct {
	ProgramIDIndex int    `json:"programIdIndex"`
	Accounts       []int  `json:"accounts"`
	Data           string `json:"data"`
}

// TransactionMeta describes post-transaction data: logs, balance diffs, etc.
type TransactionMeta struct {
	Err               json.RawMessage    `json:"err"`
	Fee               uint64             `json:"fee"`
	InnerInstructions []InnerInstruction `json:"innerInstructions"`
	LoadedAddresses   LoadedAddresses    `json:"loadedAddresses"`
	LogMessages       []string           `json:"logMessages"`
	PostBalances      []uint64           `json:"postBalances"`
	PostTokenBalances []TokenBalance     `json:"postTokenBalances"`
	PreBalances       []uint64           `json:"preBalances"`
	PreTokenBalances  []TokenBalance     `json:"preTokenBalances"`
	Rewards           []interface{}      `json:"rewards,omitempty"`
}

// InnerInstruction represents an instruction executed inside another.
type InnerInstruction struct {
	Index        int                   `json:"index"`
	Instructions []CompiledInstruction `json:"instructions"`
}

// LoadedAddresses lists additional accounts loaded dynamically.
type LoadedAddresses struct {
	Writable []string `json:"writable"`
	Readonly []string `json:"readonly"`
}

// TokenBalance describes a token balance snapshot before/after the tx.
type TokenBalance struct {
	AccountIndex int           `json:"accountIndex"`
	Mint         string        `json:"mint"`
	Owner        string        `json:"owner"`
	ProgramID    string        `json:"programId"`
	UIAmount     TokenAmountUI `json:"uiTokenAmount"`
}

// TokenAmountUI represents token amount in a UI-friendly format.
type TokenAmountUI struct {
	Amount         string  `json:"amount"`
	Decimals       int     `json:"decimals"`
	UIAmount       float64 `json:"uiAmount"`
	UIAmountString string  `json:"uiAmountString"`
}

// TransactionsNotifications subscribes to Syndica transaction updates.
func (c *C) TransactionsNotifications(
	ctx context.Context,
	request *JSONRPCRequest,
	do func(notification *TransactionNotification),
) error {
RECONNECT:
	wsConn, _, err := websocket.Dial(ctx, c.config.WssApiEndpoint, nil)
	if err != nil {
		return fmt.Errorf("cannot connect to chainstream transactions notifications: %w", err)
	}
	defer func() {
		_ = wsConn.Close(websocket.StatusNormalClosure, "subscription of transactions notifications was closed")
	}()

	if err = wsjson.Write(ctx, wsConn, request); err != nil {
		return fmt.Errorf("cannot send subscribe transactions: %w", err)
	}

	var subResp JSONRPCResponse
	if err = wsjson.Read(ctx, wsConn, &subResp); err != nil {
		return fmt.Errorf("cannot read subscribe response: %w", err)
	}
	if subResp.Result == 0 {
		return fmt.Errorf("subscribe error: result is nil")
	}

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			_ = wsConn.Ping(ctx)
		case <-ctx.Done():
			return nil
		default:
			var notification TransactionNotification
			if err = wsjson.Read(ctx, wsConn, &notification); err != nil {
				if errors.Is(err, context.Canceled) || ctx.Err() != nil {
					return nil
				}
				time.Sleep(time.Second)
				goto RECONNECT
			}
			do(&notification)
		}
	}
}
