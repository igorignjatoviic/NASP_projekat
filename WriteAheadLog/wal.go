package wal

import (
	bufferpool "NASP_projekat/BufferPool"
	"fmt"
	"strings"
)

func Unesi(dogadjaj string, kljuc string, vrednost string) {
	wal := WriteAheadLog{}
	wal.unesi(dogadjaj, kljuc, vrednost)

	bp := bufferpool.NoviBufferPool()
	podaci := bufferpool.Ucitaj()
	podaci = bufferpool.OsveziBufferPool(podaci, *bufferpool.NoviTuple(dogadjaj, kljuc, vrednost))
	bp.Unesi(podaci)
}

func Ispisi() {
	wal := WriteAheadLog{}
	wal.ucitajWriteAheadLog()

	podaci := bufferpool.Ucitaj()
	fmt.Println("\nBufferPool:")
	for _, podatak := range podaci {
		fmt.Printf("\t%s(%s, %s)\n", strings.ToUpper(podatak.Dogadjaj), podatak.Kljuc, podatak.Vrednost)
	}
}
