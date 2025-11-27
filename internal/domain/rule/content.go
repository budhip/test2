package rule

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type Content struct {
	value string
}

func NewContent(content string) (Content, error) {
	if strings.TrimSpace(content) == "" {
		return Content{}, fmt.Errorf("content cannot be empty")
	}

	if !strings.Contains(content, "rule ") {
		return Content{}, fmt.Errorf("content must contain at least one 'rule' keyword")
	}

	return Content{value: content}, nil
}

type InsertMode string

const (
	InsertAtEnd       InsertMode = "end"
	InsertAtBeginning InsertMode = "beginning"
	InsertBySalience  InsertMode = "salience"
)

func (c Content) Append(additional string, mode InsertMode) (Content, error) {
	if strings.TrimSpace(additional) == "" {
		return c, fmt.Errorf("additional content cannot be empty")
	}

	var newContent string

	switch mode {
	case InsertAtEnd:
		newContent = c.appendAtEnd(additional)
	case InsertAtBeginning:
		newContent = c.appendAtBeginning(additional)
	case InsertBySalience:
		newContent = c.insertBySalience(additional)
	default:
		return c, fmt.Errorf("invalid insert mode: %s", mode)
	}

	return Content{value: newContent}, nil
}

func (c Content) appendAtEnd(additional string) string {
	existing := c.value
	if !strings.HasSuffix(existing, "\n") {
		existing += "\n"
	}
	if !strings.HasSuffix(additional, "\n") {
		additional += "\n"
	}
	return existing + additional
}

func (c Content) appendAtBeginning(additional string) string {
	if !strings.HasSuffix(additional, "\n") {
		additional += "\n"
	}
	return additional + c.value
}

func (c Content) insertBySalience(additional string) string {
	existingRules := c.parseRules()
	newRules := parseRulesFromString(additional)

	allRules := append(existingRules, newRules...)
	sortRulesBySalience(allRules)

	var result strings.Builder
	for _, rule := range allRules {
		result.WriteString(rule.Content)
		result.WriteString("\n")
	}

	return result.String()
}

func (c Content) parseRules() []RuleParsed {
	return parseRulesFromString(c.value)
}

func parseRulesFromString(content string) []RuleParsed {
	var rules []RuleParsed

	rulePattern := regexp.MustCompile(`(?s)rule\s+(\w+)\s+.*?salience\s+(-?\d+)\s*\{.*?\}`)
	matches := rulePattern.FindAllStringSubmatch(content, -1)
	matchIndices := rulePattern.FindAllStringIndex(content, -1)

	for i, match := range matches {
		if len(match) >= 3 {
			name := match[1]
			salience, _ := strconv.Atoi(match[2])

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

	if len(rules) == 0 {
		rules = append(rules, RuleParsed{
			Name:     "unknown",
			Salience: 0,
			Content:  content,
		})
	}

	return rules
}

type RuleParsed struct {
	Name     string
	Salience int
	Content  string
}

func sortRulesBySalience(rules []RuleParsed) {
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Salience > rules[j].Salience
	})
}

func (c Content) String() string {
	return c.value
}

func (c Content) ExtractRuleNames() []string {
	var names []string
	pattern := regexp.MustCompile(`rule\s+(\w+)\s+`)
	matches := pattern.FindAllStringSubmatch(c.value, -1)

	for _, match := range matches {
		if len(match) >= 2 {
			names = append(names, match[1])
		}
	}

	return names
}
