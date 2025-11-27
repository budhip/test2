package transformation

import "context"

type TransformationRepository interface {
	SaveWalletTransaction(ctx context.Context, wt *WalletTransaction) error
	SaveTransactions(ctx context.Context, transactions []*Transaction) error

	FindWalletTransactionByID(ctx context.Context, id WalletTransactionID) (*WalletTransaction, error)
	FindTransactionsByWalletID(ctx context.Context, walletID WalletTransactionID) ([]*Transaction, error)
}
