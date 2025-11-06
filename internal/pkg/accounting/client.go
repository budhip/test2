package accounting

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"bitbucket.org/Amartha/go-megatron/internal/config"
	xlog "bitbucket.org/Amartha/go-x/log"
	"bitbucket.org/Amartha/go-x/log/ctxdata"

	"github.com/go-resty/resty/v2"
)

var logMessage = "[ACCOUNTING-CLIENT]"

type Client interface {
	GetInvestedAccountNumber(ctx context.Context, cihAccountNumber string) (accountNumber string, err error)
	GetReceivableAccountNumber(ctx context.Context, cihAccountNumber string) (accountNumber string, err error)
	GetJournalDetail(ctx context.Context, transactionId string) (journals ResponseGetJournalDetail, err error)
	GetAccountByAccountNumber(ctx context.Context, accountNumber string) (account ResponseGetAccountByAccountNumber, err error)
	GetAccountByLegacyID(ctx context.Context, legacyID string) (account ResponseGetAccountByAccountNumber, err error)
	GetAccountByAccountNumberOrLegacyID(ctx context.Context, accountNumber string) (account ResponseGetAccountByAccountNumber, err error)
	GetAccountsByParams(ctx context.Context, in DoGetAllAccountNumbersByParamRequest) (accountMap map[string][]GetAllAccountNumbersByParam, err error)
	GetLoanPartnerAccountByParams(ctx context.Context, in DoGetLoanPartnerAccountByParamsRequest) (res DoGetLoanPartnerAccountByParamsResponse, err error)
	GetEntityByParams(ctx context.Context, in DoGetEntityByParamsRequest) (res DoGetEntityResponse, err error)
}

type client struct {
	baseURL    string
	secretKey  string
	httpClient *resty.Client
}

func New(configuration config.HTTPConfiguration) Client {
	retryWaitTime := time.Duration(configuration.RetryWaitTime) * time.Millisecond

	restyClient := resty.New().
		SetRetryCount(configuration.RetryCount).
		SetRetryWaitTime(retryWaitTime)

	return client{
		baseURL:    configuration.BaseURL,
		secretKey:  configuration.SecretKey,
		httpClient: restyClient,
	}
}

func (c client) GetInvestedAccountNumber(ctx context.Context, cihAccountNumber string) (accountNumber string, err error) {
	url := fmt.Sprintf("%s/api/v1/lender-accounts/%s", c.baseURL, cihAccountNumber)

	logFields := []xlog.Field{
		xlog.String("url", url),
		xlog.String("cihAccountNumber", cihAccountNumber),
	}

	defer func() {
		if err != nil {
			xlog.Warn(ctx, logMessage, append(logFields, xlog.Err(err))...)
		}
	}()

	xlog.Info(ctx, logMessage, append(logFields, xlog.String("message", "send request to go_accounting"))...)

	httpRes, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader("Accept", "application/json;  charset=utf-8").
		SetHeader("Cache-Control", "no-cache").
		SetHeader("X-Correlation-Id", ctxdata.GetCorrelationId(ctx)).
		SetHeader("X-Secret-Key", c.secretKey).
		Get(url)
	if err != nil {
		err = fmt.Errorf("failed send request: %w", err)
		return "", err
	}

	logFields = append(logFields,
		xlog.String("httpStatusCode", httpRes.Status()),
		xlog.Any("httpResponse", httpRes.Body()))

	if httpRes.StatusCode() != http.StatusOK {
		var message string
		if message, err = c.decodeErrMessageFromBody(string(httpRes.Body())); err != nil {
			message = fmt.Sprintf("unable to parse error message: %s", err.Error())
		}
		return "", fmt.Errorf("invalid response http code: got %d, message: %s", httpRes.StatusCode(), message)
	}

	var res ResponseGetLenderAccount
	if err = json.Unmarshal(httpRes.Body(), &res); err != nil {
		err = fmt.Errorf("error unmarshal response: %w", err)
		return "", err
	}

	return res.InvestedAccountNumber, nil
}

