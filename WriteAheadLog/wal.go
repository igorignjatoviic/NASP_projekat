package wal

func Unesi(dogadjaj string, kljuc string, vrednost string) {
	wal := WriteAheadLog{}
	wal.unesi(dogadjaj, kljuc, vrednost)
}

func Ispisi() {
	wal := WriteAheadLog{}
	wal.ucitajWriteAheadLog()
}
