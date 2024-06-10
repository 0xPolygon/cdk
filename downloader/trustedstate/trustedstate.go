package trustedstate

type TrustedState struct {
	Source string
}

func (t *TrustedState) Download() string {
	return t.Source
}