func (c client) GetReceivableAccountNumber(ctx context.Context, cihAccountNumber string) (accountNumber string, err error) {
	url := fmt.Sprintf("%s/api/v1/lender-accounts/%s", c.baseURL, cihAccountNumber)

	logFields := []xlog.Field{
		xlog.String("url", url),
		xlog.String("cihAccountNumber", cihAccountNumber),
	}

	defer func() {
		if err != nil {
			xlog.Warn(ctx, logMessage, append(logFields, xlog.Err(err))...)
		}
	}()

	xlog.Info(ctx, logMessage, append(logFields, xlog.String("message", "send request to go_accounting"))...)

	httpRes, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader("Accept", "application/json;  charset=utf-8").
		SetHeader("Cache-Control", "no-cache").
		SetHeader("X-Correlation-Id", ctxdata.GetCorrelationId(ctx)).
		SetHeader("X-Secret-Key", c.secretKey).
		Get(url)
	if err != nil {
		err = fmt.Errorf("failed send request: %w", err)
		return "", err
	}

	logFields = append(logFields,
		xlog.String("httpStatusCode", httpRes.Status()),
		xlog.Any("httpResponse", httpRes.Body()))

	if httpRes.StatusCode() != http.StatusOK {
		var message string
		if message, err = c.decodeErrMessageFromBody(string(httpRes.Body())); err != nil {
			message = fmt.Sprintf("unable to parse error message: %s", err.Error())
		}
		return "", fmt.Errorf("invalid response http code: got %d, message: %s", httpRes.StatusCode(), message)
	}

	var res ResponseGetLenderAccount
	if err = json.Unmarshal(httpRes.Body(), &res); err != nil {
		err = fmt.Errorf("error unmarshal response: %w", err)
		return "", err
	}

	return res.ReceivableAccountNumber, nil
}

func (c client) GetJournalDetail(ctx context.Context, transactionId string) (journals ResponseGetJournalDetail, err error) {
	url := fmt.Sprintf("%s/api/v1/journals/%s", c.baseURL, transactionId)

	logFields := []xlog.Field{
		xlog.String("url", url),
		xlog.String("transaction-id", transactionId),
	}

	defer func() {
		if err != nil {
			xlog.Warn(ctx, logMessage, append(logFields, xlog.Err(err))...)
		}
	}()

	xlog.Info(ctx, logMessage, append(logFields, xlog.String("message", "send request to go_accounting"))...)

	httpRes, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader("Accept", "application/json;  charset=utf-8").
		SetHeader("Cache-Control", "no-cache").
		SetHeader("X-Correlation-Id", ctxdata.GetCorrelationId(ctx)).
		SetHeader("X-Secret-Key", c.secretKey).
		Get(url)
	if err != nil {
		err = fmt.Errorf("failed send request: %w", err)
		return journals, err
	}

	logFields = append(logFields,
		xlog.String("httpStatusCode", httpRes.Status()),
		xlog.Any("httpResponse", httpRes.Body()))

	if httpRes.StatusCode() != http.StatusOK {
		var message string
		if message, err = c.decodeErrMessageFromBody(string(httpRes.Body())); err != nil {
			message = fmt.Sprintf("unable to parse error message: %s", err.Error())
		}
		return journals, fmt.Errorf("invalid response http code: got %d, message: %s", httpRes.StatusCode(), message)
	}

	var res ResponseGetJournalDetail
	if err = json.Unmarshal(httpRes.Body(), &res); err != nil {
		err = fmt.Errorf("error unmarshal response: %w", err)
		return journals, err
	}

	return res, nil
}

func (c client) GetAccountByAccountNumber(ctx context.Context, accountNumber string) (account ResponseGetAccountByAccountNumber, err error) {
	url := fmt.Sprintf("%s/api/v1/accounts/%s", c.baseURL, accountNumber)

	logFields := []xlog.Field{
		xlog.String("url", url),
		xlog.String("accountNumber", accountNumber),
	}

	defer func() {
		if err != nil {
			xlog.Warn(ctx, logMessage, append(logFields, xlog.Err(err))...)
		}
	}()

	xlog.Info(ctx, logMessage, append(logFields, xlog.String("message", "send request to go_accounting"))...)

	httpRes, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader("Accept", "application/json;  charset=utf-8").
		SetHeader("Cache-Control", "no-cache").
		SetHeader("X-Correlation-Id", ctxdata.GetCorrelationId(ctx)).
		SetHeader("X-Secret-Key", c.secretKey).
		Get(url)
	if err != nil {
		return account, fmt.Errorf("failed send request: %w", err)
	}

	logFields = append(logFields,
		xlog.String("httpStatusCode", httpRes.Status()),
		xlog.Any("httpResponse", httpRes.Body()))

	if httpRes.StatusCode() != http.StatusOK {
		var message string
		if message, err = c.decodeErrMessageFromBody(string(httpRes.Body())); err != nil {
			message = fmt.Sprintf("unable to parse error message: %s", err.Error())
		}
		return account, fmt.Errorf("invalid response http code: got %d, message: %s", httpRes.StatusCode(), message)
	}

	var res ResponseGetAccountByAccountNumber
	err = json.Unmarshal(httpRes.Body(), &res)
	if err != nil {
		return account, fmt.Errorf("error unmarshal response: %w", err)
	}

	return res, nil
}

