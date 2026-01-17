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
	kljuc    string
	vrednost []byte
	sledeci  []*Cvor // sledeci[i] = sledeci cvor na nivou i
}

type SkipLista struct {
	glava          *Cvor
	maksVisina     int // maksimalni indeks nivoa, ukupno nivoa = maksVisina+1
	trenutnaVisina int // najveci nivo koji je trenutno aktivan (indeks)
	brojElemenata  int
}

// Kreiranje
// =============================

// NovaSkipLista kreira praznu skip listu.
// maksVisina je maksimalni indeks nivoa
func NovaSkipLista(maksVisina int) *SkipLista {
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
	}
}

func (s *SkipLista) Duzina() int {
	return s.brojElemenata
}

// Osnovne operacije
// =============================

// Nadji pretrazuje kljuc i vraca (vrednost, true) ako postoji
func (s *SkipLista) Nadji(kljuc string) ([]byte, bool) {
	x := s.glava

	for nivo := s.trenutnaVisina; nivo >= 0; nivo-- {
		for x.sledeci[nivo] != nil && x.sledeci[nivo].kljuc < kljuc {
			x = x.sledeci[nivo]
		}
	}

	x = x.sledeci[0]
	if x != nil && x.kljuc == kljuc {
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
		kljuc:    kljuc,
		vrednost: bytes.Clone(vrednost),
		sledeci:  make([]*Cvor, noviNivo+1),
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
	if x == nil || x.kljuc != kljuc {
		return false
	}

	// Prevezivanje, preskace se x
	for nivo := 0; nivo <= s.trenutnaVisina; nivo++ {
		if update[nivo].sledeci[nivo] != x {
			break
		}
		update[nivo].sledeci[nivo] = x.sledeci[nivo]
	}

	for s.trenutnaVisina > 0 && s.glava.sledeci[s.trenutnaVisina] == nil {
		s.trenutnaVisina--
	}

	s.brojElemenata--
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
// =============================

// Nivo cvora cuvamo da bismo pri ucitavanju rekonstruisali istu strukturu
func (s *SkipLista) Sacuvaj(putanja string) error {
	f, err := os.Create(putanja)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	defer w.Flush()

	// header
	if _, err := w.Write([]byte{'S', 'K', 'I', 'P'}); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, uint16(1)); err != nil { // verzija
		return err
	}
	if err := binary.Write(w, binary.BigEndian, uint16(s.maksVisina)); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, uint32(s.brojElemenata)); err != nil {
		return err
	}

	// elementi
	for x := s.glava.sledeci[0]; x != nil; x = x.sledeci[0] {
		nivoCvora := uint8(len(x.sledeci) - 1)
		if err := binary.Write(w, binary.BigEndian, nivoCvora); err != nil {
			return err
		}

		kb := []byte(x.kljuc)
		if err := binary.Write(w, binary.BigEndian, uint32(len(kb))); err != nil {
			return err
		}
		if _, err := w.Write(kb); err != nil {
			return err
		}

		vb := x.vrednost
		if err := binary.Write(w, binary.BigEndian, uint32(len(vb))); err != nil {
			return err
		}
		if _, err := w.Write(vb); err != nil {
			return err
		}
	}

	return nil
}

func UcitajSkipListu(putanja string) (*SkipLista, error) {
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

	sl := NovaSkipLista(int(maksVisina))

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
func (s *SkipLista) ubaciSaNivoom(kljuc string, vrednost []byte, nivoCvora int) {
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
