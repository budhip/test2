package rules

import (
	"fmt"
	"strings"

	"bitbucket.org/Amartha/go-megatron/internal/pkg/accounting"

	"github.com/hashicorp/go-multierror"
)

var (
	mapEntityCode = map[string]string{
		"amf": "001",
		"anr": "002",
		"afa": "003",
		"MMF": "004",
		"awf": "005",
	}
)

func (t *pasTransformed) GetAccountOrderTypeEntityProduct(entityCode, productType string) string {
	orderType := strings.ToLower(string(t.Payload.OrderType))
	entityCode = t.GetEntity(strings.ToLower(entityCode))
	productType = strings.ToLower(productType)
	key := fmt.Sprintf("%s_%s", entityCode, orderType)
	return t.cfg.ConfigAccount.MapAccountOrderTypesEntityProduct[key][productType]
}

func (t *pasTransformed) GetAccountByTrxTypeEntityProduct(trxType, entityCode, productType string) string {
	//using first 3 characters from trx type as order type
	orderType := strings.ToLower(trxType[:3])
	entityCode = t.GetEntity(strings.ToLower(entityCode))
	productType = strings.ToLower(productType)
	key := fmt.Sprintf("%s_%s", entityCode, orderType)
	return t.cfg.ConfigAccount.MapAccountOrderTypesEntityProduct[key][productType]
}

func (t *pasTransformed) GetEntity(entityCode string) string {
	value, exist := mapEntityCode[entityCode]
	if !exist {
		return entityCode
	}
	return value
}

func (t *pasTransformed) GetDescriptionInFirstIndex(description string) string {
	index := strings.Index(description, " ")
	if index != -1 {
		return description[:index]
	}
	return ""
}

func (t *pasTransformed) GetDescriptionInSecondIndex(description string) string {
	index := strings.Index(description, " ")
	if index != -1 {
		return description[index+1:]
	}
	return ""
}

func (t *pasTransformed) GetAccountsByAccountNumber(accountNumber string) (res accounting.GetAllAccountNumbersByParam) {
	input := accounting.DoGetAllAccountNumbersByParamRequest{
		AccountNumbers: accountNumber,
	}

	resp, err := t.accountingClient.GetAccountsByParams(t.ctx, input)
	if err != nil {
		t.prePublishErrors = multierror.Append(t.prePublishErrors, err)
		return
	}

	account, exist := resp[accountNumber]
	if !exist || len(account) == 0 {
		t.prePublishErrors = multierror.Append(t.prePublishErrors, fmt.Errorf("account number %s not found", accountNumber))
		return
	}

	return account[0]
}

// MultiGetAccountByEntityProductInMetaAndDescription1
func (t *pasTransformed) MultiGetAccountByEntityProductInMetaAndDescription1(acuan *AcuanTransaction) (value string) {
	if value = t.MultiGetAccountByEntityProductInMeta(acuan); value != "" {
		return value
	}

	entityCode := acuan.GetEntityCodeInMeta()
	productType := t.GetDescriptionInFirstIndex(acuan.Description)
	if value = t.GetAccountOrderTypeEntityProduct(entityCode, productType); value != "" {
		return value
	}

	return
}

// MultiGetAccountByEntityProductInMetaAndDescription2
func (t *pasTransformed) MultiGetAccountByEntityProductInMetaAndDescription2(acuan *AcuanTransaction) (value string) {
	if value = t.MultiGetAccountByEntityProductInMeta(acuan); value != "" {
		return value
	}

	entityCode := acuan.GetEntityCodeInMeta()
	productType := t.GetDescriptionInSecondIndex(acuan.Description)
	if value = t.GetAccountOrderTypeEntityProduct(entityCode, productType); value != "" {
		return value
	}

	return
}

// MultiGetAccountByEntityProductInMeta get account based on entity and product in the meta.
func (t *pasTransformed) MultiGetAccountByEntityProductInMeta(acuan *AcuanTransaction) string {
	entityCode := acuan.GetEntityCodeInMeta()
	productType := acuan.GetProductType()
	return t.GetAccountOrderTypeEntityProduct(entityCode, productType)
}

// MultiGetAccountByDescriptionAndAccountNumber get account based on description and account number to get entity code.
func (t *pasTransformed) MultiGetAccountByDescriptionAndAccountNumber(productDesc, accountNumber string) string {
	account := t.GetAccountsByAccountNumber(accountNumber)
	return t.GetAccountOrderTypeEntityProduct(account.EntityCode, productDesc)
}

// MultiGetAccountByEntityInMeta get account when the data provided is only an entity in the meta then the product type will be set as default.
func (t *pasTransformed) MultiGetAccountByEntityInMeta(acuan *AcuanTransaction) string {
	entityCode := acuan.GetEntityCodeInMeta()

	//for now, only used for ADMFE, which is a child transaction with different order type
	return t.GetAccountByTrxTypeEntityProduct(string(acuan.TransactionType), entityCode, "default")
}

func (t *pasTransformed) MultiGetAccountByAccEntityAndTrxTypes(accountNumber, trxType, debitCredit string) string {
	account := t.GetAccountsByAccountNumber(accountNumber)
	entityCode := t.GetEntity(account.EntityCode)
	key := fmt.Sprintf("%s_%s", entityCode, strings.ToLower(trxType))
	return t.cfg.ConfigAccount.MapAccountEntityTrxTypes[key][debitCredit]
}

func (t *pasTransformed) MultiGetAccountByEntityProductTypeTrxType(accountNumber, entity, trxType, debitCredit string) string {
	account := t.GetAccountsByAccountNumber(accountNumber)
	key := fmt.Sprintf("%s_%s_%s", entity, account.ProductTypeCode, strings.ToLower(trxType))
	return t.cfg.ConfigAccount.MapAccountEntityTrxTypes[key][debitCredit]
}

func (t *pasTransformed) GetRepaymentCashInTransitByEntity(entity string) string {
	entityCode := t.GetEntity(strings.ToLower(entity))
	return t.cfg.ConfigAccount.MapRepaymentCashInTransitEntity[entityCode]
}

func (t *pasTransformed) GetBankHOPoolingRepaymentByEntity(entity string) string {
	entityCode := t.GetEntity(strings.ToLower(entity))
	return t.cfg.ConfigAccount.MapBankHORepaymentPooling[entityCode]
}

// GetAccountByEntityDestination mendapatkan account number untuk DSBPI berdasarkan entity dari destination account
func (t *pasTransformed) GetAccountByEntityDestination(destinationAccountId string) string {
	// Get account detail untuk mendapatkan entity code
	account := t.GetAccountsByAccountNumber(destinationAccountId)
	if account.AccountNumber == "" {
		t.prePublishErrors = multierror.Append(t.prePublishErrors,
			fmt.Errorf("account not found for destination account id: %s", destinationAccountId))
		return ""
	}

	entityCode := t.GetEntity(strings.ToLower(account.EntityCode))

	return t.cfg.ConfigAccount.MapDSBPIAccountEntity[entityCode]
}
