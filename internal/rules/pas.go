package rules

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"bitbucket.org/Amartha/go-megatron/internal/config"
	"bitbucket.org/Amartha/go-megatron/internal/pkg/accounting"
	"bitbucket.org/Amartha/go-megatron/internal/pkg/dddnotification"
	"bitbucket.org/Amartha/go-megatron/internal/pkg/flag"
	"bitbucket.org/Amartha/go-megatron/internal/pkg/kafka"

	goAcuanLibModel "bitbucket.org/Amartha/go-acuan-lib/model"
	paspkg "bitbucket.org/Amartha/go-megatron/internal/pkg/pas"
	xlog "bitbucket.org/Amartha/go-x/log"

	"github.com/google/uuid"
	"github.com/hashicorp/go-multierror"
	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/shopspring/decimal"
)

var (
	TimezoneJakarta            = "Asia/Jakarta"
	DateFormatYYYYMMDDWithTime = "2006-01-02 15:04:05"

	AccountMandiriAcuanDev = "114001000000012"
	AccountMandiriPASDev   = "144001000000003"
	AccountBCAAcuanDev     = "114001000000014"
	AccountBCAPASDev       = "144001000000004"
	AccountBRIAcuanDev     = "114001000000015"
	AccountBRIPASDev       = "144001000000005"
	AccountBRIBPEAcuanDev  = "114001000000011"
	AccountBRIBPEPASDev    = "144001000000002"
	ToAccountBankPASDev    = map[string]string{
		AccountMandiriAcuanDev: AccountMandiriPASDev,
		AccountBCAAcuanDev:     AccountBCAPASDev,
		AccountBRIAcuanDev:     AccountBRIPASDev,
		AccountBRIBPEAcuanDev:  AccountBRIBPEPASDev,
	}

	AccountMandiriAcuanProd = "114001000000001"
	AccountMandiriPASProd   = "144001000000003"
	AccountBCAAcuanProd     = "114001000000020"
	AccountBCAPASProd       = "144001000000004"
	AccountBRIAcuanProd     = "114001000000014"
	AccountBRIPASProd       = "144001000000002"
	AccountBRIBPEAcuanProd  = "114001000000009"
	AccountBRIBPEPASProd    = "114001000000009"
	ToAccountBankPASProd    = map[string]string{
		AccountMandiriAcuanProd: AccountMandiriPASProd,
		AccountBCAAcuanProd:     AccountBCAPASProd,
		AccountBRIAcuanProd:     AccountBRIPASProd,
		AccountBRIBPEAcuanProd:  AccountBRIBPEPASProd,
	}

	EntityAMF = "AMF"
	EntityAFA = "AFA"

	AccountByEntityDebitPASDev = map[string]string{
		EntityAMF: "141003000000002",
		EntityAFA: "141003000000002",
	}
	AccountByEntityCreditPASDev = map[string]string{
		EntityAMF: "114001000000011",
		EntityAFA: "114003000000016",
	}

	AccountByEntityDebitPASUat = map[string]string{
		EntityAMF: "141003000000001",
		EntityAFA: "141003000000002",
	}
	AccountByEntityCreditPASUat = map[string]string{
		EntityAMF: "114001000000002",
		EntityAFA: "114003000000001",
	}

	AccountByEntityDebitPASProd = map[string]string{
		EntityAMF: "IDR1104100011000",
		EntityAFA: "141003000000001",
	}
	AccountByEntityCreditPASProd = map[string]string{
		EntityAMF: "114001000000009",
		EntityAFA: "230001713145",
	}

	EntityCodeAMF = "001"
	EntityCodeAFA = "003"

	AccountByEntityCodeCreditPASDev = map[string]string{
		EntityCodeAMF: "144001000000002",
		EntityCodeAFA: "114003000000016",
	}

	AccountByEntityCodeCreditPASUat = map[string]string{
		EntityCodeAMF: "114001000000002",
		EntityCodeAFA: "114003000000001",
	}

	AccountByEntityCodeCreditPASProd = map[string]string{
		EntityCodeAMF: "114001000000009",
		EntityCodeAFA: "230001713145",
	}
)

