package chainstream

import (
	"context"
)

type Client interface {
	TransactionsNotifications(
		ctx context.Context,
		request *TransactionNotificationRequest,
		do func(notification *TransactionNotification),
	) error
}

type C struct {
	config *Config
}

func NewClient(config *Config) *C {
	return &C{
		config,
	}
}
