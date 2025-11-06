package rules

import (
	"fmt"

	"bitbucket.org/Amartha/go-megatron/internal/pkg/accounting"
	"github.com/hashicorp/go-multierror"
)

func (t *pasTransformed) GetEntityByParams(params map[string]interface{}) string {
	request := buildGetEntityQueryWithOptions(params)

	if request.EntityCode == "" && request.Name == "" {
		t.prePublishErrors = multierror.Append(t.prePublishErrors,
			fmt.Errorf("empty entityCode or name when get entity"))
		return ""
	}

	resp, err := t.accountingClient.GetEntityByParams(t.ctx, request)
	if err != nil {
		t.prePublishErrors = multierror.Append(t.prePublishErrors, err)
		return ""
	}

	return resp.Code
}

func buildGetEntityQueryWithOptions(options map[string]interface{}) (param accounting.DoGetEntityByParamsRequest) {
	if val, ok := options["entityCode"].(string); ok {
		param.EntityCode = val
	}
	if val, ok := options["name"].(string); ok {
		param.Name = val
	}
	return param
}
