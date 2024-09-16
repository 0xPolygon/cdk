package types

type Migration struct {
	ID     string
	SQL    string
	Prefix string
}
