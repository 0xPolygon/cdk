package translator

type ConfigRuleFullMatch struct {
	ContextName string `mapstructure:"ContextName"`
	Old         string `mapstructure:"Old"`
	New         string `mapstructure:"New"`
}

type Config struct {
	FullMatchRules []ConfigRuleFullMatch `mapstructure:"FullMatchRules"`
}
