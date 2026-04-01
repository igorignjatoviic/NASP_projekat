package bufferpool

import (
	configuration "NASP_projekat/Configuration"
	"encoding/binary"
	"fmt"
	"os"
)

type Tuple struct {
	Dogadjaj        string
	DuzinaKljuca    uint64
	DuzinaVrednosti uint64
	Kljuc           string
	Vrednost        string
}

func NoviTuple(dogadjaj string, kljuc string, vrednost string) *Tuple {
	return &Tuple{
		Dogadjaj:        dogadjaj,
		Kljuc:           kljuc,
		Vrednost:        vrednost,
		DuzinaKljuca:    uint64(len(kljuc)),
		DuzinaVrednosti: uint64(len(vrednost)),
	}
}

type BufferPool struct {
	velicina uint64
	Podaci   []Tuple
}

func (bp *BufferPool) Unesi(niz []Tuple) error {
	bp.Podaci = append(bp.Podaci, niz...) // MILICA JE DODALA OVU LINIJU
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
		if err := binary.Write(fajl, binary.BigEndian, uint64(podatak.DuzinaKljuca)); err != nil {
			fmt.Println(err)
		}
		if err := binary.Write(fajl, binary.BigEndian, uint64(podatak.DuzinaVrednosti)); err != nil {
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
	konfiguracija := configuration.UcitajKonfiguraciju()
	velicina := konfiguracija["BufferPool"]["Velicina segmenta"]

	return &BufferPool{
		velicina: uint64(velicina),
		Podaci:   make([]Tuple, 0, velicina),
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
	bp.Podaci = make([]Tuple, bp.velicina)

	for i := range 5 {
		t := NoviTuple("put", "igor", "njnjnjnjnj")
		bp.Podaci[i] = *t
	}

	fajl, err := os.Create("BufferPool/resources/bufferpool.bin")
	if err != nil {
		fmt.Println("Greska prilikom otvaranja fajla.")
	}
	defer fajl.Close()

	for _, podatak := range bp.Podaci {
		tombstone := 0
		if podatak.Dogadjaj == "delete" {
			tombstone = 1
		}

		if err := binary.Write(fajl, binary.BigEndian, uint8(tombstone)); err != nil {
			fmt.Println(err)
		}
		if err := binary.Write(fajl, binary.BigEndian, podatak.DuzinaKljuca); err != nil {
			fmt.Println(err)
		}
		if err := binary.Write(fajl, binary.BigEndian, podatak.DuzinaVrednosti); err != nil {
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
		return []Tuple{} // MILICA JE DODALA OVO
	}
	defer fajl.Close()

	// Prvo pročitaj koliko zapisa ima u fajlu
	podaci := make([]Tuple, 0) // MILICA JE DODALA
	i := uint64(0)             // MILICA JE DODALA
	for i < bp.velicina {
		//fmt.Print(i) // MILICA JE OVO DODALA JER MI JE IZBACIVALO GRESKU DA SE I NE KORISTI
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

		//bp.Podaci[i] = *NoviTuple(dogadjaj, string(kljuc), string(vrednost))	// MILICA JE ZAKOMENTARISALA
		podaci = append(podaci, *NoviTuple(dogadjaj, string(kljuc), string(vrednost))) // MILICA JE DODALA
		i++
	}

	//return bp.Podaci
	return podaci
}
