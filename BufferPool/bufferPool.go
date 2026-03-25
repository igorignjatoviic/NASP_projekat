package bufferpool

import (
	"encoding/binary"
	"fmt"
	"os"
)

// novi bufferpool
type Tuple struct {
	Dogadjaj        string
	duzinaKljuca    uint64
	duzinaVrednosti uint64
	Kljuc           string
	Vrednost        string
}

func NoviTuple(dogadjaj string, kljuc string, vrednost string) *Tuple {
	return &Tuple{
		Dogadjaj:        dogadjaj,
		Kljuc:           kljuc,
		Vrednost:        vrednost,
		duzinaKljuca:    uint64(len(kljuc)),
		duzinaVrednosti: uint64(len(vrednost)),
	}
}

type BufferPool struct {
	velicina uint64
	podaci   []Tuple
}

func (bp *BufferPool) Unesi(niz []Tuple) error {
	fajl, err := os.Create("BufferPool/resources/bufferpool.bin")
	if err != nil {
		return err
	}
	defer fajl.Close()

	for _, podatak := range niz {
		tombstone := 0
		if podatak.Dogadjaj == "delete" {
			tombstone = 1
		}

		if err := binary.Write(fajl, binary.BigEndian, uint8(tombstone)); err != nil {
			fmt.Println(err)
		}
		if err := binary.Write(fajl, binary.BigEndian, uint64(podatak.duzinaKljuca)); err != nil {
			fmt.Println(err)
		}
		if err := binary.Write(fajl, binary.BigEndian, uint64(podatak.duzinaVrednosti)); err != nil {
			fmt.Println(err)
		}

		kljuc := []byte(podatak.Kljuc)
		if err := binary.Write(fajl, binary.BigEndian, kljuc); err != nil {
			fmt.Println(err)
		}

		vrednost := []byte(podatak.Vrednost)
		if err := binary.Write(fajl, binary.BigEndian, vrednost); err != nil {
			fmt.Println(err)
		}
	}

	return nil
}

func NoviBufferPool() *BufferPool {
	velicina := 5
	return &BufferPool{
		velicina: uint64(velicina), // zapravo cita iz konfiguracija
		podaci:   make([]Tuple, velicina),
	}
}

func OsveziBufferPool(buffer []Tuple, element Tuple) []Tuple {
	for i := range len(buffer) - 1 {
		buffer[i] = buffer[i+1]
	}
	buffer[len(buffer)-1] = element

	return buffer
}

func GenerisiPocetniFajl() {
	bp := BufferPool{}
	bp.velicina = 5
	bp.podaci = make([]Tuple, bp.velicina)

	for i := range 5 {
		t := NoviTuple("put", "igor", "njnjnjnjnj")
		bp.podaci[i] = *t
	}

	fajl, err := os.Create("BufferPool/resources/bufferpool.bin")
	if err != nil {
		fmt.Println("Greska prilikom otvaranja fajla.")
	}
	defer fajl.Close()

	for _, podatak := range bp.podaci {
		tombstone := 0
		if podatak.Dogadjaj == "delete" {
			tombstone = 1
		}

		if err := binary.Write(fajl, binary.BigEndian, uint8(tombstone)); err != nil {
			fmt.Println(err)
		}
		if err := binary.Write(fajl, binary.BigEndian, podatak.duzinaKljuca); err != nil {
			fmt.Println(err)
		}
		if err := binary.Write(fajl, binary.BigEndian, podatak.duzinaVrednosti); err != nil {
			fmt.Println(err)
		}

		kljuc := []byte(podatak.Kljuc)
		if err := binary.Write(fajl, binary.BigEndian, kljuc); err != nil {
			fmt.Println(err)
		}

		vrednost := []byte(podatak.Vrednost)
		if err := binary.Write(fajl, binary.BigEndian, vrednost); err != nil {
			fmt.Println(err)
		}
	}
}

func Ucitaj() []Tuple {
	bp := NoviBufferPool()

	fajl, err := os.Open("BufferPool/resources/bufferpool.bin")
	if err != nil {
		fmt.Println("Greska prilikom otvaranja fajla u citanju.")
	}
	defer fajl.Close()

	for i := range bp.velicina { // fiksna velicina za sada
		var tombstone uint8
		binary.Read(fajl, binary.BigEndian, &tombstone)

		var duzinaKljuca uint64
		binary.Read(fajl, binary.BigEndian, &duzinaKljuca)

		var duzinaVrednosti uint64
		binary.Read(fajl, binary.BigEndian, &duzinaVrednosti)

		kljuc := make([]byte, duzinaKljuca)
		binary.Read(fajl, binary.BigEndian, &kljuc)

		vrednost := make([]byte, duzinaVrednosti)
		binary.Read(fajl, binary.BigEndian, &vrednost)

		dogadjaj := "put"
		if tombstone == 1 {
			dogadjaj = "delete"
		}

		bp.podaci[i] = *NoviTuple(dogadjaj, string(kljuc), string(vrednost))
	}

	return bp.podaci
}
