package chainstream

type SubscribeResponse struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Result  int    `json:"result"` // subscription ID
}
