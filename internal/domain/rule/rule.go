package rule

import (
	"time"
)

type Rule struct {
	id        RuleID
	name      RuleName
	env       Environment
	version   Version
	content   Content
	isActive  bool
	createdAt time.Time
	updatedAt time.Time
	createdBy string
	updatedBy string
}

func NewRule(name string, env string, version string, content string) (*Rule, error) {
	ruleID := NewRuleID()

	ruleName, err := NewRuleName(name)
	if err != nil {
		return nil, err
	}

	environment, err := NewEnvironment(env)
	if err != nil {
		return nil, err
	}

	ver, err := NewVersion(version)
	if err != nil {
		return nil, err
	}

	cont, err := NewContent(content)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	return &Rule{
		id:        ruleID,
		name:      ruleName,
		env:       environment,
		version:   ver,
		content:   cont,
		isActive:  true,
		createdAt: now,
		updatedAt: now,
	}, nil
}

func ReconstructRule(
	id RuleID,
	name RuleName,
	env Environment,
	version Version,
	content Content,
	isActive bool,
	createdAt time.Time,
	updatedAt time.Time,
	createdBy string,
	updatedBy string,
) *Rule {
	return &Rule{
		id:        id,
		name:      name,
		env:       env,
		version:   version,
		content:   content,
		isActive:  isActive,
		createdAt: createdAt,
		updatedAt: updatedAt,
		createdBy: createdBy,
		updatedBy: updatedBy,
	}
}

func (r *Rule) UpdateContent(newContent string) error {
	content, err := NewContent(newContent)
	if err != nil {
		return err
	}

	r.content = content
	r.updatedAt = time.Now()
	return nil
}

func (r *Rule) BumpVersion(bumpType VersionBumpType) error {
	newVersion, err := r.version.Bump(bumpType)
	if err != nil {
		return err
	}

	r.version = newVersion
	r.updatedAt = time.Now()
	return nil
}

func (r *Rule) AppendContent(additionalContent string, mode InsertMode) error {
	newContent, err := r.content.Append(additionalContent, mode)
	if err != nil {
		return err
	}

	r.content = newContent
	r.updatedAt = time.Now()
	return nil
}

func (r *Rule) Deactivate() {
	r.isActive = false
	r.updatedAt = time.Now()
}

func (r *Rule) Activate() {
	r.isActive = true
	r.updatedAt = time.Now()
}

func (r *Rule) SetCreatedBy(user string) {
	r.createdBy = user
}

func (r *Rule) SetUpdatedBy(user string) {
	r.updatedBy = user
}

func (r *Rule) ID() RuleID {
	return r.id
}

func (r *Rule) Name() RuleName {
	return r.name
}

func (r *Rule) Env() Environment {
	return r.env
}

func (r *Rule) Version() Version {
	return r.version
}

func (r *Rule) Content() Content {
	return r.content
}

func (r *Rule) IsActive() bool {
	return r.isActive
}

func (r *Rule) CreatedAt() time.Time {
	return r.createdAt
}

func (r *Rule) UpdatedAt() time.Time {
	return r.updatedAt
}

func (r *Rule) CreatedBy() string {
	return r.createdBy
}

func (r *Rule) UpdatedBy() string {
	return r.updatedBy
}
