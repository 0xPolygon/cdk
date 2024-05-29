package inmem

type Inmem struct {
}

func (inmem *Inmem) GetData() []byte {
	return []byte("Inmem")
}