type pas struct {
	rule
	publisher          kafka.Publisher
	accountingClient   accounting.Client
	notificationClient dddnotification.Client
	flagClient         flag.Client
	cfg                *config.Configuration
}

type pasPublisher struct {
	ctx                context.Context
	publisher          kafka.Publisher
	accountingClient   accounting.Client
	notificationClient dddnotification.Client

	// errPublish is error information when publishing to kafka
	errPublish error
}

// pasTransformed is struct to be used as data context in rule engine
// it will be used to transform acuan transaction to journal request
type (
	pasTransformed struct {
		cfg *config.Configuration

		pasPublisher
		Payload Payload
		timeLoc *time.Location

		prePublishErrors *multierror.Error

		// IsReadyToPublish is flag to determine if transaction is ready to publish to journal_stream
		IsReadyToPublish bool
	}

	Payload struct {
		Type            string                    `json:"type"`
		ReferenceNumber string                    `json:"referenceNumber"`
		TransactionID   *uuid.UUID                `json:"transactionId"`
		OrderType       goAcuanLibModel.OrderType `json:"orderType"`
		TransactionDate string                    `json:"transactionDate"`
		ProcessingDate  string                    `json:"processingDate"`
		Currency        string                    `json:"currency"`
		Transactions    []*JournalTransaction     `json:"transactions"`
		Metadata        interface{}               `json:"metadata"`
	}
	JournalTransaction struct {
		TransactionType     goAcuanLibModel.TransactionType `json:"transactionType"`
		TransactionTypeName string                          `json:"transactionTypeName"`
		Account             string                          `json:"account"`
		Narrative           string                          `json:"narrative"`
		Amount              decimal.Decimal                 `json:"amount"`
		IsDebit             bool                            `json:"isDebit"`
	}

	AcuanTransaction struct {
		Identifier           string
		AcuanStatus          string
		OrderTime            goAcuanLibModel.AcuanTime         `json:"orderTime"`
		OrderType            goAcuanLibModel.OrderType         `json:"orderType"`
		RefNumber            string                            `json:"refNumber"`
		Id                   *uuid.UUID                        `json:"id"`
		Amount               decimal.Decimal                   `json:"amount"`
		Currency             string                            `json:"currency"`
		SourceAccountId      string                            `json:"sourceAccountId"`
		DestinationAccountId string                            `json:"destinationAccountId"`
		Description          string                            `json:"description"`
		Method               goAcuanLibModel.TransactionMethod `json:"method"`
		TransactionType      goAcuanLibModel.TransactionType   `json:"transactionType"`
		TransactionTypeName  string                            `json:"transactionTypeName"`
		TransactionTime      string                            `json:"transactionTime"`
		Status               goAcuanLibModel.TransactionStatus `json:"status"`
		Meta                 map[string]interface{}            `json:"meta"`

		AccountBalances   map[string]paspkg.AccountBalance `json:"accountBalances"`
		WalletTransaction paspkg.WalletTransaction         `json:"WalletTransaction"`
	}
)

func (a *AcuanTransaction) GetAccountNumber() string {
	return fmt.Sprint(a.Meta["accountNumber"])
}

func (a *AcuanTransaction) GetEntityInMeta() string {
	return fmt.Sprint(a.Meta["entity"])
}

func (a *AcuanTransaction) GetEntityCodeInMeta() string {
	if entity, exists := a.Meta["entity"]; exists {
		return fmt.Sprint(entity)
	}
	return fmt.Sprint(a.Meta["entityCode"])
}

func (a *AcuanTransaction) GetLoanTypeInMeta() string {
	return fmt.Sprint(a.Meta["loanType"])
}

func (a *AcuanTransaction) GetDebitAccountInMeta() string {
	return fmt.Sprint(a.Meta["debit_account"])
}

func (a *AcuanTransaction) GetSourceBankAccountInMeta() string {
	if v, ok := a.Meta["sourceBankAccount"]; ok && v != nil {
		return fmt.Sprint(v)
	}
	return ""
}

