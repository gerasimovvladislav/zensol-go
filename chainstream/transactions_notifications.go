package chainstream

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gagliardetto/solana-go"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type TransactionNotificationRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Method  string `json:"method"`
	Params  struct {
		Network  string `json:"network"`
		Verified bool   `json:"verified"`
		Filter   struct {
			ExcludeVotes bool   `json:"excludeVotes"`
			Commitment   string `json:"commitment"`
			AccountKeys  struct {
				All     []string `json:"all,omitempty"`
				OneOf   []string `json:"oneOf,omitempty"`
				Exclude []string `json:"exclude,omitempty"`
			} `json:"accountKeys"`
		} `json:"filter"`
	} `json:"params"`
}

type TransactionNotification struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  struct {
		Subscription int64 `json:"subscription"`
		Result       struct {
			Context struct {
				SlotStatus string    `json:"slotStatus"`
				NodeTime   time.Time `json:"nodeTime"`
				IsVote     bool      `json:"isVote"`
				Signature  string    `json:"signature"`
				Index      int       `json:"index"`
			} `json:"context"`
			Value struct {
				BlockTime interface{} `json:"blockTime"`
				Meta      struct {
					Err               interface{} `json:"err"`
					Fee               int         `json:"fee"`
					InnerInstructions []struct {
						Index        int `json:"index"`
						Instructions []struct {
							ProgramIdIndex int    `json:"programIdIndex"`
							Accounts       []int  `json:"accounts"`
							Data           string `json:"data"`
						} `json:"instructions"`
					} `json:"innerInstructions"`
					LoadedAddresses struct {
						Writable []interface{} `json:"writable"`
						Readonly []string      `json:"readonly"`
					} `json:"loadedAddresses"`
					LogMessages       []string `json:"logMessages"`
					PostBalances      []int64  `json:"postBalances"`
					PostTokenBalances []struct {
						AccountIndex  int    `json:"accountIndex"`
						Mint          string `json:"mint"`
						Owner         string `json:"owner"`
						ProgramId     string `json:"programId"`
						UiTokenAmount struct {
							Amount         string  `json:"amount"`
							Decimals       int     `json:"decimals"`
							UiAmount       float64 `json:"uiAmount"`
							UiAmountString string  `json:"uiAmountString"`
						} `json:"uiTokenAmount"`
					} `json:"postTokenBalances"`
					PreBalances      []int64 `json:"preBalances"`
					PreTokenBalances []struct {
						AccountIndex  int    `json:"accountIndex"`
						Mint          string `json:"mint"`
						Owner         string `json:"owner"`
						ProgramId     string `json:"programId"`
						UiTokenAmount struct {
							Amount         string  `json:"amount"`
							Decimals       int     `json:"decimals"`
							UiAmount       float64 `json:"uiAmount"`
							UiAmountString string  `json:"uiAmountString"`
						} `json:"uiTokenAmount"`
					} `json:"preTokenBalances"`
					Rewards []interface{} `json:"rewards"`
				} `json:"meta"`
				Slot        int                 `json:"slot"`
				Transaction *solana.Transaction `json:"transaction"`
			} `json:"value"`
		} `json:"result"`
	} `json:"params"`
}

func (c *C) TransactionsNotifications(
	ctx context.Context,
	request *TransactionNotificationRequest,
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

	var subResp SubscribeResponse
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
				// graceful context cancel
				if errors.Is(err, context.Canceled) || ctx.Err() != nil {
					return nil
				}

				// temporary issue — try reconnecting
				time.Sleep(1 * time.Second)

				goto RECONNECT
			}

			do(&notification)
		}
	}
}
