// Package chainstream provides shared types for working with the Syndica ChainStream WebSocket API.
package chainstream

import "time"

// JSONRPCRequest is a generic JSON-RPC request payload.
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

// JSONRPCResponse is a generic JSON-RPC response payload.
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Result  interface{} `json:"result"`
	Error   *RPCError   `json:"error,omitempty"`
}

// RPCError represents an error returned by the JSON-RPC API.
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// AccountKeysFilter defines filtering rules based on Solana account keys.
type AccountKeysFilter struct {
	All     []string `json:"all,omitempty"`
	OneOf   []string `json:"oneOf,omitempty"`
	Exclude []string `json:"exclude,omitempty"`
}

// TransactionFilter contains optional filters for transaction subscriptions.
type TransactionFilter struct {
	ExcludeVotes bool               `json:"excludeVotes"`
	Commitment   string             `json:"commitment"`
	AccountKeys  *AccountKeysFilter `json:"accountKeys,omitempty"`
}

// TransactionSubscribeParams contains params for transactionsSubscribe.
type TransactionSubscribeParams struct {
	Network  string            `json:"network"`
	Verified bool              `json:"verified"`
	Filter   TransactionFilter `json:"filter"`
}

// BlockSubscribeParams contains params for blocksSubscribe.
type BlockSubscribeParams struct {
	Network  string `json:"network"`
	Verified bool   `json:"verified"`
}

// SlotSubscribeParams contains params for slotsSubscribe.
type SlotSubscribeParams struct {
	Network  string `json:"network"`
	Verified bool   `json:"verified"`
}

// ContextMetadata provides contextual information tied to a transaction/slot/block.
type ContextMetadata struct {
	Slot       uint64    `json:"slot"`
	SlotStatus string    `json:"slotStatus"`
	NodeTime   time.Time `json:"nodeTime"`
	IsVote     bool      `json:"isVote"`
	Signature  string    `json:"signature"`
	Index      int       `json:"index"`
}
