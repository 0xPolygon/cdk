package translator

type Translator interface {
	Translate(contextName string, data string) string
}
