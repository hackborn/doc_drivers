package domain

// FavouritesSetting stores indexes to the list of favourites.
// It is used to test custom structs that are excluded from the metadata.
type FavouritesSetting struct {
	Name  string `doc:"key"`
	Value []FavEntry

	_table int `doc:"name(settings)"`
}

type FavEntry struct {
	Id       int64
	LastUsed int64

	_table int `doc:"-"`
}