func (a *AcuanTransaction) GetNotesInMeta() string {
	return fmt.Sprint(a.Meta["notes"])
}

func (a *AcuanTransaction) GetDebitInMeta() string {
	if v, ok := a.Meta["debit"]; ok && v != nil {
		return fmt.Sprint(v)
	}
	return ""
}

func (a *AcuanTransaction) GetDebit2InMeta() string {
	if v, ok := a.Meta["debit2"]; ok && v != nil {
		return fmt.Sprint(v)
	}
	return ""
}

func (a *AcuanTransaction) GetCreditAccountInMeta() string {
	return fmt.Sprint(a.Meta["credit_account"])
}

func (a *AcuanTransaction) GetCreditInMeta() string {
	if v, ok := a.Meta["credit"]; ok && v != nil {
		return fmt.Sprint(v)
	}
	return ""
}

func (a *AcuanTransaction) GetCredit2InMeta() string {
	if v, ok := a.Meta["credit2"]; ok && v != nil {
		return fmt.Sprint(v)
	}
	return ""
}

func (a *AcuanTransaction) GetLoanIdInMeta() string {
	return fmt.Sprint(a.Meta["loanId"])
}

func (a *AcuanTransaction) GetLoanIDsInMeta() string {
	loanIdsInterface, ok := a.Meta["loanIds"].([]interface{})
	if !ok {
		return ""
	}

	var loanIds []string

	for _, v := range loanIdsInterface {
		if str, ok := v.(string); ok {
			loanIds = append(loanIds, str)
		} else {
			xlog.Info(context.Background(), "loanIds contains a non-string value")
		}
	}

	result := strings.Join(loanIds, ", ")

	return result
}

func (a *AcuanTransaction) GetDebitAccount() string {
	return fmt.Sprint(a.Meta["debitAccount"])
}

func (a *AcuanTransaction) GetProductType() string {
	return strings.ToLower(fmt.Sprint(a.Meta["productType"]))
}

func (a *AcuanTransaction) GetDescriptionInUniformCase() string {
	return strings.ToLower(fmt.Sprint(a.Description))
}

func (a *AcuanTransaction) GetLoanAccountNumber() string {
	if v, ok := a.Meta["loanAccountNumber"]; ok && v != nil {
		return fmt.Sprint(v)
	}
	return ""
}

func (a *AcuanTransaction) GetLoanAccountNumberModal() string {
	return fmt.Sprint(a.Meta["loanAccountNumberModal"])
}

func (a *AcuanTransaction) GetRepaymentDate() string {
	return fmt.Sprint(a.Meta["repaymentDate"])
}

func (a *AcuanTransaction) GetPaymentDate() string {
	return fmt.Sprint(a.Meta["paymentDate"])
}

func (a *AcuanTransaction) GetVAPoint() string {
	return fmt.Sprint(a.Meta["virtualAccountPoint"])
}

func (a *AcuanTransaction) GetDisbursementDate() string {
	return fmt.Sprint(a.Meta["disbursementDate"])
}

func (a *AcuanTransaction) GetVoucherCode() string {
	return fmt.Sprint(a.Meta["voucherCode"])
}

func (a *AcuanTransaction) GetNewLoanAccountNumber() string {
	return fmt.Sprint(a.Meta["newLoanAccountNumber"])
}

func (a *AcuanTransaction) GetOldLoanAccountNumber() string {
	return fmt.Sprint(a.Meta["oldLoanAccountNumber"])
}

func (a *AcuanTransaction) GetCustomerNumber() string {
	return fmt.Sprint(a.Meta["customerNumber"])
}

func (a *AcuanTransaction) GetDebitAcctNo() string {
	return fmt.Sprint(a.Meta["debit_acct_no"])
}

func (a *AcuanTransaction) GetCreditAcctNo() string {
	return fmt.Sprint(a.Meta["credit_acct_no"])
}

func (a *AcuanTransaction) GetNewEntityLoanAccountNumber() string {
	return fmt.Sprint(a.Meta["newEntityLoanAccountNumber"])
}