func (c client) GetAccountByLegacyID(ctx context.Context, legacyID string) (account ResponseGetAccountByAccountNumber, err error) {
	url := fmt.Sprintf("%s/api/v1/accounts/t24/%s", c.baseURL, legacyID)

	logFields := []xlog.Field{
		xlog.String("url", url),
		xlog.String("accountNumber", legacyID),
	}

	defer func() {
		if err != nil {
			xlog.Warn(ctx, logMessage, append(logFields, xlog.Err(err))...)
		}
	}()

	xlog.Info(ctx, logMessage, append(logFields, xlog.String("message", "send request to go_accounting"))...)

	httpRes, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader("Accept", "application/json;  charset=utf-8").
		SetHeader("Cache-Control", "no-cache").
		SetHeader("X-Correlation-Id", ctxdata.GetCorrelationId(ctx)).
		SetHeader("X-Secret-Key", c.secretKey).
		Get(url)
	if err != nil {
		return account, fmt.Errorf("failed send request: %w", err)
	}

	logFields = append(logFields,
		xlog.String("httpStatusCode", httpRes.Status()),
		xlog.Any("httpResponse", httpRes.Body()))

	if httpRes.StatusCode() != http.StatusOK {
		var message string
		if message, err = c.decodeErrMessageFromBody(string(httpRes.Body())); err != nil {
			message = fmt.Sprintf("unable to parse error message: %s", err.Error())
		}
		return account, fmt.Errorf("invalid response http code: got %d, message: %s", httpRes.StatusCode(), message)
	}

	var res ResponseGetAccountByAccountNumber
	err = json.Unmarshal(httpRes.Body(), &res)
	if err != nil {
		return account, fmt.Errorf("error unmarshal response: %w", err)
	}

	return res, nil
}

func (c client) GetAccountByAccountNumberOrLegacyID(ctx context.Context, accountNumber string) (account ResponseGetAccountByAccountNumber, err error) {
	urlAccount := fmt.Sprintf("%s/api/v1/accounts/%s", c.baseURL, accountNumber)

	logFields := []xlog.Field{
		xlog.String("url", urlAccount),
		xlog.String("accountNumber", accountNumber),
	}

	defer func() {
		if err != nil {
			xlog.Warn(ctx, logMessage, append(logFields, xlog.Err(err))...)
		}
	}()

	xlog.Info(ctx, logMessage, append(logFields, xlog.String("message", "send request to go_accounting"))...)

	httpRes, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader("Accept", "application/json;  charset=utf-8").
		SetHeader("Cache-Control", "no-cache").
		SetHeader("X-Correlation-Id", ctxdata.GetCorrelationId(ctx)).
		SetHeader("X-Secret-Key", c.secretKey).
		Get(urlAccount)
	if err != nil {
		return account, fmt.Errorf("failed send request: %w", err)
	}

	logFields = append(logFields,
		xlog.String("httpStatusCode", httpRes.Status()),
		xlog.Any("httpResponse", httpRes.Body()))

	switch {
	case httpRes.StatusCode() == http.StatusOK:
		var res ResponseGetAccountByAccountNumber
		err = json.Unmarshal(httpRes.Body(), &res)
		if err != nil {
			return account, fmt.Errorf("error unmarshal response: %w", err)
		}

		return res, nil
	case httpRes.StatusCode() == http.StatusNotFound:
		accountLegacy, errGetAcc := c.GetAccountByLegacyID(ctx, accountNumber)
		if errGetAcc != nil {
			return accountLegacy, fmt.Errorf("failed send request: %w", errGetAcc)
		}

		return accountLegacy, nil
	default:
		var message string
		if message, err = c.decodeErrMessageFromBody(string(httpRes.Body())); err != nil {
			message = fmt.Sprintf("unable to parse error message: %s", err.Error())
		}
		return account, fmt.Errorf("invalid response http code: got %d, message: %s", httpRes.StatusCode(), message)
	}
}

