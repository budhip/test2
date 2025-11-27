package dto

import (
	"time"

	"bitbucket.org/Amartha/go-megatron/internal/domain/rule"
)

type RuleDTO struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Env       string    `json:"env"`
	Version   string    `json:"version"`
	Content   string    `json:"content"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedBy string    `json:"created_by,omitempty"`
	UpdatedBy string    `json:"updated_by,omitempty"`
}

func FromRuleEntity(r *rule.Rule) *RuleDTO {
	return &RuleDTO{
		ID:        r.ID().String(),
		Name:      r.Name().String(),
		Env:       r.Env().String(),
		Version:   r.Version().String(),
		Content:   r.Content().String(),
		IsActive:  r.IsActive(),
		CreatedAt: r.CreatedAt(),
		UpdatedAt: r.UpdatedAt(),
		CreatedBy: r.CreatedBy(),
		UpdatedBy: r.UpdatedBy(),
	}
}
