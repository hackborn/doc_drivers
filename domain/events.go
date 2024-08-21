package domain

// Events stores a sequence of events.
type Events struct {
	Time  uint64 `doc:"key, autoinc"`
	Name  string
	Value string

	_table int `doc:"name(events)"`
}
