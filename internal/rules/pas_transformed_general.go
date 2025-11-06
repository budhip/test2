package rules

import (
	"fmt"

	"bitbucket.org/Amartha/go-megatron/internal/pkg/accounting"

	"github.com/hashicorp/go-multierror"
)

type accountTypeEnum int64

const (
	accountType1 accountTypeEnum = iota + 1
	accountType2
	accountType3
	accountType4
	accountType5
	accountType6
)

func (a accountTypeEnum) String() string {
	switch a {
	case accountType1:
		return "CASH_IN_TRANSIT_DISBURSE"
	case accountType2:
		return "CASH_IN_TRANSIT_REPAYMENT"
	case accountType3:
		return "INTERNAL_ACCOUNTS_REVENUE_AMARTHA"
	case accountType4:
		return "INTERNAL_ACCOUNTS_ADMIN_FEE_AMARTHA"
	case accountType5:
		return "INTERNAL_ACCOUNTS_PPH_AMARTHA"
	case accountType6:
		return "INTERNAL_ACCOUNTS_PPN_AMARTHA"
	default:
		return "UNKNOWN"
	}
}

func (t *pasTransformed) GetLoanPartnerByLoanAccountNumberAccountType(loanAccountNumber string, accountTypeKey int64) string {
	accountType := accountTypeEnum(accountTypeKey).String()

	resp, err := t.accountingClient.GetLoanPartnerAccountByParams(t.ctx, accounting.DoGetLoanPartnerAccountByParamsRequest{
		LoanAccountNumber: loanAccountNumber,
		AccountType:       accountType,
	})
	if err != nil {
		t.prePublishErrors = multierror.Append(t.prePublishErrors, err)
		return ""
	}

	if len(resp.Contents) == 0 {
		t.prePublishErrors = multierror.Append(t.prePublishErrors,
			fmt.Errorf("loan partner account not found by loanAccountNumber:%v - accountType:%v", loanAccountNumber, accountType))
		return ""
	}

	return resp.Contents[0].AccountNumber
}

func (t *pasTransformed) GetLoanPartnerByParams(params map[string]interface{}, accountTypeKey int64) string {
	accountType := accountTypeEnum(accountTypeKey).String()
	payload := buildLoanPartnerQueryWithOptions(params)
	resp, err := t.accountingClient.GetLoanPartnerAccountByParams(t.ctx, accounting.DoGetLoanPartnerAccountByParamsRequest{
		LoanKind:            payload.LoanKind,
		AccountType:         accountType,
		PartnerId:           payload.PartnerId,
		AccountNumber:       payload.AccountNumber,
		EntityCode:          payload.EntityCode,
		LoanSubCategoryCode: payload.LoanSubCategoryCode,
		LoanAccountNumber:   payload.LoanAccountNumber,
	})
	if err != nil {
		t.prePublishErrors = multierror.Append(t.prePublishErrors, err)
		return ""
	}

	if len(resp.Contents) == 0 {
		t.prePublishErrors = multierror.Append(t.prePublishErrors,
			fmt.Errorf("loan partner account not found by payload:%v", payload))
		return ""
	}
	return resp.Contents[0].AccountNumber
}

func buildLoanPartnerQueryWithOptions(options map[string]interface{}) (param accounting.DoGetLoanPartnerAccountByParamsRequest) {
	if val, ok := options["loanKind"].(string); ok {
		param.LoanKind = val
	}
	if val, ok := options["partnerId"].(string); ok {
		param.PartnerId = val
	}
	if val, ok := options["accountType"].(string); ok {
		param.AccountType = val
	}
	if val, ok := options["accountNumber"].(string); ok {
		param.AccountNumber = val
	}
	if val, ok := options["entityCode"].(string); ok {
		param.EntityCode = val
	}
	if val, ok := options["loanSubCategoryCode"].(string); ok {
		param.LoanSubCategoryCode = val
	}
	if val, ok := options["loanAccountNumber"].(string); ok {
		param.LoanAccountNumber = val
	}
	return param
}

func (t *pasTransformed) MappingInterface(kv ...interface{}) map[string]interface{} {
	m := make(map[string]interface{})
	for i := 0; i < len(kv)-1; i += 2 {
		key, ok := kv[i].(string)
		if !ok {
			continue
		}
		m[key] = kv[i+1]
	}
	return m
}
