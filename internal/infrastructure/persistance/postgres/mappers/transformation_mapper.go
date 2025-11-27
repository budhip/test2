package mappers

import (
	"encoding/json"

	"go-megatron/internal/domain/transformation"
	"go-megatron/internal/infrastructure/persistence/postgres"
)

func ToDomainWalletTransaction(dbModel postgres.DBWalletTransaction) (*transformation.WalletTransaction, error) {
	var metadata map[string]interface{}
	if err := json.Unmarshal(dbModel.Metadata, &metadata); err != nil {
		metadata = make(map[string]interface{})
	}

	return transformation.NewWalletTransaction(
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
		metadata,
	)
}

func ToDBWalletTransaction(wt *transformation.WalletTransaction) postgres.DBWalletTransaction {
	metadataJSON, _ := json.Marshal(wt.Metadata().ToMap())

	return postgres.DBWalletTransaction{
		ID:                       wt.ID().String(),
		Status:                   wt.Status().String(),
		AccountNumber:            wt.AccountNumber().String(),
		DestinationAccountNumber: wt.DestinationAccountNumber().String(),
		RefNumber:                wt.RefNumber().String(),
		TransactionType:          wt.TransactionType().String(),
		TransactionTime:          wt.TransactionTime(),
		TransactionFlow:          wt.TransactionFlow().String(),
		NetAmount:                wt.NetAmount().Value(),
		Currency:                 wt.NetAmount().Currency().Code(),
		Description:              wt.Description(),
		Metadata:                 metadataJSON,
		CreatedAt:                wt.CreatedAt(),
	}
}

func ToDomainTransaction(dbModel postgres.DBTransaction) (*transformation.Transaction, error) {
	var metadata map[string]interface{}
	if err := json.Unmarshal(dbModel.Metadata, &metadata); err != nil {
		metadata = make(map[string]interface{})
	}

	tx, err := transformation.NewTransaction(
		dbModel.ID,
		dbModel.FromAccount,
		dbModel.ToAccount,
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
		metadata,
	)
	if err != nil {
		return nil, err
	}

	tx.SetNarratives(dbModel.FromNarrative, dbModel.ToNarrative)

	return tx, nil
}

func ToDBTransaction(tx *transformation.Transaction) postgres.DBTransaction {
	metadataJSON, _ := json.Marshal(tx.Metadata().ToMap())

	return postgres.DBTransaction{
		ID:              tx.ID().String(),
		FromAccount:     tx.FromAccount().String(),
		ToAccount:       tx.ToAccount().String(),
		FromNarrative:   tx.FromNarrative(),
		ToNarrative:     tx.ToNarrative(),
		TransactionDate: tx.TransactionDate().Time(),
		Amount:          tx.Amount().Value(),
		Currency:        tx.Amount().Currency().Code(),
		Status:          tx.Status().String(),
		TransactionType: tx.TransactionType().String(),
		Description:     tx.Description(),
		RefNumber:       tx.RefNumber().String(),
		OrderType:       tx.OrderType().String(),
		OrderTime:       tx.OrderTime(),
		TransactionTime: tx.TransactionTime(),
		Metadata:        metadataJSON,
		CreatedAt:       tx.CreatedAt(),
	}
}