func (c client) GetAccountsByParams(ctx context.Context, in DoGetAllAccountNumbersByParamRequest) (accountMap map[string][]GetAllAccountNumbersByParam, err error) {
	params := url.Values{}
	if in.OwnerId != "" {
		params.Add("ownerId", in.OwnerId)
	}
	if in.AltId != "" {
		params.Add("altId", in.AltId)
	}
	if in.AccountNumbers != "" {
		params.Add("accountNumbers", in.AccountNumbers)
	}
	url := fmt.Sprintf("%s/api/v1/accounts/account-numbers?%s", c.baseURL, params.Encode())

	logFields := []xlog.Field{
		xlog.String("url", url),
		xlog.Any("params", params),
	}

	defer func() {
		if err != nil {
			xlog.Warn(ctx, logMessage, append(logFields, xlog.Err(err))...)
		}
	}()

	xlog.Info(ctx, logMessage, append(logFields, xlog.String("message", "send request to go_accounting"))...)

	httpRes, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader("Accept", "application/json;  charset=utf-8").
		SetHeader("Cache-Control", "no-cache").
		SetHeader("X-Correlation-Id", ctxdata.GetCorrelationId(ctx)).
		SetHeader("X-Secret-Key", c.secretKey).
		Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed send request: %w", err)
	}

	logFields = append(logFields,
		xlog.String("httpStatusCode", httpRes.Status()),
		xlog.Any("httpResponse", httpRes.Body()))

	if httpRes.StatusCode() != http.StatusOK {
		var message string
		if message, err = c.decodeErrMessageFromBody(string(httpRes.Body())); err != nil {
			message = fmt.Sprintf("unable to parse error message: %s", err.Error())
		}
		return nil, fmt.Errorf("invalid response http code: got %d, message: %s", httpRes.StatusCode(), message)
	}

	var res DoGetAllAccountNumbersByParamResponse
	if err = json.Unmarshal(httpRes.Body(), &res); err != nil {
		return nil, fmt.Errorf("error unmarshal response: %w", err)
	}

	if len(res.Contents) == 0 {
		return nil, fmt.Errorf("data not found %+v", in)
	}

	accountMap = make(map[string][]GetAllAccountNumbersByParam)
	for _, v := range res.Contents {
		key := fmt.Sprintf("%s+%s", v.AltId, v.SubCategoryCode)
		accountMap[key] = append(accountMap[key], v)
		accountMap[v.AccountNumber] = append(accountMap[v.AccountNumber], v)
	}

	return accountMap, nil
}

func (c client) GetLoanPartnerAccountByParams(ctx context.Context, in DoGetLoanPartnerAccountByParamsRequest) (res DoGetLoanPartnerAccountByParamsResponse, err error) {
	params := url.Values{}
	if in.PartnerId != "" {
		params.Add("partnerId", in.PartnerId)
	}
	if in.LoanKind != "" {
		params.Add("loanKind", in.LoanKind)
	}
	if in.AccountType != "" {
		params.Add("accountType", in.AccountType)
	}
	if in.EntityCode != "" {
		params.Add("entityCode", in.EntityCode)
	}
	if in.LoanSubCategoryCode != "" {
		params.Add("loanSubCategoryCode", in.LoanSubCategoryCode)
	}
	if in.LoanAccountNumber != "" {
		params.Add("loanAccountNumber", in.LoanAccountNumber)
	}
	url := fmt.Sprintf("%s/api/v1/loan-partner-accounts?%s", c.baseURL, params.Encode())

	logFields := []xlog.Field{
		xlog.String("url", url),
		xlog.Any("params", params),
	}

	defer func() {
		if err != nil {
			xlog.Warn(ctx, logMessage, append(logFields, xlog.Err(err))...)
		}
	}()

	xlog.Info(ctx, logMessage, append(logFields, xlog.String("message", "send request to go_accounting"))...)

	httpRes, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader("Accept", "application/json;  charset=utf-8").
		SetHeader("Cache-Control", "no-cache").
		SetHeader("X-Correlation-Id", ctxdata.GetCorrelationId(ctx)).
		SetHeader("X-Secret-Key", c.secretKey).
		Get(url)
	if err != nil {
		return res, fmt.Errorf("failed send request: %w", err)
	}

	logFields = append(logFields,
		xlog.String("httpStatusCode", httpRes.Status()),
		xlog.Any("httpResponse", httpRes.Body()))

	if httpRes.StatusCode() != http.StatusOK {
		var message string
		if message, err = c.decodeErrMessageFromBody(string(httpRes.Body())); err != nil {
			message = fmt.Sprintf("unable to parse error message: %s", err.Error())
		}
		return res, fmt.Errorf("invalid response http code: got %d, message: %s", httpRes.StatusCode(), message)
	}

	if err = json.Unmarshal(httpRes.Body(), &res); err != nil {
		return res, fmt.Errorf("error unmarshal response: %w", err)
	}

	return res, nil
}

