package request

import "fmt"

type CreateRuleRequest struct {
	Name    string `json:"name"`
	Env     string `json:"env"`
	Version string `json:"version"`
	Content string `json:"content"`
}

func (r *CreateRuleRequest) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("name is required")
	}
	if r.Env == "" {
		return fmt.Errorf("env is required")
	}
	if r.Version == "" {
		return fmt.Errorf("version is required")
	}
	if r.Content == "" {
		return fmt.Errorf("content is required")
	}

	validEnvs := map[string]bool{"dev": true, "uat": true, "prod": true}
	if !validEnvs[r.Env] {
		return fmt.Errorf("env must be one of: dev, uat, prod")
	}

	return nil
}
