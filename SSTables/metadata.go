package SSTables

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"
)

// Merkle stablo za proveru integriteta podataka
// Format fajla: broj cvorova pa svaki cvor - velicina hasha + hash

type MerkleCvor struct {
	hash  string
	levi  *MerkleCvor
	desni *MerkleCvor
}

type MerkleStabloSST struct {
	koren *MerkleCvor
}

func merkleHash(data string) string {
	h := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", h)
}

// pravi merkle stablo od vrednosti zapisa
func NapraviMerkleStablo(zapisi []*SSTableRecord) *MerkleStabloSST {
	if len(zapisi) == 0 {
		return &MerkleStabloSST{koren: nil}
	}

	// Pravim listove -hash svake vrednosti
	cvorovi := make([]*MerkleCvor, 0, len(zapisi))
	for _, z := range zapisi {
		h := merkleHash(string(z.Value))
		cvorovi = append(cvorovi, &MerkleCvor{hash: h})
	}

	// Ako je neparan broj, dodaje prazan cvor- stablo mora biti balansirano
	if len(cvorovi)%2 != 0 {
		cvorovi = append(cvorovi, &MerkleCvor{hash: merkleHash("")})
	}

	// stablo se gradi odozdo prema gore
	for len(cvorovi) > 1 {
		var noviNivo []*MerkleCvor
		for i := 0; i < len(cvorovi); i += 2 {
			levi := cvorovi[i]
			desni := cvorovi[i+1]
			roditelj := &MerkleCvor{
				hash:  merkleHash(levi.hash + desni.hash),
				levi:  levi,
				desni: desni,
			}
			noviNivo = append(noviNivo, roditelj)
		}
		if len(noviNivo)%2 != 0 && len(noviNivo) != 1 {
			noviNivo = append(noviNivo, &MerkleCvor{hash: merkleHash("")})
		}
		cvorovi = noviNivo
	}

	return &MerkleStabloSST{koren: cvorovi[0]}
}

// poredi dva stabla i vraca gde su razlike
func ValidacijaRazlika(a, b *MerkleCvor, rezultat *[]string) {
	if a == nil || b == nil {
		return
	}
	if a.hash == b.hash {
		return
	}
	// Ako je list, tu je razlika
	if a.levi == nil && a.desni == nil {
		*rezultat = append(*rezultat, fmt.Sprintf("Razlika: %s vs %s", a.hash, b.hash))
		return
	}
	ValidacijaRazlika(a.levi, b.levi, rezultat)
	ValidacijaRazlika(a.desni, b.desni, rezultat)
}

// cuva Merkle stablo na disku
type MetadataSegment struct {
	putanja   string
	blockSize uint64
}

func NoviMetadataSegment(putanja string, blockSize uint64) *MetadataSegment {
	return &MetadataSegment{putanja: putanja, blockSize: blockSize}
}

// upisuje merkle stablo na disk
func (m *MetadataSegment) Upisi(zapisi []*SSTableRecord) error {
	stablo := NapraviMerkleStablo(zapisi)
	return m.serijalizuj(stablo.koren)
}

// funkcija proverava da li su podaci osteceni
func (m *MetadataSegment) Validiraj(zapisi []*SSTableRecord) ([]string, error) {
	// Ucitamo sacuvano stablo
	sacuvanKoren, err := m.deserijalizuj()
	if err != nil {
		return nil, fmt.Errorf("ne mogu da ucitam metadata fajl: %w", err)
	}

	// Napravimo novo stablo od trenutnih podataka
	novoStablo := NapraviMerkleStablo(zapisi)

	var razlike []string
	ValidacijaRazlika(sacuvanKoren, novoStablo.koren, &razlike)
	return razlike, nil
}

// Serijalizacija- BFS redosled

func (m *MetadataSegment) serijalizuj(koren *MerkleCvor) error {
	fajl, err := os.Create(m.putanja)
	if err != nil {
		return err
	}
	defer fajl.Close()

	// BFS obilazak - skupljam sve hashove
	var hashovi []string
	red := []*MerkleCvor{koren}
	for len(red) > 0 {
		c := red[0]
		red = red[1:]
		if c == nil {
			hashovi = append(hashovi, "")
			continue
		}
		hashovi = append(hashovi, c.hash)
		red = append(red, c.levi)
		red = append(red, c.desni)
	}

	// upisujem broj cvorova pa svaki hash
	blok := make([]byte, 0, m.blockSize)

	brojCvorova := uint64(len(hashovi))
	brBuf := make([]byte, 8)
	binary.BigEndian.PutUint64(brBuf, brojCvorova)
	blok = append(blok, brBuf...)

	for _, h := range hashovi {
		hashBytes := []byte(h)
		hashSize := uint64(len(hashBytes))

		sizeBuf := make([]byte, 8)
		binary.BigEndian.PutUint64(sizeBuf, hashSize)
		blok = append(blok, sizeBuf...)
		blok = append(blok, hashBytes...)

		if uint64(len(blok)) >= m.blockSize {
			if _, err := fajl.Write(blok); err != nil {
				return err
			}
			blok = blok[:0]
		}
	}

	if len(blok) > 0 {
		_, err = fajl.Write(blok)
		return err
	}
	return nil
}

// ucitava merkle stablo sa diska
func (m *MetadataSegment) deserijalizuj() (*MerkleCvor, error) {
	fajl, err := os.Open(m.putanja)
	if err != nil {
		return nil, err
	}
	defer fajl.Close()

	var brojCvorova uint64
	if err := binary.Read(fajl, binary.BigEndian, &brojCvorova); err != nil {
		return nil, err
	}

	hashovi := make([]string, 0, brojCvorova)
	for i := uint64(0); i < brojCvorova; i++ {
		var hashSize uint64
		if err := binary.Read(fajl, binary.BigEndian, &hashSize); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		hashBytes := make([]byte, hashSize)
		if hashSize > 0 {
			if _, err := io.ReadFull(fajl, hashBytes); err != nil {
				return nil, err
			}
		}
		hashovi = append(hashovi, string(hashBytes))
	}

	return rekonstruisiStablo(hashovi), nil
}

// pravi stablo iz BFS liste hashova
func rekonstruisiStablo(hashovi []string) *MerkleCvor {
	if len(hashovi) == 0 || hashovi[0] == "" {
		return nil
	}

	koren := &MerkleCvor{hash: hashovi[0]}
	red := []*MerkleCvor{koren}
	i := 1

	for i < len(hashovi) {
		trenutni := red[0]
		red = red[1:]

		if i < len(hashovi) {
			if hashovi[i] != "" {
				trenutni.levi = &MerkleCvor{hash: hashovi[i]}
				red = append(red, trenutni.levi)
			}
			i++
		}
		if i < len(hashovi) {
			if hashovi[i] != "" {
				trenutni.desni = &MerkleCvor{hash: hashovi[i]}
				red = append(red, trenutni.desni)
			}
			i++
		}
	}

	return koren
}

// ispis stabla za proveru
func IspisiStablo(c *MerkleCvor, nivo int) {
	if c == nil {
		return
	}
	fmt.Println(strings.Repeat("  ", nivo) + c.hash)
	IspisiStablo(c.levi, nivo+1)
	IspisiStablo(c.desni, nivo+1)
}
