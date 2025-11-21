package models

import (
	"errors"
	"fmt"
	"time"
)

// Rule represents a rule entity in the database
type Rule struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	Env       string    `db:"env"`
	Version   string    `db:"version"`
	Content   string    `db:"content"`
	IsActive  bool      `db:"is_active"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

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

// MessageResponse represents a simple message response
type MessageResponse struct {
	Message string `json:"message"`
}

// ToRuleResponse converts a repository.Rule to RuleResponse
func ToRuleResponse(rule *Rule) RuleResponse {
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

func (r *AppendRuleRequest) Validate() error {
	if r.Content == "" {
		return fmt.Errorf("content is required")
	}

	// Validate insert mode
	validModes := map[string]bool{
		"end":       true,
		"beginning": true,
		"salience":  true,
	}
	if r.InsertMode != "" && !validModes[r.InsertMode] {
		return fmt.Errorf("insert_mode must be one of: end, beginning, salience")
	}

	// Validate version bump
	validBumps := map[string]bool{
		"major": true,
		"minor": true,
		"patch": true,
	}
	if r.VersionBump != "" && !validBumps[r.VersionBump] {
		return fmt.Errorf("version_bump must be one of: major, minor, patch")
	}

	return nil
}
