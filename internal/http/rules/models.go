package rules

import (
	"errors"
	"time"

	"bitbucket.org/Amartha/go-megatron/internal/repository"
)

// CreateRuleRequest represents the request to create a new rule
type CreateRuleRequest struct {
	Name    string `json:"name" example:"papa.grl"`
	Env     string `json:"env" example:"dev"`
	Version string `json:"version" example:"0.0.1"`
	Content string `json:"content"`
}

func (r *CreateRuleRequest) Validate() error {
	if r.Name == "" {
		return errors.New("name is required")
	}
	if r.Env == "" {
		return errors.New("env is required")
	}
	if r.Version == "" {
		return errors.New("version is required")
	}
	if r.Content == "" {
		return errors.New("content is required")
	}

	// Validate env values
	validEnvs := map[string]bool{"dev": true, "uat": true, "prod": true}
	if !validEnvs[r.Env] {
		return errors.New("env must be one of: dev, uat, prod")
	}

	return nil
}

// UpdateRuleRequest represents the request to update a rule
type UpdateRuleRequest struct {
	Name    string `json:"name" example:"papa.grl"`
	Env     string `json:"env" example:"dev"`
	Version string `json:"version" example:"0.0.1"`
	Content string `json:"content"`
}

func (r *UpdateRuleRequest) Validate() error {
	if r.Content == "" {
		return errors.New("content is required")
	}
	return nil
}

// RuleResponse represents the rule response
type RuleResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Env       string    `json:"env"`
	Version   string    `json:"version"`
	Content   string    `json:"content"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// MessageResponse represents a simple message response
type MessageResponse struct {
	Message string `json:"message"`
}

// toRuleResponse converts a repository.Rule to RuleResponse
func toRuleResponse(rule *repository.Rule) RuleResponse {
	return RuleResponse{
		ID:        rule.ID,
		Name:      rule.Name,
		Env:       rule.Env,
		Version:   rule.Version,
		Content:   rule.Content,
		IsActive:  rule.IsActive,
		CreatedAt: rule.CreatedAt,
		UpdatedAt: rule.UpdatedAt,
	}
}

// AppendRuleRequest represents request to append a new rule to existing content
type AppendRuleRequest struct {
	Content     string `json:"content" example:"rule NewRule {...}"`
	AutoVersion bool   `json:"auto_version" example:"true"`            // Auto-increment version
	VersionBump string `json:"version_bump,omitempty" example:"patch"` // major, minor, patch
	InsertMode  string `json:"insert_mode,omitempty" example:"end"`    // end, salience, beginning
}
