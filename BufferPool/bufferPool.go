package main

import (
	"bufio"
	"fmt"
	"os"
)

type BufferPool struct {
	velicina int            //velicina poola
	redosled []int          //niz id blokova
	podaci   map[int][]byte //mapa id-sadrzaj bloka
}

// kad pravimo prazan bp prosledjujemo velicinu pool
func Napravi(velPool int) *BufferPool {
	return &BufferPool{
		velicina: velPool,
		redosled: make([]int, 0, velPool),
		podaci:   make(map[int][]byte),
	}
}

// upis-upise na kraj ako je popunjen vrati inf da je popunjen
// prosledjuje mu se id stranice bloka i sadrzaj bloka
func (bp *BufferPool) Upisi(id int, sadrzaj []byte) {
	//ako vec postoji azurira mu se pozicija
	_, postoji := bp.podaci[id]
	if postoji {
		//ukloni iz liste
		for i, vr := range bp.redosled {
			if vr == id {
				bp.redosled = append(bp.redosled[:i], bp.redosled[i+1:]...)
				break
			}
		}
		//dodaj na kraj
		bp.redosled = append(bp.redosled, id)
		//azuriraj blok
		bp.podaci[id] = sadrzaj

	} else {
		//da li je pun
		if len(bp.redosled) >= bp.velicina {
			fmt.Printf("Buffer pool je pun uradi flush. Hvala Milice!")
			najstariji := bp.redosled[0]
			//izbaci iz mape i liste najstariji da bi bilo mesta za novi
			delete(bp.podaci, najstariji)
			bp.redosled = bp.redosled[1:]
		}
		//ako ne postoji i nije pun dodaj u mapu i na kraj niza
		bp.podaci[id] = sadrzaj
		bp.redosled = append(bp.redosled, id)
	}
}

// citanje-nadje ga-izbaci ga i upise na kraj-vrati ga
func (bp *BufferPool) Citaj(id int) []byte {
	//nadjemo ga u mapi
	pod, postoji := bp.podaci[id]
	if !postoji {
		fmt.Printf("Ne postoji taj blok.")
		return nil
	}
	//ukloni iz liste
	for i, vr := range bp.redosled {
		if vr == id {
			bp.redosled = append(bp.redosled[:i], bp.redosled[i+1:]...)
			break
		}
	}
	//dodaj na kraj
	bp.redosled = append(bp.redosled, id)
	return pod

}

// cuvam sve podatke iz niza u fajlu
func (bp *BufferPool) SacuvajUFajl(fajl string) error {
	f, err := os.Create((fajl))
	if err != nil {
		return err
	}
	defer f.Close() //zatvara fajl na kraju funkcije sta god da se desi

	for _, id := range bp.redosled {
		pod := bp.podaci[id]
		linija := fmt.Sprintf("%d:%s\n", id, string(pod))
		f.WriteString(linija)
	}
	return nil

}

// pravim bp na osnovu fajla
func UcitajIzFajla(fajl string, vel int) (*BufferPool, error) {
	f, err := os.Open(fajl)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	bp := Napravi(vel)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var id int
		var tekst string
		fmt.Sscanf(scanner.Text(), "%d:%s", &id, &tekst)
		bp.Upisi(id, []byte(tekst))
	}
	return bp, nil
}
