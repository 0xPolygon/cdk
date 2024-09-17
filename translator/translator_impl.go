package translator

import "github.com/0xPolygon/cdk/log"

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
	return t.FullMatchString == data
}

func (t *TranslatorFullMatchRule) Translate(contextName string, data string) string {
	return t.NewString
}

func NewTranslatorFullMatchRule(
	contextName *string, fullMatchString string, newString string,
) *TranslatorFullMatchRule {
	return &TranslatorFullMatchRule{
		ContextName:     contextName,
		FullMatchString: fullMatchString,
		NewString:       newString,
	}
}

type TranslatorImpl struct {
	logger         *log.Logger
	FullMatchRules []TranslatorFullMatchRule
}

func NewTranslatorImpl(logger *log.Logger) *TranslatorImpl {
	return &TranslatorImpl{
		logger:         logger,
		FullMatchRules: []TranslatorFullMatchRule{},
	}
}

func (t *TranslatorImpl) Translate(contextName string, data string) string {
	for _, rule := range t.FullMatchRules {
		if rule.Match(contextName, data) {
			translated := rule.Translate(contextName, data)
			t.logger.Debugf("Translated (ctxName=%s) %s to %s", contextName, data, translated)
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
			name := v.ContextName
			contextName = &name
		}
		rule := NewTranslatorFullMatchRule(contextName, v.Old, v.New)
		t.AddRule(*rule)
	}
}