func (a *AcuanTransaction) GetOldEntityLoanAccountNumber() string {
	return fmt.Sprint(a.Meta["oldEntityLoanAccountNumber"])
}

func (a *AcuanTransaction) GetLoanKind() string {
	if v, ok := a.Meta["loanKind"]; ok && v != nil {
		return fmt.Sprint(v)
	}
	return ""
}

func (a *AcuanTransaction) GetPartnerId() string {
	if v, ok := a.Meta["partnerId"]; ok && v != nil {
		return fmt.Sprint(v)
	}
	return ""
}

func (a *AcuanTransaction) GetAccountNumberBank() string {
	if v, ok := a.Meta["accountNumberBank"]; ok && v != nil {
		return fmt.Sprint(v)
	}
	return ""
}

func (a *AcuanTransaction) GetAccountBankPASDev(accountAcuan string) string {
	return ToAccountBankPASDev[accountAcuan]
}

func (a *AcuanTransaction) GetAccountBankPASProd(accountAcuan string) string {
	return ToAccountBankPASProd[accountAcuan]
}

func (a *AcuanTransaction) GetAccountByEntityDebitPASDev() string {
	return AccountByEntityDebitPASDev[a.GetEntityInMeta()]
}

func (a *AcuanTransaction) GetAccountByEntityCreditPASDev() string {
	return AccountByEntityCreditPASDev[a.GetEntityInMeta()]
}

func (a *AcuanTransaction) GetAccountByEntityDebitPASProd() string {
	return AccountByEntityDebitPASProd[a.GetEntityInMeta()]
}

func (a *AcuanTransaction) GetAccountByEntityCreditPASProd() string {
	return AccountByEntityCreditPASProd[a.GetEntityInMeta()]
}

func (a *AcuanTransaction) GetAccountByEntityDebitPASUat() string {
	return AccountByEntityDebitPASUat[a.GetEntityInMeta()]
}

func (a *AcuanTransaction) GetAccountByEntityCreditPASUat() string {
	return AccountByEntityCreditPASUat[a.GetEntityInMeta()]
}

func (a *AcuanTransaction) GetAccountByDescEntityCodeDev() string {
	descSep := strings.Split(a.Description, "-")
	return AccountByEntityCodeCreditPASDev[strings.TrimSpace(descSep[len(descSep)-1])]
}

func (a *AcuanTransaction) GetAccountByDescEntityCodeUat() string {
	descSep := strings.Split(a.Description, "-")
	return AccountByEntityCodeCreditPASUat[strings.TrimSpace(descSep[len(descSep)-1])]
}

func (a *AcuanTransaction) GetAccountByDescEntityCodeProd() string {
	descSep := strings.Split(a.Description, "-")
	return AccountByEntityCodeCreditPASProd[strings.TrimSpace(descSep[len(descSep)-1])]
}

func (a *AcuanTransaction) GetAmountAndAdminFeeInMeta() (decimal.Decimal, decimal.Decimal, decimal.Decimal) {
	topupAmount, err := decimal.NewFromString(fmt.Sprintf("%v", a.Meta["amount"]))
	if err != nil {
		xlog.Warn(context.Background(), "GetAmountAndAdminFeeInMeta", xlog.Err(err))
	}
	adminFee, err := decimal.NewFromString(fmt.Sprintf("%v", a.Meta["adminFee"]))
	if err != nil {
		xlog.Warn(context.Background(), "GetAmountAndAdminFeeInMeta", xlog.Err(err))
	}

	netAmount := topupAmount.Sub(adminFee)

	return netAmount, topupAmount, adminFee
}

func (a *AcuanTransaction) GetNetAmount() decimal.Decimal {
	netAmount, _, _ := a.GetAmountAndAdminFeeInMeta()
	return netAmount
}

func (a *AcuanTransaction) GetPaymentMethodInMeta() string {
	return fmt.Sprint(a.Meta["paymentMethod"])
}

func (a *AcuanTransaction) GetVANumber() string {
	return fmt.Sprint(a.Meta["virtualAccountNumber"])
}

