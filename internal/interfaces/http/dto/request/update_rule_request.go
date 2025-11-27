package request

import "fmt"

type UpdateRuleRequest struct {
	Content string `json:"content"`
}

func (r *UpdateRuleRequest) Validate() error {
	if r.Content == "" {
		return fmt.Errorf("content is required")
	}
	return nil
}

type AppendRuleRequest struct {
	Content     string `json:"content"`
	AutoVersion bool   `json:"auto_version"`
	VersionBump string `json:"version_bump,omitempty"`
	InsertMode  string `json:"insert_mode,omitempty"`
}

func (r *AppendRuleRequest) Validate() error {
	if r.Content == "" {
		return fmt.Errorf("content is required")
	}

	validModes := map[string]bool{
		"end":       true,
		"beginning": true,
		"salience":  true,
	}
	if r.InsertMode != "" && !validModes[r.InsertMode] {
		return fmt.Errorf("insert_mode must be one of: end, beginning, salience")
	}

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
