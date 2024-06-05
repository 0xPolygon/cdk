package datastream

type Datastream struct {
	Source string
}

func (d *Datastream) Source() string {
	return d.Source
}