func (a *AcuanTransaction) HandleTopupScenario() (int, []decimal.Decimal, []decimal.Decimal) {
	var (
		topupAmount, adminFee, beforeBalance decimal.Decimal
		scenario                             int
		amounts                              []decimal.Decimal
		adminFees                            []decimal.Decimal
	)

	_, ok := a.Meta["amount"]
	if ok {
		beforeBalance = a.AccountBalances[a.GetAccountNumber()].Before.ActualBalance.Value
		_, topupAmount, adminFee = a.GetAmountAndAdminFeeInMeta()

	} else {
		beforeBalance = a.AccountBalances[a.WalletTransaction.AccountNumber].Before.ActualBalance.Value
		topupAmount = a.WalletTransaction.NetAmount.Value
	}

	netAmount := topupAmount.Sub(adminFee)

	switch true {
	case beforeBalance.IsNegative() && (topupAmount.Add(beforeBalance).IsNegative() || topupAmount.Add(beforeBalance).IsZero()):
		scenario = 2
		amounts = append(amounts, netAmount)

	case beforeBalance.IsNegative() && topupAmount.Add(beforeBalance).IsPositive():
		scenario = 3
		afterBalance := netAmount.Add(beforeBalance)
		amounts = append(amounts, beforeBalance.Abs(), afterBalance)

	default:
		scenario = 1
		amounts = append(amounts, netAmount)
	}
	adminFees = append(adminFees, adminFee)
	return scenario, amounts, adminFees
}

var (
	logKey                  = "[ACUAN-TRANSACTION-NOTIF-TRANSFORMER]"
	skippedTransactionTypes = map[goAcuanLibModel.TransactionType]bool{
		"DSBAP": true,
		"DSBFD": true,
	}
	allowedTrxTypeWithScenario = map[goAcuanLibModel.TransactionType]bool{
		"ADMME": true,
		"ITRTF": true,
		"TUPPY": true,
		"TUPVA": true,
	}
)

func NewPAS(
	cfg *config.Configuration,
	publisher kafka.Publisher,
	accountingClient accounting.Client,
	flagClient flag.Client,
	notificationClient dddnotification.Client,
) (Rule[paspkg.OutMessage], error) {
	r, err := newRule(cfg, "PAS Rule", "pas.grl")
	p := &pas{
		rule:               r,
		publisher:          publisher,
		cfg:                cfg,
		accountingClient:   accountingClient,
		flagClient:         flagClient,
		notificationClient: notificationClient,
	}
	return p, err
}