func (c client) GetEntityByParams(ctx context.Context, in DoGetEntityByParamsRequest) (res DoGetEntityResponse, err error) {
	params := url.Values{}
	if in.EntityCode != "" {
		params.Add("entityCode", in.EntityCode)
	}
	if in.Name != "" {
		params.Add("name", in.Name)
	}
	url := fmt.Sprintf("%s/api/v1/entities/search?%s", c.baseURL, params.Encode())

	logFields := []xlog.Field{
		xlog.String("url", url),
		xlog.Any("params", params),
	}

	defer func() {
		if err != nil {
			xlog.Warn(ctx, logMessage, append(logFields, xlog.Err(err))...)
		}
	}()

	xlog.Info(ctx, logMessage, append(logFields, xlog.String("message", "send request to go_accounting"))...)

	httpRes, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader("Accept", "application/json;  charset=utf-8").
		SetHeader("Cache-Control", "no-cache").
		SetHeader("X-Correlation-Id", ctxdata.GetCorrelationId(ctx)).
		SetHeader("X-Secret-Key", c.secretKey).
		Get(url)
	if err != nil {
		return res, fmt.Errorf("failed send request: %w", err)
	}

	logFields = append(logFields,
		xlog.String("httpStatusCode", httpRes.Status()),
		xlog.Any("httpResponse", httpRes.Body()))

	if httpRes.StatusCode() != http.StatusOK {
		var message string
		if message, err = c.decodeErrMessageFromBody(string(httpRes.Body())); err != nil {
			message = fmt.Sprintf("unable to parse error message: %s", err.Error())
		}
		return res, fmt.Errorf("invalid response http code: got %d, message: %s", httpRes.StatusCode(), message)
	}

	if err = json.Unmarshal(httpRes.Body(), &res); err != nil {
		return res, fmt.Errorf("error unmarshal response: %w", err)
	}

	return res, nil
}

func (c client) decodeErrMessageFromBody(body string) (string, error) {
	if strings.HasPrefix(body, "<") {
		msg := extractMessageFromHTML(body)
		return "", fmt.Errorf("non-JSON error: %s", msg)
	}

	var res map[string]interface{}

	if err := json.Unmarshal([]byte(body), &res); err != nil {
		return "", err
	}

	// Safely extract "message" key
	if msg, ok := res["message"].(string); ok {
		return msg, nil
	}

	return "", fmt.Errorf("message field not found or not a string")
}

func extractMessageFromHTML(html string) string {
	// Try to extract <title> content
	reTitle := regexp.MustCompile(`(?i)<title>(.*?)</title>`)
	if match := reTitle.FindStringSubmatch(html); len(match) == 2 {
		return strings.TrimSpace(match[1])
	}

	// Try to extract <h1> content
	reH1 := regexp.MustCompile(`(?i)<h1>(.*?)</h1>`)
	if match := reH1.FindStringSubmatch(html); len(match) == 2 {
		return strings.TrimSpace(match[1])
	}

	// Fallback: return first 100 chars
	return firstN(html, 100)
}

func firstN(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
