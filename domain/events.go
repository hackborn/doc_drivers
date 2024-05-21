package domain

// Events stores a sequence of events.
type Events struct {
	Time  int64 `doc:"key"`
	Name  string
	Value string

	_table int `doc:"name(events)"`
}
