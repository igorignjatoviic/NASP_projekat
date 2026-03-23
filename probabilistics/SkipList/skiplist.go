package skiplist

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math/rand"
	"os"
	"time"
)

// Tipovi
// =============================

type Cvor struct {
	kljuc     string
	vrednost  []byte
	timestamp int64
	tombstone bool
	sledeci   []*Cvor // sledeci[i] = sledeci cvor na nivou i
}

type SkipLista struct {
	glava          *Cvor
	maksVisina     int // maksimalni indeks nivoa, ukupno nivoa = maksVisina+1
	trenutnaVisina int // najveci nivo koji je trenutno aktivan (indeks)
	brojElemenata  int
	maxElemenata   int
}

type Unos struct {
	vrednost  []byte
	timestamp int64
	tombstone bool
}

// Kreiranje
// =============================

// NovaSkipLista kreira praznu skip listu.
// maksVisina je maksimalni indeks nivoa
func NovaSkipLista(maksVisina int, maxElemenata int) *SkipLista {
	if maksVisina < 1 {
		maksVisina = 16
	}

	// Seed jednom, da bacanje novcica ne bude uvek isto
	rand.Seed(time.Now().UnixNano())

	return &SkipLista{
		glava: &Cvor{
			sledeci: make([]*Cvor, maksVisina+1),
		},
		maksVisina:     maksVisina,
		trenutnaVisina: 0,
		brojElemenata:  0,
		maxElemenata:   maxElemenata,
	}
}

func (s *SkipLista) Duzina() int {
	return s.brojElemenata
}

// Osnovne operacije
// =============================

// Nadji pretrazuje kljuc i vraca (vrednost, true) ako postoji
func (s *SkipLista) Dobavi(kljuc string) ([]byte, bool) {
	x := s.glava

	for nivo := s.trenutnaVisina; nivo >= 0; nivo-- {
		for x.sledeci[nivo] != nil && x.sledeci[nivo].kljuc < kljuc {
			x = x.sledeci[nivo]
		}
	}

	x = x.sledeci[0]
	if x != nil && x.kljuc == kljuc {
		if x.tombstone {
			return nil, false
		}
		return bytes.Clone(x.vrednost), true
	}

	return nil, false
}

// Ubaci dodaje (kljuc, vrednost). Ako kljuc vec postoji, overwrite-uje vrednost
func (s *SkipLista) Ubaci(kljuc string, vrednost []byte) {
	update := make([]*Cvor, s.maksVisina+1)
	x := s.glava

	for nivo := s.trenutnaVisina; nivo >= 0; nivo-- {
		for x.sledeci[nivo] != nil && x.sledeci[nivo].kljuc < kljuc {
			x = x.sledeci[nivo]
		}
		update[nivo] = x
	}

	x = x.sledeci[0]

	if x != nil && x.kljuc == kljuc {
		x.vrednost = bytes.Clone(vrednost)
		x.timestamp = time.Now().UnixNano()
		x.tombstone = false
		return
	}

	noviNivo := s.izvuciNivo()

	if noviNivo > s.trenutnaVisina {
		for nivo := s.trenutnaVisina + 1; nivo <= noviNivo; nivo++ {
			update[nivo] = s.glava
		}
		s.trenutnaVisina = noviNivo
	}

	novi := &Cvor{
		kljuc:     kljuc,
		vrednost:  bytes.Clone(vrednost),
		timestamp: time.Now().UnixNano(),
		tombstone: false,
		sledeci:   make([]*Cvor, noviNivo+1),
	}

	// Prevezi pokazivace na svim nivoima koje novi cvor ima
	for nivo := 0; nivo <= noviNivo; nivo++ {
		novi.sledeci[nivo] = update[nivo].sledeci[nivo]
		update[nivo].sledeci[nivo] = novi
	}

	s.brojElemenata++
}

// Obrisi brise element sa datim kljucem. Vraca true ako je obrisan
func (s *SkipLista) Obrisi(kljuc string) bool {
	update := make([]*Cvor, s.maksVisina+1)
	x := s.glava

	for nivo := s.trenutnaVisina; nivo >= 0; nivo-- {
		for x.sledeci[nivo] != nil && x.sledeci[nivo].kljuc < kljuc {
			x = x.sledeci[nivo]
		}
		update[nivo] = x
	}

	x = x.sledeci[0]
	if x != nil && x.kljuc == kljuc {
		if x.tombstone {
			return false //već obrisan
		}
		x.tombstone = true
		x.timestamp = time.Now().UnixNano()
		return true
	}
	//ukoliko trazeni kljuc ne postoji, ubacujemo tombstone cvor na odgovarajuce mesto
	noviNivo := s.izvuciNivo()

	if noviNivo > s.trenutnaVisina {
		for nivo := s.trenutnaVisina + 1; nivo <= noviNivo; nivo++ {
			update[nivo] = s.glava
		}
	}

	novi := &Cvor{
		kljuc:     kljuc,
		vrednost:  nil,
		timestamp: time.Now().UnixNano(),
		tombstone: true,
		sledeci:   make([]*Cvor, noviNivo+1),
	}

	for nivo := 0; nivo <= noviNivo; nivo++ {
		novi.sledeci[nivo] = update[nivo].sledeci[nivo]
		update[nivo].sledeci[nivo] = novi
	}
	s.brojElemenata++
	return true
}

