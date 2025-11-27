package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"go-megatron/internal/domain/transformation"
	"go-megatron/internal/infrastructure/persistence/postgres/mappers"
)

type TransformationRepository struct {
	db *sql.DB
}

func NewTransformationRepository(db *sql.DB) *TransformationRepository {
	return &TransformationRepository{db: db}
}

func (r *TransformationRepository) SaveWalletTransaction(ctx context.Context, wt *transformation.WalletTransaction) error {
	dbModel := mappers.ToDBWalletTransaction(wt)

	query := `
		INSERT INTO wallet_transactions (
			id, status, account_number, destination_account_number, ref_number,
			transaction_type, transaction_time, transaction_flow,
			net_amount, currency, description, metadata, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		dbModel.ID,
		dbModel.Status,
		dbModel.AccountNumber,
		dbModel.DestinationAccountNumber,
		dbModel.RefNumber,
		dbModel.TransactionType,
		dbModel.TransactionTime,
		dbModel.TransactionFlow,
		dbModel.NetAmount,
		dbModel.Currency,
		dbModel.Description,
		dbModel.Metadata,
		dbModel.CreatedAt,
	)

	return err
}

func (r *TransformationRepository) SaveTransactions(ctx context.Context, transactions []*transformation.Transaction) error {
	if len(transactions) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO transactions (
			id, from_account, to_account, from_narrative, to_narrative,
			transaction_date, amount, currency, status, transaction_type,
			description, ref_number, order_type, order_time, transaction_time,
			metadata, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, txEntity := range transactions {
		dbModel := mappers.ToDBTransaction(txEntity)

		_, err := stmt.ExecContext(
			ctx,
			dbModel.ID,
			dbModel.FromAccount,
			dbModel.ToAccount,
			dbModel.FromNarrative,
			dbModel.ToNarrative,
			dbModel.TransactionDate,
			dbModel.Amount,
			dbModel.Currency,
			dbModel.Status,
			dbModel.TransactionType,
			dbModel.Description,
			dbModel.RefNumber,
			dbModel.OrderType,
			dbModel.OrderTime,
			dbModel.TransactionTime,
			dbModel.Metadata,
			dbModel.CreatedAt,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *TransformationRepository) FindWalletTransactionByID(ctx context.Context, id transformation.WalletTransactionID) (*transformation.WalletTransaction, error) {
	query := `
		SELECT id, status, account_number, destination_account_number, ref_number,
			   transaction_type, transaction_time, transaction_flow,
			   net_amount, currency, description, metadata, created_at
		FROM wallet_transactions
		WHERE id = $1
	`

	var dbModel DBWalletTransaction
	err := r.db.QueryRowContext(ctx, query, id.String()).Scan(
		&dbModel.ID,
		&dbModel.Status,
		&dbModel.AccountNumber,
		&dbModel.DestinationAccountNumber,
		&dbModel.RefNumber,
		&dbModel.TransactionType,
		&dbModel.TransactionTime,
		&dbModel.TransactionFlow,
		&dbModel.NetAmount,
		&dbModel.Currency,
		&dbModel.Description,
		&dbModel.Metadata,
		&dbModel.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("wallet transaction not found")
	}
	if err != nil {
		return nil, err
	}

	return mappers.ToDomainWalletTransaction(dbModel)
}

func (r *TransformationRepository) FindTransactionsByWalletID(ctx context.Context, walletID transformation.WalletTransactionID) ([]*transformation.Transaction, error) {
	query := `
		SELECT id, from_account, to_account, from_narrative, to_narrative,
			   transaction_date, amount, currency, status, transaction_type,
			   description, ref_number, order_type, order_time, transaction_time,
			   metadata, created_at
		FROM transactions
		WHERE wallet_transaction_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, walletID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*transformation.Transaction

	for rows.Next() {
		var dbModel DBTransaction
		err := rows.Scan(
			&dbModel.ID,
			&dbModel.FromAccount,
			&dbModel.ToAccount,
			&dbModel.FromNarrative,
			&dbModel.ToNarrative,
			&dbModel.TransactionDate,
			&dbModel.Amount,
			&dbModel.Currency,
			&dbModel.Status,
			&dbModel.TransactionType,
			&dbModel.Description,
			&dbModel.RefNumber,
			&dbModel.OrderType,
			&dbModel.OrderTime,
			&dbModel.TransactionTime,
			&dbModel.Metadata,
			&dbModel.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		tx, err := mappers.ToDomainTransaction(dbModel)
		if err != nil {
			return nil, err
		}

		transactions = append(transactions, tx)
	}

	return transactions, rows.Err()
}