func (r pas) Execute(ctx context.Context, data paspkg.OutMessage) (err error) {
	loc, err := time.LoadLocation(TimezoneJakarta)
	if err != nil {
		return err
	}

	for _, v := range data.AcuanData.Body.Data.Order.Transactions {
		tx := &pasTransformed{
			cfg: r.cfg,
			pasPublisher: pasPublisher{
				ctx:                ctx,
				publisher:          r.publisher,
				accountingClient:   r.accountingClient,
				notificationClient: r.notificationClient,
			},
			timeLoc: loc,
		}

		if skippedTransactionTypes[v.TransactionType] {
			continue
		}

		dctx := ast.NewDataContext()
		incomingData := buildRetryMessage(data, v)
		transactionId := v.Id.String()
		logField := []xlog.Field{
			xlog.String("transaction-id", transactionId),
			xlog.Any("incoming-data", incomingData),
		}

		if v.Amount.IsZero() {
			xlog.Warnf(ctx, "skip the transforms because the amount is zero")
			continue
		}
		meta, ok := v.Meta.(map[string]interface{})
		if !ok {
			xlog.Warnf(ctx, "want type map[string]interface{};  got %T", v.Meta)
		}

		trxDate := v.TransactionTime.Time.In(loc)
		orderTime := data.AcuanData.Body.Data.Order.OrderTime.In(loc)
		if trxDate.Hour() == 0 && trxDate.Minute() == 0 && trxDate.Second() == 0 {
			trxDate = time.Date(trxDate.Year(), trxDate.Month(), trxDate.Day(), orderTime.Hour(), orderTime.Minute(), orderTime.Second(), orderTime.Nanosecond(), orderTime.Location())
		}

		acuanTransaction := AcuanTransaction{
			AcuanStatus:          data.Status,
			Identifier:           data.Identifier,
			OrderTime:            data.AcuanData.Body.Data.Order.OrderTime,
			OrderType:            goAcuanLibModel.OrderType(v.TransactionType[:3]),
			RefNumber:            data.AcuanData.Body.Data.Order.RefNumber,
			Id:                   v.Id,
			Amount:               v.Amount,
			Currency:             v.Currency,
			SourceAccountId:      v.SourceAccountId,
			DestinationAccountId: v.DestinationAccountId,
			Description:          v.Description,
			Method:               v.Method,
			TransactionType:      v.TransactionType,
			TransactionTime:      trxDate.Format(DateFormatYYYYMMDDWithTime),
			Status:               v.Status,
			Meta:                 meta,

			AccountBalances:   data.AccountBalances,
			WalletTransaction: data.WalletTransaction,
		}

		if err = dctx.Add("Flag", &flagClient{r.flagClient}); err != nil {
			notifyAndLogError(ctx, tx, acuanTransaction, logField, "Add Flag to Rules", incomingData, err)
			continue
		}

		if err = dctx.Add("Acuan", &acuanTransaction); err != nil {
			notifyAndLogError(ctx, tx, acuanTransaction, logField, "Add Acuan to Rules", incomingData, err)
			continue
		}

		if err = dctx.Add("Transaction", tx); err != nil {
			notifyAndLogError(ctx, tx, acuanTransaction, logField, "Add Transaction to Rules", incomingData, err)
			continue
		}

		tx.Payload.Transactions = make([]*JournalTransaction, 0, 6)
		if err = dctx.Add("Journal", &tx.Payload); err != nil {
			notifyAndLogError(ctx, tx, acuanTransaction, logField, "Add Journal to Rules", incomingData, err)
			continue
		}

		var JournalDebit1, JournalCredit1, journalDebit2, journalCredit2 JournalTransaction
		if err = dctx.Add("JournalDebit1", &JournalDebit1); err != nil {
			notifyAndLogError(ctx, tx, acuanTransaction, logField, "Add JournalDebit1 to Rules", incomingData, err)
			continue
		}
		if err = dctx.Add("JournalCredit1", &JournalCredit1); err != nil {
			notifyAndLogError(ctx, tx, acuanTransaction, logField, "Add JournalCredit1 to Rules", incomingData, err)
			continue
		}
		if err = dctx.Add("JournalDebit2", &journalDebit2); err != nil {
			notifyAndLogError(ctx, tx, acuanTransaction, logField, "Add JournalDebit2 to Rules", incomingData, err)
			continue
		}
		if err = dctx.Add("JournalCredit2", &journalCredit2); err != nil {
			notifyAndLogError(ctx, tx, acuanTransaction, logField, "Add JournalCredit2 to Rules", incomingData, err)
			continue
		}

		var (
			scenario           int
			amounts, adminFees []decimal.Decimal
		)
		if allowedTrxTypeWithScenario[v.TransactionType] && acuanTransaction.AccountBalances != nil {
			scenario, amounts, adminFees = acuanTransaction.HandleTopupScenario()
			if scenario != 0 {
				if err = dctx.Add("Amounts", amounts); err != nil {
					notifyAndLogError(ctx, tx, acuanTransaction, logField, "Add Amounts to Rules", incomingData, err)
					continue
				}
				if err = dctx.Add("AdminFees", adminFees); err != nil {
					notifyAndLogError(ctx, tx, acuanTransaction, logField, "Add AdminFees to Rules", incomingData, err)
					continue
				}
			}
		}
		if err = dctx.Add("Scenario", scenario); err != nil {
			notifyAndLogError(ctx, tx, acuanTransaction, logField, "Add Scenario to Rules", incomingData, err)
		}

		if err = r.executeEngine(ctx, dctx); err != nil {
			if data.IsRetry {
				notifyAndLogError(ctx, tx, acuanTransaction, logField, "Retry Process Transform", acuanTransaction, err)
				return
			}
			notifyAndLogError(ctx, tx, acuanTransaction, logField, "Process Transform", acuanTransaction, err)
			publishToDLQ(ctx, tx, acuanTransaction, logField, incomingData, err)
			continue
		}

		if err = tx.prePublishErrors.ErrorOrNil(); err != nil {
			if data.IsRetry {
				notifyAndLogError(ctx, tx, acuanTransaction, logField, "Retry Process Pre Publish", tx.Payload, err)
				return
			}
			notifyAndLogError(ctx, tx, acuanTransaction, logField, "Process Pre Publish", tx.Payload, err)
			publishToDLQ(ctx, tx, acuanTransaction, logField, incomingData, err)
			continue
		}

		if err = tx.errPublish; err != nil {
			if data.IsRetry {
				notifyAndLogError(ctx, tx, acuanTransaction, logField, "Retry Process Publish", tx.Payload, tx.errPublish)
				return
			}
			notifyAndLogError(ctx, tx, acuanTransaction, logField, "Process Publish", tx.Payload, tx.errPublish)
			publishToDLQ(ctx, tx, acuanTransaction, logField, incomingData, tx.errPublish)
			continue
		}
		xlog.Info(ctx, logKey, logField...)
	}

	return nil
}