// Bacanje novcica
// =============================

// izvuciNivo simulira bacanje novcica: dok dobijamo 1, dizemo nivo
// Maksimum je s.maksVisina
func (s *SkipLista) izvuciNivo() int {
	nivo := 0
	for ; rand.Int31n(2) == 1; nivo++ {
		if nivo >= s.maksVisina {
			return nivo
		}
	}
	return nivo
}

// Serijalizacija
// ===========================
func UcitajSkipListu(putanja string) (*SkipLista, error) { //za testiranje, nije neophodno
	f, err := os.Open(putanja)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := bufio.NewReader(f)

	// magic
	magic := make([]byte, 4)
	if _, err := io.ReadFull(r, magic); err != nil {
		return nil, err
	}
	if !bytes.Equal(magic, []byte{'S', 'K', 'I', 'P'}) {
		return nil, errors.New("pogresan format")
	}

	var verzija uint16
	if err := binary.Read(r, binary.BigEndian, &verzija); err != nil {
		return nil, err
	}
	if verzija != 1 {
		return nil, errors.New("nepodrzana verzija formata")
	}

	var maksVisina uint16
	if err := binary.Read(r, binary.BigEndian, &maksVisina); err != nil {
		return nil, err
	}

	var broj uint32
	if err := binary.Read(r, binary.BigEndian, &broj); err != nil {
		return nil, err
	}

	sl := NovaSkipLista(int(maksVisina), 0)

	for i := uint32(0); i < broj; i++ {
		var nivoCvora uint8
		if err := binary.Read(r, binary.BigEndian, &nivoCvora); err != nil {
			return nil, err
		}

		var duzinaK uint32
		if err := binary.Read(r, binary.BigEndian, &duzinaK); err != nil {
			return nil, err
		}
		kb := make([]byte, duzinaK)
		if _, err := io.ReadFull(r, kb); err != nil {
			return nil, err
		}

		var duzinaV uint32
		if err := binary.Read(r, binary.BigEndian, &duzinaV); err != nil {
			return nil, err
		}
		vb := make([]byte, duzinaV)
		if _, err := io.ReadFull(r, vb); err != nil {
			return nil, err
		}

		sl.ubaciSaNivoom(string(kb), vb, int(nivoCvora))
	}

	return sl, nil
}

// ubaciSaNivoom ubacuje cvor sa zadatim nivoom, koristi se pri ucitavanju
func (s *SkipLista) ubaciSaNivoom(kljuc string, vrednost []byte, nivoCvora int) { //za tesiranje, nije neophodno
	if nivoCvora > s.maksVisina {
		nivoCvora = s.maksVisina
	}
	update := make([]*Cvor, s.maksVisina+1)
	x := s.glava

	for nivo := s.trenutnaVisina; nivo >= 0; nivo-- {
		for x.sledeci[nivo] != nil && x.sledeci[nivo].kljuc < kljuc {
			x = x.sledeci[nivo]
		}
		update[nivo] = x
	}

	x = x.sledeci[0]
	if x != nil && x.kljuc == kljuc {
		x.vrednost = bytes.Clone(vrednost)
		return
	}

	if nivoCvora > s.trenutnaVisina {
		for nivo := s.trenutnaVisina + 1; nivo <= nivoCvora; nivo++ {
			update[nivo] = s.glava
		}
		s.trenutnaVisina = nivoCvora
	}

	novi := &Cvor{
		kljuc:    kljuc,
		vrednost: bytes.Clone(vrednost),
		sledeci:  make([]*Cvor, nivoCvora+1),
	}
	for nivo := 0; nivo <= nivoCvora; nivo++ {
		novi.sledeci[nivo] = update[nivo].sledeci[nivo]
		update[nivo].sledeci[nivo] = novi
	}
	s.brojElemenata++
}

func (s *SkipLista) DobaviSve() []Unos {
	var rezultat []Unos

	x := s.glava.sledeci[0]

	if x != nil {
		rezultat = append(rezultat, Unos{
			vrednost:  bytes.Clone(x.vrednost),
			timestamp: x.timestamp,
			tombstone: x.tombstone,
		})
		x = x.sledeci[0]
	}
	return rezultat
}

func (s *SkipLista) Isprazni() {
	s.glava = &Cvor{
		sledeci: make([]*Cvor, s.maksVisina+1),
	}
	s.trenutnaVisina = 0
	s.brojElemenata = 0
}

func (s *SkipLista) DaLiFlush() bool {
	if s.maxElemenata <= 0 {
		return false
	}
	return s.brojElemenata >= s.maxElemenata
}
