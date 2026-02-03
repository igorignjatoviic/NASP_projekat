package main

import (
	"sort"
	"time"
)

type Unos struct {
	vrednost  []byte
	timestamp int64
	tombstone bool
}

type HashMemtable struct {
	podaci   map[string]Unos
	maxUnosa int
}

func NapraviHashMemtable(maxUnosa int) *HashMemtable {
	return &HashMemtable{
		podaci:   make(map[string]Unos),
		maxUnosa: maxUnosa,
	}
}

func (m *HashMemtable) Ubaci(kljuc string, vrednost []byte) {
	m.podaci[kljuc] = Unos{
		vrednost:  vrednost,
		timestamp: time.Now().UnixNano(),
		tombstone: false,
	}
}

func (m *HashMemtable) Dobavi(kljuc string) ([]byte, bool) { //	put
	unos, postoji := m.podaci[kljuc]
	if !postoji || unos.tombstone {
		return nil, false // ne postoji trazeni kljuc
	}
	return unos.vrednost, true
}

func (m *HashMemtable) Obrisi(kljuc string) bool { // delete
	unos, postoji := m.podaci[kljuc]
	if postoji && !unos.tombstone { // ako unos postoji i nije vec obrisan
		unos.timestamp = time.Now().UnixNano()
		unos.tombstone = true
		m.podaci[kljuc] = unos
		return true
	}
	return false // ako kljuc ne postoji ili je logicki obrisan, ne moze ni da se obrise
}

func (m *HashMemtable) DaLiFlush() bool {
	return len(m.podaci) >= m.maxUnosa
}

func (m *HashMemtable) SortirajPoKljucu() ([]string, []Unos) { // sortira memtabelu po kljucevima
	kljucevi := make([]string, 0, len(m.podaci))
	for k := range m.podaci {
		kljucevi = append(kljucevi, k)
	}
	sort.Strings(kljucevi)
	unosi := make([]Unos, len(kljucevi))
	for i, k := range kljucevi {
		unosi[i] = m.podaci[k]
	}
	return kljucevi, unosi
}

func (m *HashMemtable) Isprazni() { // prazni memtabelu
	m.podaci = make(map[string]Unos)
}

func (m *HashMemtable) DobaviSve() []string {
	kljucevi := make([]string, 0, len(m.podaci))
	for k := range m.podaci {
		kljucevi = append(kljucevi, k)
	}
	return kljucevi
}
