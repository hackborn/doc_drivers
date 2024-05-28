package domain

// Skip is an example of using an unexported
// field tag to omit a struct from the databse.
type Skip struct {
	Name  string `doc:"key"`
	Value float64

	_table int `doc:"-"`
}
