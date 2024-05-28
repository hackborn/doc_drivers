package domain

// CollectionSetting tests a slice of int64s. Slices are
// unhandled by some backends so need to be serialized.
// This is an example of the serializing being handled
// automatically.
type CollectionSetting struct {
	Name  string `doc:"key"`
	Value []int64

	_table int `doc:"name(settings)"`
}