func buildRetryMessage(data paspkg.OutMessage, transaction goAcuanLibModel.Transaction) paspkg.OutMessage {
	rm := data
	rm.AcuanData.Body.Data.Order.Transactions = []goAcuanLibModel.Transaction{transaction}
	return rm
}

var skipSlackRules = []struct {
	TransactionType string
	ErrorSubstring  string
}{
	{
		TransactionType: "RVRSL",
		ErrorSubstring:  "invalid response http code: got 404, message: transaction id not found",
	},
}

func shouldSkipSlack(acuan AcuanTransaction, err error) bool {
	for _, rule := range skipSlackRules {
		if acuan.TransactionType == goAcuanLibModel.TransactionType(rule.TransactionType) &&
			strings.Contains(err.Error(), rule.ErrorSubstring) {
			return true
		}
	}
	return false
}

func notifyAndLogError(
	ctx context.Context,
	tx *pasTransformed, acuan AcuanTransaction,
	logField []xlog.Field,
	operation string,
	payload interface{},
	errCauser error,
) {
	logField = append(logField,
		xlog.Any("processing-data", payload),
		xlog.Err(errCauser))
	xlog.Warn(ctx, logKey, logField...)

	if shouldSkipSlack(acuan, errCauser) {
		return
	}

	if err := tx.notificationClient.SendMessageToSlack(ctx, dddnotification.MessageData{
		Operation: operation,
		Message:   fmt.Sprintf("<!channel> Failed with trx id: %s, trx type %s, err: %s", acuan.Id.String(), acuan.TransactionType, errCauser.Error()),
	}); err != nil {
		logField = append(logField, xlog.Any("error-send-notif", err))
		xlog.Error(ctx, logKey, logField...)
	}
}

func publishToDLQ(
	ctx context.Context,
	tx *pasTransformed,
	acuan AcuanTransaction,
	logField []xlog.Field,
	payload paspkg.OutMessage,
	errorCauser error,
) error {
	if errors.Is(errorCauser, errNotMatch) {
		return nil
	}

	payload.IsRetry = true
	msg := struct {
		paspkg.OutMessage
		ErrCauser   interface{} `json:"errorCauser"`
		PublishedAt time.Time   `json:"publishedAt"`
	}{
		OutMessage:  payload,
		ErrCauser:   errorCauser.Error(),
		PublishedAt: time.Now().In(tx.timeLoc),
	}

	if err := tx.publisher.PublishSyncWithKeyAndLog(ctx,
		"publish transaction to transformer_stream_dlq",
		tx.cfg.Kafka.Publishers.TransformerStreamDLQ.Topic,
		acuan.Id.String(),
		msg,
	); err != nil {
		notifyAndLogError(ctx, tx, acuan, logField, "Publish to DLQ Failed", payload, err)
		return err
	}
	return nil
}
