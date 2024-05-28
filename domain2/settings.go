package domain2

// UiSetting tests a map of strings. Maps are
// unhandled by some backends so need to be serialized.
// This is an example of the serializing being handled
// by a format tag.
// Additionally, this tests multiple domain folders,
// making sure everything gets combined correctly.
type UiSetting struct {
	Name  string            `doc:"key"`
	Value map[string]string `doc:"format(json)"`

	_table int `doc:"name(settings)"`
}
