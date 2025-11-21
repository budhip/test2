package rules

import (
	"bitbucket.org/Amartha/go-megatron/internal/common/grule"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	xlog "bitbucket.org/Amartha/go-x/log"

	"bitbucket.org/Amartha/go-megatron/internal/models"
	"bitbucket.org/Amartha/go-megatron/internal/repositories"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	repo repositories.RuleRepository
}

func NewHandler(repo repositories.RuleRepository) *Handler {
	return &Handler{repo: repo}
}

// CreateRule creates a new rule
func (h *Handler) CreateRule(c echo.Context) error {
	ctx := c.Request().Context()
	var req models.CreateRuleRequest

	if err := c.Bind(&req); err != nil {
		xlog.Warn(ctx, "[RULES_HANDLER] Invalid request body", xlog.Err(err))
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
	}

	if err := req.Validate(); err != nil {
		xlog.Warn(ctx, "[RULES_HANDLER] Validation failed", xlog.Err(err))
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Validation failed",
			Message: err.Error(),
		})
	}

	// Validate Grule syntax
	validator := grule.NewRuleValidator()
	if err := validator.ValidateGruleSyntax(req.Content, req.Name); err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid Grule syntax",
			Message: err.Error(),
		})
	}

	rule := &models.Rule{
		Name:     req.Name,
		Env:      req.Env,
		Version:  req.Version,
		Content:  req.Content,
		IsActive: true,
	}

	if err := h.repo.CreateRule(ctx, rule); err != nil {
		xlog.Error(ctx, "[RULES_HANDLER] Failed to create rule", xlog.Err(err))
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to create rule",
			Message: err.Error(),
		})
	}

	xlog.Info(ctx, "[RULES_HANDLER] Rule created successfully",
		xlog.String("name", req.Name),
		xlog.Int64("id", rule.ID))

	return c.JSON(http.StatusCreated, models.ToRuleResponse(rule))
}

// GetRule gets a specific rule
func (h *Handler) GetRule(c echo.Context) error {
	ctx := c.Request().Context()
	name := c.Param("name")
	env := c.QueryParam("env")
	version := c.QueryParam("version")

	if name == "" || env == "" || version == "" {
		xlog.Warn(ctx, "[RULES_HANDLER] Missing parameters")
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Missing parameters",
			Message: "name, env, and version are required",
		})
	}

	rule, err := h.repo.GetRule(ctx, name, env, version)
	if err != nil {
		xlog.Error(ctx, "[RULES_HANDLER] Failed to get rule", xlog.Err(err))
		return c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "Rule not found",
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, models.ToRuleResponse(rule))
}

// ListRules lists all rules for an environment
func (h *Handler) ListRules(c echo.Context) error {
	ctx := c.Request().Context()
	env := c.QueryParam("env")

	if env == "" {
		xlog.Warn(ctx, "[RULES_HANDLER] Missing env parameter")
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Missing parameter",
			Message: "env is required",
		})
	}

	rules, err := h.repo.ListRules(ctx, env)
	if err != nil {
		xlog.Error(ctx, "[RULES_HANDLER] Failed to list rules", xlog.Err(err))
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to list rules",
			Message: err.Error(),
		})
	}

	response := make([]models.RuleResponse, len(rules))
	for i, rule := range rules {
		response[i] = models.ToRuleResponse(rule)
	}

	return c.JSON(http.StatusOK, response)
}

// UpdateRule updates an existing rule
func (h *Handler) UpdateRule(c echo.Context) error {
	ctx := c.Request().Context()
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		xlog.Warn(ctx, "[RULES_HANDLER] Invalid ID", xlog.Err(err))
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid ID",
			Message: err.Error(),
		})
	}

	var req models.UpdateRuleRequest
	if err := c.Bind(&req); err != nil {
		xlog.Warn(ctx, "[RULES_HANDLER] Invalid request body", xlog.Err(err))
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
	}

	if err := req.Validate(); err != nil {
		xlog.Warn(ctx, "[RULES_HANDLER] Validation failed", xlog.Err(err))
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Validation failed",
			Message: err.Error(),
		})
	}

	rule := &models.Rule{
		ID:      id,
		Content: req.Content,
	}

	if err := h.repo.UpdateRule(ctx, rule); err != nil {
		xlog.Error(ctx, "[RULES_HANDLER] Failed to update rule", xlog.Err(err))
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to update rule",
			Message: err.Error(),
		})
	}

	xlog.Info(ctx, "[RULES_HANDLER] Rule updated successfully", xlog.Int64("id", id))

	return c.JSON(http.StatusOK, models.MessageResponse{
		Message: "Rule updated successfully",
	})
}

