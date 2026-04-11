package main

import (
	"crypto/md5"
	"fmt"
	"os"
	"strings"
)

type Cvor struct {
	hes   string
	levi  *Cvor //pokazivaci na levo i desno podstablo
	desni *Cvor
}

type MerkleStablo struct {
	koren *Cvor
}

func Hash(blok string) string {
	h := md5.Sum([]byte(blok))
	return fmt.Sprintf("%x", h)
}

func Napravi(blokovi []string) *MerkleStablo {
	//ako je prazan
	if len(blokovi) == 0 {
		return &MerkleStablo{koren: nil}
	}
	//ako je neparan broj dodaj prazan
	if len(blokovi)%2 != 0 {
		blokovi = append(blokovi, "")
	}

	//dodajemo listove
	var cvorovi []*Cvor
	for _, b := range blokovi {
		h := Hash(b)
		cvorovi = append(cvorovi, &Cvor{hes: h})
	}

	//pravimo stablo
	for len(cvorovi) > 1 {
		var noviNiz []*Cvor

		for i := 0; i < len(cvorovi); i += 2 {
			levi := cvorovi[i]
			desni := cvorovi[i+1]

			novi := levi.hes + desni.hes
			h := Hash(novi)

			roditelj := &Cvor{
				hes:   h,
				levi:  levi,
				desni: desni,
			}
			noviNiz = append(noviNiz, roditelj)
		}
		//da li je opet neparan
		if len(noviNiz)%2 != 0 && len(noviNiz) != 1 {
			noviNiz = append(noviNiz, &Cvor{hes: Hash("")})
		}
		//sad sve isto radimo na novom nizu
		cvorovi = noviNiz
	}
	//kad dobijemo koren vratimo merklestablo
	return &MerkleStablo{koren: cvorovi[0]}
}

// na osnovu korena merkla trazi razliku
func Uporedi(a, b *Cvor) {
	if a == nil || b == nil {
		return
	}
	if a.hes == b.hes {
		return
	}
	//ako je list tu je razlika
	if a.levi == nil && a.desni == nil {
		fmt.Println("Razlika:")
		fmt.Println("A:", a.hes)
		fmt.Println("B:", b.hes)
		return
	}
	//rekurzivno proveravam za oba podstabla
	Uporedi(a.levi, b.levi)
	Uporedi(a.desni, b.desni)
}

func Ispisi(c *Cvor, nivo int) {
	if c == nil {
		return
	}
	for i := 0; i < nivo; i++ {
		fmt.Print("  ")
	}
	fmt.Println(c.hes)
	Ispisi(c.levi, nivo+1)
	Ispisi(c.desni, nivo+1)
}

func Serijalizuj(root *Cvor) []string {
	if root == nil {
		return []string{}
	}

	var rezultat []string
	queue := []*Cvor{root}

	for len(queue) > 0 {
		c := queue[0]
		queue = queue[1:]

		if c == nil {
			rezultat = append(rezultat, "nil")
			continue
		}

		rezultat = append(rezultat, c.hes)

		queue = append(queue, c.levi)
		queue = append(queue, c.desni)
	}

	return rezultat
}

func Deserijalizuj(podaci []string) *Cvor {
	if len(podaci) == 0 {
		return nil
	}

	if podaci[0] == "nil" {
		return nil
	}

	root := &Cvor{hes: podaci[0]}
	queue := []*Cvor{root}
	i := 1

	for i < len(podaci) {
		trenutni := queue[0]
		queue = queue[1:]

		//levi
		if podaci[i] != "nil" {
			trenutni.levi = &Cvor{hes: podaci[i]}
			queue = append(queue, trenutni.levi)
		}
		i++

		if i >= len(podaci) {
			break
		}

		//desni
		if podaci[i] != "nil" {
			trenutni.desni = &Cvor{hes: podaci[i]}
			queue = append(queue, trenutni.desni)
		}
		i++
	}

	return root
}

func Sacuvaj(root *Cvor, filename string) error {
	podaci := Serijalizuj(root)

	fajl, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer fajl.Close()

	tekst := strings.Join(podaci, ",")
	_, err = fajl.WriteString(tekst)
	return err
}

func Ucitaj(filename string) (*Cvor, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	tekst := string(data)
	podaci := strings.Split(tekst, ",")

	root := Deserijalizuj(podaci)
	return root, nil
}
