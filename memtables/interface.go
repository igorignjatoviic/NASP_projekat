package memtables

type Struktura interface {
	Ubaci(kljuc string, vrednost []byte)
	Dobavi(kljuc string) ([]byte, bool)
	Obrisi(kljuc string) bool
	NadjiUnos(kljuc string) (Unos, bool)
	DobaviSve() []Unos
	Duzina() int
	DaLiFlush() bool
	Isprazni()
}