// DeleteRule soft deletes a rule
func (h *Handler) DeleteRule(c echo.Context) error {
	ctx := c.Request().Context()
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		xlog.Warn(ctx, "[RULES_HANDLER] Invalid ID", xlog.Err(err))
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid ID",
			Message: err.Error(),
		})
	}

	if err := h.repo.DeleteRule(ctx, id); err != nil {
		xlog.Error(ctx, "[RULES_HANDLER] Failed to delete rule", xlog.Err(err))
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to delete rule",
			Message: err.Error(),
		})
	}

	xlog.Info(ctx, "[RULES_HANDLER] Rule deleted successfully", xlog.Int64("id", id))

	return c.JSON(http.StatusOK, models.MessageResponse{
		Message: "Rule deleted successfully",
	})
}

// AppendRule appends a new rule to existing content without replacing
func (h *Handler) AppendRule(c echo.Context) error {
	ctx := c.Request().Context()

	// Get rule ID from path
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid rule ID",
			Message: err.Error(),
		})
	}

	// Parse request
	var req models.AppendRuleRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
	}

	if err := req.Validate(); err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Validation failed",
			Message: err.Error(),
		})
	}

	// Set defaults
	if req.InsertMode == "" {
		req.InsertMode = "end"
	}
	if req.VersionBump == "" && req.AutoVersion {
		req.VersionBump = "patch"
	}

	xlog.Info(ctx, "[RULES_HANDLER] Appending rule",
		xlog.Int64("id", id),
		xlog.String("mode", req.InsertMode),
		xlog.Bool("auto_version", req.AutoVersion))

	// Get existing rule
	existingRule, err := h.repo.GetRuleByID(ctx, id)
	if err != nil {
		xlog.Error(ctx, "[RULES_HANDLER] Failed to get rule", xlog.Err(err))
		return c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "Rule not found",
			Message: err.Error(),
		})
	}

	// Merge content based on insert mode
	newContent, err := mergeRuleContent(
		existingRule.Content,
		req.Content,
		req.InsertMode,
	)
	if err != nil {
		xlog.Error(ctx, "[RULES_HANDLER] Failed to merge content", xlog.Err(err))
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to merge content",
			Message: err.Error(),
		})
	}

	// Calculate new version
	newVersion := existingRule.Version
	if req.AutoVersion {
		newVersion, err = bumpVersion(existingRule.Version, req.VersionBump)
		if err != nil {
			xlog.Warn(ctx, "[RULES_HANDLER] Failed to bump version, keeping same",
				xlog.Err(err))
		}
	}

	// Update rule
	existingRule.Content = newContent
	existingRule.Version = newVersion

	if err := h.repo.UpdateRule(ctx, existingRule); err != nil {
		xlog.Error(ctx, "[RULES_HANDLER] Failed to update rule", xlog.Err(err))
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to update rule",
			Message: err.Error(),
		})
	}

	xlog.Info(ctx, "[RULES_HANDLER] Rule appended successfully",
		xlog.Int64("id", id),
		xlog.String("old_version", existingRule.Version),
		xlog.String("new_version", newVersion))

	// Extract rule names from new content
	addedRules := extractRuleNames(req.Content)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":     "Rule appended successfully",
		"rule":        models.ToRuleResponse(existingRule),
		"old_version": existingRule.Version,
		"new_version": newVersion,
		"added_rules": addedRules,
		"insert_mode": req.InsertMode,
	})
}

