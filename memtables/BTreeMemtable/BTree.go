package btree

import (
	"NASP_projekat/memtables"
	"bytes"
	"time"
)

type BTreeNode struct {
	kljucevi   []string
	vrednosti  [][]byte
	timestamps []int64
	tombstones []bool
	deca       []*BTreeNode
	list       bool
}

type BTree struct {
	root          *BTreeNode
	m             int //red stabla
	maxElemenata  int
	brojElemenata int
}

func NovoBStablo(m int, maxElemenata int) *BTree {
	if m < 3 {
		m = 3
	}

	root := &BTreeNode{
		kljucevi:   []string{},
		vrednosti:  [][]byte{},
		timestamps: []int64{},
		tombstones: []bool{},
		deca:       []*BTreeNode{},
		list:       true,
	}

	return &BTree{
		root:          root,
		m:             m,
		maxElemenata:  maxElemenata,
		brojElemenata: 0,
	}
}

func (b *BTree) maxKljuceva() int { //pomocna funkcija za maksimalno kljuceva u stablu reda m
	return b.m - 1
}

func DobaviIzCvora(cvor *BTreeNode, kljuc string) ([]byte, bool) {
	i := 0
	for i < len(cvor.kljucevi) && kljuc > cvor.kljucevi[i] {
		i++
	}

	if i < len(cvor.kljucevi) && kljuc == cvor.kljucevi[i] {
		if cvor.tombstones[i] {
			return nil, false
		}
		return bytes.Clone(cvor.vrednosti[i]), true
	}

	if cvor.list {
		return nil, false
	}

	return DobaviIzCvora(cvor.deca[i], kljuc)
}

func (b *BTree) Dobavi(kljuc string) ([]byte, bool) {
	return DobaviIzCvora(b.root, kljuc)
}

func (b *BTree) NadjiCvor(kljuc string) (*BTreeNode, int, bool) { //pomocna funkcija za Ubaci, ukoliko kljuc vec postoji, vraca cvor i indeks
	return NadjiCvorRekurzivno(b.root, kljuc)
}

func NadjiCvorRekurzivno(cvor *BTreeNode, kljuc string) (*BTreeNode, int, bool) {
	i := 0
	for i < len(cvor.kljucevi) && kljuc > cvor.kljucevi[i] {
		i++
	}
	if i < len(cvor.kljucevi) && kljuc == cvor.kljucevi[i] {
		return cvor, i, true
	}
	if cvor.list {
		return nil, -1, false
	}
	return NadjiCvorRekurzivno(cvor.deca[i], kljuc)
}

func (b *BTree) Ubaci(kljuc string, vrednost []byte) {
	if cvor, idx, postoji := b.NadjiCvor(kljuc); postoji {
		cvor.vrednosti[idx] = bytes.Clone(vrednost)
		cvor.timestamps[idx] = time.Now().UnixNano()
		cvor.tombstones[idx] = false
		return
	}

	if len(b.root.kljucevi) == b.maxKljuceva() {
		stariRoot := b.root
		noviRoot := &BTreeNode{
			kljucevi:   []string{},
			vrednosti:  [][]byte{},
			timestamps: []int64{},
			tombstones: []bool{},
			deca:       []*BTreeNode{stariRoot},
			list:       false,
		}
		b.root = noviRoot
		b.podeliDete(noviRoot, 0)
	}
	b.UbaciUNepuniCvor(b.root, kljuc, vrednost)
	b.brojElemenata++
}

func (b *BTree) podeliDete(roditelj *BTreeNode, index int) {
	dete := roditelj.deca[index]
	mid := len(dete.kljucevi) / 2

	desni := &BTreeNode{ //pravimo novi desni cvor
		kljucevi:   append([]string{}, dete.kljucevi[mid+1:]...),
		vrednosti:  append([][]byte{}, dete.vrednosti[mid+1:]...),
		timestamps: append([]int64{}, dete.timestamps[mid+1:]...),
		tombstones: append([]bool{}, dete.tombstones[mid+1:]...),
		list:       dete.list,
	}

	if !dete.list { //ukoliko nije list, delimo njegovu decu na dva dela
		desni.deca = append([]*BTreeNode{}, dete.deca[mid+1:]...)
	}

	srednjiKljuc := dete.kljucevi[mid] //izdvajamo srednji kljuc iz pocetnog deteta
	srednjaVrednost := dete.vrednosti[mid]
	srednjiTimestamp := dete.timestamps[mid]
	srednjiTombstone := dete.tombstones[mid]

	dete.kljucevi = dete.kljucevi[:mid] //sredjujemo levi cvor
	dete.vrednosti = dete.vrednosti[:mid]
	dete.timestamps = dete.timestamps[:mid]
	dete.tombstones = dete.tombstones[:mid]

	if !dete.list { //ukoliko nije list, delimo njegovu decu na dva dela
		dete.deca = dete.deca[:mid+1]
	}

	roditelj.kljucevi = append(roditelj.kljucevi, "") //pravimo prazno mesto u roditeljskom cvoru da bismo mogli da ispomeramo element za jedno mesto udesno prilikom ubacivanja novog kljuca
	roditelj.vrednosti = append(roditelj.vrednosti, nil)
	roditelj.timestamps = append(roditelj.timestamps, 0)
	roditelj.tombstones = append(roditelj.tombstones, false)
	roditelj.deca = append(roditelj.deca, nil)

	copy(roditelj.kljucevi[index+1:], roditelj.kljucevi[index:]) //pomeramo elemente za jedno mesto udesno
	copy(roditelj.vrednosti[index+1:], roditelj.vrednosti[index:])
	copy(roditelj.timestamps[index+1:], roditelj.timestamps[index:])
	copy(roditelj.tombstones[index+1:], roditelj.tombstones[index:])
	copy(roditelj.deca[index+2:], roditelj.deca[index+1:])

	roditelj.kljucevi[index] = srednjiKljuc //ubacujemo srednji kljuc iz deteta u roditelja na index
	roditelj.vrednosti[index] = srednjaVrednost
	roditelj.timestamps[index] = srednjiTimestamp
	roditelj.tombstones[index] = srednjiTombstone
	roditelj.deca[index+1] = desni
}

