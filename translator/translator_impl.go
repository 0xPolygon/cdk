package translator

import "github.com/0xPolygonHermez/zkevm-synchronizer-l1/log"

type TranslatorFullMatchRule struct {
	// If null match any context
	ContextName *string
	// If null match any data
	FullMatchString string
	NewString       string
}

func (t *TranslatorFullMatchRule) Match(contextName string, data string) bool {
	if t.ContextName != nil && *t.ContextName != contextName {
		return false
	}
	if t.FullMatchString != data {
		return false
	}
	return true
}

func (t *TranslatorFullMatchRule) Translate(contextName string, data string) string {
	return t.NewString
}

func NewTranslatorFullMatchRule(contextName *string, fullMatchString string, newString string) *TranslatorFullMatchRule {
	return &TranslatorFullMatchRule{
		ContextName:     contextName,
		FullMatchString: fullMatchString,
		NewString:       newString,
	}
}

type TranslatorImpl struct {
	FullMatchRules []TranslatorFullMatchRule
}

func NewTranslatorImpl() *TranslatorImpl {
	return &TranslatorImpl{
		FullMatchRules: []TranslatorFullMatchRule{},
	}
}

func (t *TranslatorImpl) Translate(contextName string, data string) string {
	for _, rule := range t.FullMatchRules {
		if rule.Match(contextName, data) {
			translated := rule.Translate(contextName, data)
			log.Debugf("Translated (ctxName=%s) %s to %s", contextName, data, translated)
			return translated
		}
	}
	return data
}

func (t *TranslatorImpl) AddRule(rule TranslatorFullMatchRule) {
	t.FullMatchRules = append(t.FullMatchRules, rule)
}

func (t *TranslatorImpl) AddConfigRules(cfg Config) {
	for _, v := range cfg.FullMatchRules {
		var contextName *string
		if v.ContextName != "" {
			contextName = &v.ContextName
		}
		rule := NewTranslatorFullMatchRule(contextName, v.Old, v.New)
		t.AddRule(*rule)
	}
}