// mergeRuleContent merges new content with existing content
func mergeRuleContent(existing, newContent, mode string) (string, error) {
	// Ensure both contents end with newline
	if !strings.HasSuffix(existing, "\n") {
		existing += "\n"
	}
	if !strings.HasSuffix(newContent, "\n") {
		newContent += "\n"
	}

	switch mode {
	case "end":
		// Simply append to the end
		return existing + newContent, nil

	case "beginning":
		// Add to the beginning
		return newContent + existing, nil

	case "salience":
		// Insert based on salience value (more complex)
		return insertBySalience(existing, newContent)

	default:
		return existing + newContent, nil
	}
}

// insertBySalience inserts new rules based on salience value
func insertBySalience(existing, newContent string) (string, error) {
	// Parse existing rules
	existingRules := parseRules(existing)

	// Parse new rules
	newRules := parseRules(newContent)

	// Merge and sort by salience (descending)
	allRules := append(existingRules, newRules...)

	// Sort by salience (highest first)
	sortRulesBySalience(allRules)

	// Reconstruct content
	var result strings.Builder
	for _, rule := range allRules {
		result.WriteString(rule.Content)
		result.WriteString("\n")
	}

	return result.String(), nil
}

// RuleParsed represents a parsed rule with salience
type RuleParsed struct {
	Name     string
	Salience int
	Content  string
}

// parseRules parses rule content and extracts rules with their salience
func parseRules(content string) []RuleParsed {
	var rules []RuleParsed

	// Regex to match: rule Name "Description" salience N { ... }
	rulePattern := regexp.MustCompile(`(?s)rule\s+(\w+)\s+.*?salience\s+(-?\d+)\s*\{.*?\}`)

	matches := rulePattern.FindAllStringSubmatch(content, -1)
	matchIndices := rulePattern.FindAllStringIndex(content, -1)

	for i, match := range matches {
		if len(match) >= 3 {
			name := match[1]
			salience, _ := strconv.Atoi(match[2])

			// Extract full rule content
			start := matchIndices[i][0]
			end := matchIndices[i][1]
			ruleContent := content[start:end]

			rules = append(rules, RuleParsed{
				Name:     name,
				Salience: salience,
				Content:  ruleContent,
			})
		}
	}

	// If no salience found, treat entire content as one rule with salience 0
	if len(rules) == 0 {
		rules = append(rules, RuleParsed{
			Name:     "unknown",
			Salience: 0,
			Content:  content,
		})
	}

	return rules
}

// sortRulesBySalience sorts rules by salience (descending)
func sortRulesBySalience(rules []RuleParsed) {
	// Bubble sort (simple, good enough for small rule sets)
	n := len(rules)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if rules[j].Salience < rules[j+1].Salience {
				rules[j], rules[j+1] = rules[j+1], rules[j]
			}
		}
	}
}

// extractRuleNames extracts rule names from content
func extractRuleNames(content string) []string {
	var names []string

	// Regex to match: rule Name
	pattern := regexp.MustCompile(`rule\s+(\w+)\s+`)
	matches := pattern.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) >= 2 {
			names = append(names, match[1])
		}
	}

	return names
}

// bumpVersion increments version based on bump type
func bumpVersion(current, bump string) (string, error) {
	// Parse semantic version: major.minor.patch
	parts := strings.Split(current, ".")

	// Ensure we have 3 parts
	for len(parts) < 3 {
		parts = append(parts, "0")
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return current, fmt.Errorf("invalid major version: %s", parts[0])
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return current, fmt.Errorf("invalid minor version: %s", parts[1])
	}

	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return current, fmt.Errorf("invalid patch version: %s", parts[2])
	}

	// Increment based on bump type
	switch bump {
	case "major":
		major++
		minor = 0
		patch = 0
	case "minor":
		minor++
		patch = 0
	case "patch":
		patch++
	default:
		return current, fmt.Errorf("invalid bump type: %s", bump)
	}

	return fmt.Sprintf("%d.%d.%d", major, minor, patch), nil
}
