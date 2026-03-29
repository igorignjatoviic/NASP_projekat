package memtableinterface

import (
	"NASP_projekat/memtables"
	"sort"
)

type Instanca struct { //jedna instanca je jedna tabela
	Podaci        memtables.Struktura
	samoZaCitanje bool
}

type Memtable struct {
	tabele       []*Instanca //struktura se sastoji od vise tabela, od kojih je samo jedna aktivna, a ostale su read-only
	maxTabela    int
	tip          string
	maxElemenata int
	maxVisina    int
	redBStabla   int
}

func NovaMemtable(tip string, maxTabela int, maxElemenata int, maxVisina int, redBStabla int) *Memtable {
	prva := &Instanca{
		Podaci:        NovaStruktura(tip, maxElemenata, maxVisina, redBStabla),
		samoZaCitanje: false,
	}
	return &Memtable{
		tabele:       []*Instanca{prva},
		maxTabela:    maxTabela,
		tip:          tip,
		maxElemenata: maxElemenata,
		maxVisina:    maxVisina,
		redBStabla:   redBStabla,
	}
}

func (m *Memtable) aktivnaTabela() *Instanca {
	return m.tabele[len(m.tabele)-1]
}

func (m *Memtable) Ubaci(kljuc string, vrednost []byte) bool { //povratna vrednost nam govori da li memtable treba da uradi flush
	aktivna := m.aktivnaTabela()
	aktivna.Podaci.Ubaci(kljuc, vrednost)

	if aktivna.Podaci.DaLiFlush() {
		aktivna.samoZaCitanje = true

		if len(m.tabele) < m.maxTabela {
			nova := &Instanca{
				Podaci:        NovaStruktura(m.tip, m.maxElemenata, m.maxVisina, m.redBStabla),
				samoZaCitanje: false,
			}
			m.tabele = append(m.tabele, nova)
			return false
		}
		return true
	}
	return false
}

func (m *Memtable) Dobavi(kljuc string) ([]byte, bool) { //ide od najnovije ka najstrijoj tabeli
	for i := len(m.tabele) - 1; i >= 0; i-- {
		unos, postoji := m.tabele[i].Podaci.NadjiUnos(kljuc)
		if postoji {
			if unos.Tombstone {
				return nil, false
			}
			return unos.Vrednost, true
		}
	}
	return nil, false
}

func (m *Memtable) Obrisi(kljuc string) bool {
	aktivna := m.aktivnaTabela()
	aktivna.Podaci.Obrisi(kljuc)

	if aktivna.Podaci.DaLiFlush() {
		aktivna.samoZaCitanje = true

		if len(m.tabele) < m.maxTabela {
			nova := &Instanca{
				Podaci:        NovaStruktura(m.tip, m.maxElemenata, m.maxVisina, m.redBStabla),
				samoZaCitanje: false,
			}
			m.tabele = append(m.tabele, nova)
			return false
		}
		return true
	}
	return false
}

func (m *Memtable) NadjiUnos(kljuc string) (memtables.Unos, bool) {
	for i := len(m.tabele) - 1; i >= 0; i-- {
		unos, postoji := m.tabele[i].Podaci.NadjiUnos(kljuc)
		if postoji {
			return unos, true
		}
	}
	return memtables.Unos{}, false
}

func (m *Memtable) DobaviSve() []memtables.Unos { //kada treba flush-ovati sve tabele, mora da se dobije jedan konacan skup zapisa
	mapa := make(map[string]memtables.Unos) //iako isti kljuc moze da se nadje u vise tabela, mora da se sacuva najnovija vrezija

	for i := 0; i < len(m.tabele); i++ { //noviji zapis pregazi stariji
		zapisi := m.tabele[i].Podaci.DobaviSve()
		for _, z := range zapisi {
			mapa[z.Kljuc] = z
		}
	}
	rezultat := make([]memtables.Unos, 0, len(mapa))
	for _, z := range mapa {
		rezultat = append(rezultat, z)
	}
	sort.Slice(rezultat, func(i, j int) bool {
		return rezultat[i].Kljuc < rezultat[j].Kljuc
	})
	return rezultat
}

func (m *Memtable) Duzina() int {
	ukupno := 0
	for _, t := range m.tabele {
		ukupno += t.Podaci.Duzina()
	}
	return ukupno
}

func (m *Memtable) DaLiFlush() bool {
	return len(m.tabele) == m.maxTabela && m.aktivnaTabela().Podaci.DaLiFlush()
}

func (m *Memtable) Isprazni() {
	prva := &Instanca{
		Podaci:        NovaStruktura(m.tip, m.maxElemenata, m.maxVisina, m.redBStabla),
		samoZaCitanje: false,
	}
	m.tabele = []*Instanca{prva}
}
