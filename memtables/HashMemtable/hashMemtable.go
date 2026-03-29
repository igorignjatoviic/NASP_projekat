package hashmemtable

import (
	"NASP_projekat/memtables"
	"bytes"
	"sort"
	"time"
)

type HashMemtable struct {
	podaci   map[string]memtables.Unos
	maxUnosa int
}

func NapraviHashMemtable(maxUnosa int) *HashMemtable {
	return &HashMemtable{
		podaci:   make(map[string]memtables.Unos),
		maxUnosa: maxUnosa,
	}
}

func (m *HashMemtable) Ubaci(kljuc string, vrednost []byte) {
	m.podaci[kljuc] = memtables.Unos{
		Kljuc:     kljuc,
		Vrednost:  vrednost,
		Timestamp: time.Now().UnixNano(),
		Tombstone: false,
	}
}

func (m *HashMemtable) Dobavi(kljuc string) ([]byte, bool) { //	put
	unos, postoji := m.podaci[kljuc]
	if !postoji || unos.Tombstone {
		return nil, false // ne postoji trazeni kljuc
	}
	return unos.Vrednost, true
}

func (m *HashMemtable) Obrisi(kljuc string) bool { // delete
	unos, postoji := m.podaci[kljuc]
	if postoji {
		if unos.Tombstone {
			return false
		}
		unos.Timestamp = time.Now().UnixNano()
		unos.Tombstone = true
		unos.Vrednost = nil
		m.podaci[kljuc] = unos
		return true
	}

	// ako kljuc ne postoji, upisi tombstone zapis
	m.podaci[kljuc] = memtables.Unos{
		Kljuc:     kljuc,
		Vrednost:  nil,
		Timestamp: time.Now().UnixNano(),
		Tombstone: true,
	}
	return true
}

func (m *HashMemtable) DaLiFlush() bool {
	return len(m.podaci) >= m.maxUnosa
}

func (m *HashMemtable) Isprazni() { // prazni memtabelu
	m.podaci = make(map[string]memtables.Unos)
}

func (m *HashMemtable) DobaviSve() []memtables.Unos {
	kljucevi := make([]string, 0, len(m.podaci))
	for k := range m.podaci {
		kljucevi = append(kljucevi, k)
	}
	sort.Strings(kljucevi)

	rezultat := make([]memtables.Unos, 0, len(kljucevi))
	for _, k := range kljucevi {
		unos := m.podaci[k]
		rezultat = append(rezultat, memtables.Unos{
			Kljuc:     unos.Kljuc,
			Vrednost:  bytes.Clone(unos.Vrednost),
			Timestamp: unos.Timestamp,
			Tombstone: unos.Tombstone,
		})
	}
	return rezultat
}

func (m *HashMemtable) NadjiUnos(kljuc string) (memtables.Unos, bool) { //pomocna metoda za memtable sa vise instanci
	unos, postoji := m.podaci[kljuc]
	if !postoji {
		return memtables.Unos{}, false
	}
	return memtables.Unos{
		Kljuc:     unos.Kljuc,
		Vrednost:  bytes.Clone(unos.Vrednost),
		Timestamp: unos.Timestamp,
		Tombstone: unos.Tombstone,
	}, true
}

func (m *HashMemtable) Duzina() int {
	return len(m.podaci)
}