func (b *BTree) UbaciUNepuniCvor(cvor *BTreeNode, kljuc string, vrednost []byte) {
	i := len(cvor.kljucevi) - 1

	if cvor.list {
		cvor.kljucevi = append(cvor.kljucevi, "")
		cvor.vrednosti = append(cvor.vrednosti, nil)
		cvor.timestamps = append(cvor.timestamps, 0)
		cvor.tombstones = append(cvor.tombstones, false)

		for i >= 0 && kljuc < cvor.kljucevi[i] {
			cvor.kljucevi[i+1] = cvor.kljucevi[i]
			cvor.vrednosti[i+1] = cvor.vrednosti[i]
			cvor.timestamps[i+1] = cvor.timestamps[i]
			cvor.tombstones[i+1] = cvor.tombstones[i]
			i--
		}

		cvor.kljucevi[i+1] = kljuc
		cvor.vrednosti[i+1] = bytes.Clone(vrednost)
		cvor.timestamps[i+1] = time.Now().UnixNano()
		cvor.tombstones[i+1] = false
		return
	}

	for i >= 0 && kljuc < cvor.kljucevi[i] { //ukoliko cvor nije list, pozicioniramo se na odgovarajuce dete
		i--
	}
	i++

	if len(cvor.deca[i].kljucevi) == b.maxKljuceva() { //ukoliko je odgovarajuce dete popunjeno, uradi podelu
		b.podeliDete(cvor, i)

		if kljuc > cvor.kljucevi[i] { //posto se podelom struktura promenila, opet se pozicioniraj na odgovarajuce dete
			i++
		}
	}
	b.UbaciUNepuniCvor(cvor.deca[i], kljuc, vrednost) //rekurzivno se kreci kroz stablo dok ne naidjes na list u koji moze da se ubaci
}

func (b *BTree) Obrisi(kljuc string) bool {
	if cvor, idx, postoji := b.NadjiCvor(kljuc); postoji {
		if cvor.tombstones[idx] {
			return false
		}
		cvor.tombstones[idx] = true
		cvor.timestamps[idx] = time.Now().UnixNano()
		return true
	}
	b.Ubaci(kljuc, nil)
	cvor, idx, postoji := b.NadjiCvor(kljuc)
	if !postoji {
		return false
	}
	cvor.tombstones[idx] = true
	cvor.timestamps[idx] = time.Now().UnixNano()
	return true
}

func (b *BTree) DobaviSve() []memtables.Unos {
	var rezultat []memtables.Unos
	poredjaj(b.root, &rezultat)
	return rezultat
}

func poredjaj(cvor *BTreeNode, rezultat *[]memtables.Unos) { //vrati sortirano
	if cvor == nil {
		return
	}
	for i := 0; i < len(cvor.kljucevi); i++ {
		if !cvor.list {
			poredjaj(cvor.deca[i], rezultat)
		}
		*rezultat = append(*rezultat, memtables.Unos{
			Kljuc:     cvor.kljucevi[i],
			Vrednost:  bytes.Clone(cvor.vrednosti[i]),
			Timestamp: cvor.timestamps[i],
			Tombstone: cvor.tombstones[i],
		})
	}
	if !cvor.list {
		poredjaj(cvor.deca[len(cvor.deca)-1], rezultat)
	}
}

func (b *BTree) Duzina() int {
	return b.brojElemenata
}

func (b *BTree) DaLiFlush() bool {
	return b.brojElemenata >= b.maxElemenata
}

func (b *BTree) Isprazni() {
	b.root = &BTreeNode{
		kljucevi:   []string{},
		vrednosti:  [][]byte{},
		timestamps: []int64{},
		tombstones: []bool{},
		list:       true,
	}
	b.brojElemenata = 0
}

func (b *BTree) NadjiUnos(kljuc string) (memtables.Unos, bool) { //pomocna metoda za memtable sa vise instanci
	return b.NadjiUnosIzCvora(b.root, kljuc)
}

func (b *BTree) NadjiUnosIzCvora(cvor *BTreeNode, kljuc string) (memtables.Unos, bool) {
	i := 0
	for i < len(cvor.kljucevi) && kljuc > cvor.kljucevi[i] {
		i++
	}

	if i < len(cvor.kljucevi) && kljuc == cvor.kljucevi[i] {
		return memtables.Unos{
			Kljuc:     cvor.kljucevi[i],
			Vrednost:  bytes.Clone(cvor.vrednosti[i]),
			Timestamp: cvor.timestamps[i],
			Tombstone: cvor.tombstones[i],
		}, true
	}

	if cvor.list {
		return memtables.Unos{}, false
	}

	return b.NadjiUnosIzCvora(cvor.deca[i], kljuc)
}
