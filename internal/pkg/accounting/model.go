package accounting

type ResponseGetLenderAccount struct {
	Kind                    string `json:"kind"`
	CihAccountNumber        string `json:"cihAccountNumber"`
	InvestedAccountNumber   string `json:"investedAccountNumber"`
	ReceivableAccountNumber string `json:"receivablesAccountNumber"`
}

type ResponseGetJournalDetail struct {
	Kind      string          `json:"kind"`
	Journals  []JournalDetail `json:"contents"`
	TotalRows int             `json:"total_rows"`
}

type JournalDetail struct {
	Kind            string `json:"kind"`
	TransactionId   string `json:"transactionId"`
	AccountNumber   string `json:"accountNumber"`
	TransactionType string `json:"transactionType"`
	Amount          string `json:"amount"`
	TransactionDate string `json:"transactionDate"`
	Narrative       string `json:"narrative"`
	IsDebit         bool   `json:"isDebit"`
}

type ResponseGetAccountByAccountNumber struct {
	Kind            string                  `json:"kind" example:"account"`
	AccountNumber   string                  `json:"accountNumber" example:"21100100000001"`
	AccountName     string                  `json:"accountName" example:"Lender Yang Baik"`
	CoaTypeCode     string                  `json:"coaTypeCode" example:"LIA"`
	CoaTypeName     string                  `json:"coaTypeName" example:"Liability"`
	CategoryCode    string                  `json:"categoryCode" example:"211"`
	CategoryName    string                  `json:"categoryName" example:"Marketplace Payable (Lender Balance)"`
	SubCategoryCode string                  `json:"subCategoryCode" example:"21101"`
	SubCategoryName string                  `json:"subCategoryName" example:"Lender Balance - Individual Non RDL"`
	AltID           string                  `json:"altID" example:"535235235235325"`
	EntityCode      string                  `json:"entityCode" example:"001"`
	EntityName      string                  `json:"entityName" example:"PT. Amartha Mikro Fintek (AMF)"`
	ProductTypeCode string                  `json:"productTypeCode" example:"1001"`
	ProductTypeName string                  `json:"productTypeName" example:"Group Loan"`
	OwnerID         string                  `json:"ownerID" example:"211000000412"`
	Status          string                  `json:"status" example:"active"`
	Currency        string                  `json:"currency" example:"IDR"`
	AccountType     string                  `json:"accountType" example:"LENDER_RETAIL"`
	LegacyId        *map[string]interface{} `json:"legacyId,omitempty" swaggertype:"object,string" example:"t24AccountNumber:1234567890,t24ArrangementId:1234567890"`
	Metadata        *map[string]interface{} `json:"metadata,omitempty" swaggertype:"object,string" example:"t24AccountNumber:1234567890,t24ArrangementId:1234567890"`
	CreatedAt       string                  `json:"createdAt" example:"2006-01-02 15:04:05"`
	UpdatedAt       string                  `json:"updatedAt" example:"2006-01-02 15:04:05"`
}

type (
	DoGetAllAccountNumbersByParamRequest struct {
		OwnerId        string
		AltId          string
		AccountNumbers string
	}
	DoGetAllAccountNumbersByParamResponse struct {
		Contents []GetAllAccountNumbersByParam `json:"contents"`
	}
	GetAllAccountNumbersByParam struct {
		Kind            string `json:"kind"`
		OwnerId         string `json:"ownerId"`
		AccountNumber   string `json:"accountNumber"`
		AltId           string `json:"altId"`
		Name            string `json:"name"`
		AccountType     string `json:"accountType"`
		EntityCode      string `json:"entityCode"`
		ProductTypeCode string `json:"productTypeCode"`
		CategoryCode    string `json:"categoryCode"`
		SubCategoryCode string `json:"subCategoryCode"`
		Currency        string `json:"currency"`
		Status          string `json:"status"`
	}
)

type (
	DoGetLoanPartnerAccountByParamsRequest struct {
		PartnerId           string `json:"partnerId,omitempty"`
		LoanKind            string `json:"loanKind,omitempty"`
		AccountNumber       string `json:"accountNumber,omitempty"`
		AccountType         string `json:"accountType,omitempty"`
		EntityCode          string `json:"entityCode,omitempty"`
		LoanSubCategoryCode string `json:"loanSubCategoryCode,omitempty"`
		LoanAccountNumber   string `json:"loanAccountNumber,omitempty"`
	}
	DoGetLoanPartnerAccountByParamsResponse struct {
		Contents []GetLoanPartnerAccountByParams `json:"contents"`
	}
	GetLoanPartnerAccountByParams struct {
		Kind                string `json:"kind"`
		PartnerId           string `json:"partnerId"`
		LoanKind            string `json:"loanKind"`
		AccountNumber       string `json:"accountNumber"`
		AccountType         string `json:"accountType"`
		EntityCode          string `json:"entityCode"`
		LoanSubCategoryCode string `json:"loanSubCategoryCode"`
		CreatedAt           string `json:"createdAt"`
		UpdatedAt           string `json:"updatedAt"`
	}

	DoGetEntityByParamsRequest struct {
		EntityCode string `query:"entityCode"`
		Name       string `query:"name"`
	}
	DoGetEntityResponse struct {
		Kind        string `json:"kind" example:"entity"`
		Code        string `json:"code" example:"001"`
		Name        string `json:"name,omitempty" example:"AMF"`
		Description string `json:"description,omitempty" example:"PT. Amartha Mikro Fintek"`
		Status      string `json:"status,omitempty" example:"active"`
	}
)
