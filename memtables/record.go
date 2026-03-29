package memtables

type Unos struct {
	Kljuc     string
	Vrednost  []byte
	Timestamp int64
	Tombstone bool
}
