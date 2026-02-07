package main

import "fmt"

//prvo pravimo izgled za mesto u memoriji u kome se nalzi blok
type Frejm struct {
	id     int    //id stranice
	menjan bool   //da li je menjan
	brKor  int    //broj korisnika
	pod    []byte //sadrzaj stranice
}

type BufferPool struct {
	velicina int         //velicina poola
	frejmovi []Frejm     //niz u koji upisujemo blokove
	mapa     map[int]int //za brzu pretragu stranica-gde je vrednost index u frame
	slob     int         //koliko imamo prostora- da li treba flush
}

//kad pravimo prazan bp prosledjujemo velicinu pool i velicinu stranica/blokova
func Napravi(velPool int, velStr int) *BufferPool {
	frejmovi := make([]Frejm, velPool)
	//ovo moramo da popunimo praznu staranu u frame
	for i := 0; i < velPool; i++ {
		frejmovi[i] = Frejm{
			id:     -1,
			menjan: false,
			brKor:  0,
			pod:    make([]byte, velStr),
		}
	}

	return &BufferPool{
		velicina: velPool,
		frejmovi: frejmovi,
		mapa:     make(map[int]int),
		slob:     velPool,
	}
}

//upis-upise na kraj ako je popunjen vrati inf da je popunjen
//citanje-nadje ga-izbaci ga i upise na kraj-vrati ga

func (bp *BufferPool) Upisi(id int, podaci []byte) {
	//ako vec postoji
	if _, ok := bp.mapa[id]; ok {
		fmt.Printf("Stranica vec postoji")
		return
	}
	//da li je pun
	if bp.slob == 0 {
		fmt.Printf("Buffer pool je pun uradi flush. Hvala Milice!")
		return
	}
	//inace nadji prvi slobodan i upisi
	for i := 0; i < bp.velicina; i++ {
		if bp.frejmovi[i].id == -1 {
			//dodajemo u frejm
			bp.frejmovi[i] = Frejm{
				id:     id,
				menjan: false, //mozda treba true
				brKor:  1,
				pod:    podaci}
			//dodajemo u mapu
			bp.mapa[id] = i
			//smanjujemo br slob mesta
			bp.slob--
			break
		}
	}
}

//ispravljena verzija
func (bp *BufferPool) Citaj(id int) []byte {
	//nadjemo ga u mapi
	inx, postoji := bp.mapa[id]
	if !postoji {
		fmt.Printf("Ne postoji ta stranica.")
		return nil
	}
	//taj frejm stavi u pomocnu prom
	pom := bp.frejmovi[inx]
	delete(bp.mapa, pom.id)
	//pomeri ulevo i izmenim u mapi
	korisceni := bp.velicina - bp.slob
	for i := inx; i < korisceni-1; i++ {
		bp.frejmovi[i] = bp.frejmovi[i+1]
		bp.mapa[bp.frejmovi[i].id] = i
	}
	//upisi na kraj
	posl := korisceni - 1
	bp.frejmovi[posl] = pom
	bp.mapa[pom.id] = posl

	//broj koriscenja se povecava
	bp.frejmovi[posl].brKor++
	return pom.pod

}

//oslobadjanje stranice
func (bp *BufferPool) Oslobodi(id int) {
	if idx, ok := bp.mapa[id]; ok {
		if bp.frejmovi[idx].brKor > 0 {
			bp.frejmovi[idx].brKor--
		}
	}
}

//oznaci da je izmenjen
func (bp *BufferPool) oznaciMenjanje(id int) {
	if idx, ok := bp.mapa[id]; ok {
		bp.frejmovi[idx].menjan = true
	}
}

//kad se pozove citanje pincount kaze da ovu stranicu neko koristi
//to znaci da se ne sme izbaciti dok ne zavrsi s atom stranicom
