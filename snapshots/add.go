package snapshots

type Add struct {
	files []string
}

func (a *Add) NewAdd() *Add {
	return &Add{
		files: []string{},
	}
}
