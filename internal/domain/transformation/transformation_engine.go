package transformation

import "context"

type TransformationEngine interface {
	Transform(ctx context.Context, wt *WalletTransaction, ruleContent string) ([]*Transaction, error)
}
