package bufferpool

import (
	bufferpool "NASP_projekat/BufferPool"
	configuration "NASP_projekat/Configuration"
	"encoding/binary"
	"fmt"
	"os"
)

type Block struct {
	kapacitet uint64 // maks broj tupleova u bloku
	Podaci    []bufferpool.Tuple
	id        uint64 // indeks bloka
}

type BlockManager struct {
	bp            *bufferpool.BufferPool
	Blokovi       []Block
	velicinaBloka uint64 // broj tupleova po bloku
	putanja       string // fajl gde se cuvaju blokovi
}

func NoviBlockManager(bp *bufferpool.BufferPool, putanja string) *BlockManager {
	konfiguracija := configuration.UcitajKonfiguraciju()
	velicinaBloka := uint64(10) // podrazumevano ako nema u konfiguraciji
	if vrednost, postoji := konfiguracija["BlockManager"]["Velicina bloka"]; postoji {
		velicinaBloka = uint64(vrednost)
	}
	bm := &BlockManager{
		bp:            bp,
		velicinaBloka: velicinaBloka,
		putanja:       putanja,
		Blokovi:       []Block{},
	}
	bm.KreirajBlokove()
	return bm
}

// deli BufferPool na blokove zadate velicine
func (bm *BlockManager) KreirajBlokove() { // poziva se pri kreiranju block managera
	if len(bm.bp.Podaci) == 0 {
		return
	}
	ukupnoBlokova := (uint64(len(bm.bp.Podaci)) + bm.velicinaBloka - 1) / bm.velicinaBloka
	for i := uint64(0); i < ukupnoBlokova; i++ {
		pocetak := i * bm.velicinaBloka // pocetak i kraj bloka
		kraj := pocetak + bm.velicinaBloka
		if kraj > uint64(len(bm.bp.Podaci)) {
			kraj = uint64(len(bm.bp.Podaci))
		}
		blok := Block{
			kapacitet: bm.velicinaBloka,
			Podaci:    bm.bp.Podaci[pocetak:kraj],
			id:        i,
		}
		bm.Blokovi = append(bm.Blokovi, blok)
	}
}

func (bm *BlockManager) UpisiBlok(idBloka int) error { // upis bloka na disk
	if idBloka < 0 || idBloka >= len(bm.Blokovi) {
		return fmt.Errorf("Id bloka %d ne postoji", idBloka)
	}
	blok := bm.Blokovi[idBloka]
	imeFajla := fmt.Sprintf("%s/block_%d.bin", bm.putanja, idBloka)
	fajl, err := os.Create(imeFajla)
	if err != nil {
		return err
	}
	defer fajl.Close()

	if err := binary.Write(fajl, binary.BigEndian, blok.kapacitet); err != nil {
		return err
	}

	brojTupleova := uint64(len(blok.Podaci))

	if err := binary.Write(fajl, binary.BigEndian, brojTupleova); err != nil {
		return err
	}

	for _, tuple := range blok.Podaci {
		tombstone := uint8(0) // odredjujemo tip dogadjaja (tombstone)
		if tuple.Dogadjaj == "delete" {
			tombstone = 1
		}

		if err := binary.Write(fajl, binary.BigEndian, tombstone); err != nil {
			return err
		}

		if err := binary.Write(fajl, binary.BigEndian, tuple.DuzinaKljuca); err != nil {
			return err
		}

		if err := binary.Write(fajl, binary.BigEndian, tuple.DuzinaVrednosti); err != nil {
			return err
		}

		kljucBajtovi := []byte(tuple.Kljuc)
		if err := binary.Write(fajl, binary.BigEndian, kljucBajtovi); err != nil {
			return err
		}

		vrednostBajtovi := []byte(tuple.Vrednost)
		if err := binary.Write(fajl, binary.BigEndian, vrednostBajtovi); err != nil {
			return err
		}
	}

	return nil
}

func (bm *BlockManager) UpisiSveBlokove() error { // upis svih blokova na disk
	for i := range bm.Blokovi {
		if err := bm.UpisiBlok(i); err != nil {
			return fmt.Errorf("Greska pri upisu bloka %d: %v", i, err)
		}
	}
	return nil
}

func (bm *BlockManager) UcitajBlok(idBloka int) (*Block, error) { // ucitava blok sa diska
	imeFajla := fmt.Sprintf("%s/block_%d.bin", bm.putanja, idBloka)

	fajl, err := os.Open(imeFajla)
	if err != nil {
		return nil, err
	}
	defer fajl.Close()

	var kapacitet uint64
	if err := binary.Read(fajl, binary.BigEndian, &kapacitet); err != nil {
		return nil, err
	}

	var brojTupleova uint64
	if err := binary.Read(fajl, binary.BigEndian, &brojTupleova); err != nil {
		return nil, err
	}

	podaci := make([]bufferpool.Tuple, brojTupleova)

	for i := uint64(0); i < brojTupleova; i++ {
		var tombstone uint8
		if err := binary.Read(fajl, binary.BigEndian, &tombstone); err != nil {
			return nil, err
		}

		var duzinaKljuca uint64
		if err := binary.Read(fajl, binary.BigEndian, &duzinaKljuca); err != nil {
			return nil, err
		}

		var duzinaVrednosti uint64
		if err := binary.Read(fajl, binary.BigEndian, &duzinaVrednosti); err != nil {
			return nil, err
		}

		kljucBajtovi := make([]byte, duzinaKljuca)
		if err := binary.Read(fajl, binary.BigEndian, &kljucBajtovi); err != nil {
			return nil, err
		}

		vrednostBajtovi := make([]byte, duzinaVrednosti)
		if err := binary.Read(fajl, binary.BigEndian, &vrednostBajtovi); err != nil {
			return nil, err
		}

		// odredjivanje dogadjaja na osnovu tombstonea
		dogadjaj := "put"
		if tombstone == 1 {
			dogadjaj = "delete"
		}

		podaci[i] = *bufferpool.NoviTuple(dogadjaj, string(kljucBajtovi), string(vrednostBajtovi))
	}

	blok := &Block{
		kapacitet: kapacitet,
		Podaci:    podaci,
		id:        uint64(idBloka),
	}

	return blok, nil
}

func (bm *BlockManager) OsveziBlokove() { // kada se buffer pool promeni, ponovo se kreiraju svi blokovi
	bm.Blokovi = []Block{}
	bm.KreirajBlokove()
}

func (bm *BlockManager) PronadjiBlok(kljuc string) (int, int) { // pronalazi blok u kom se nalazi kljuc
	for i, blok := range bm.Blokovi {
		for j, tuple := range blok.Podaci {
			if tuple.Kljuc == kljuc {
				return i, j
			}
		}
	}
	return -1, -1
}

func TestBlockManager() {
	bp := bufferpool.NoviBufferPool()
	testPodaci := []bufferpool.Tuple{
		*bufferpool.NoviTuple("put", "pera", "Peric"),
		*bufferpool.NoviTuple("put", "mika", "Mikic"),
		*bufferpool.NoviTuple("put", "zika", "Zikic"),
		*bufferpool.NoviTuple("delete", "mika", ""),
		*bufferpool.NoviTuple("put", "laza", "Lazic"),
	}
	err := bp.Unesi(testPodaci)
	if err != nil {
		fmt.Printf("Greska pri upisu: %v\n", err)
		return
	}
	os.MkdirAll("test_blocks", 0755) // direktorijum za blokove, njega mogu da obrisem kada zavrsim sa testom
	bm := NoviBlockManager(bp, "test_blocks")
	if bm == nil {
		fmt.Println("Greska pri kreiranju BlockManager-a")
		return
	}

	// upis blokova na disk
	err = bm.UpisiSveBlokove()
	if err != nil {
		fmt.Printf("Greska pri upisu blokova na disk: %v\n", err)
		return
	}
	fmt.Printf("Uspesno upisano %d blokova na disk\n", len(bm.Blokovi))

	// citanje blokova sa diska
	for i := 0; i < len(bm.Blokovi); i++ {
		ucitaniBlok, err := bm.UcitajBlok(i)
		if err != nil {
			fmt.Printf("Greska pri citanju bloka %d: %v\n", i, err)
			continue
		}
		fmt.Printf("Blok %d: %d tuple-ova\n", i, len(ucitaniBlok.Podaci))
		// prikaz sadrzaja bloka
		for j, tuple := range ucitaniBlok.Podaci {
			fmt.Printf("      [%d] %s = %s (%s)\n", j, tuple.Kljuc, tuple.Vrednost, tuple.Dogadjaj)
		}
	}

	// pretraga kljuceva
	trazeniKljuc := "zika"
	indeksBloka, indeksTuple := bm.PronadjiBlok(trazeniKljuc)
	if indeksBloka != -1 {
		fmt.Printf("Kljuc '%s' pronadjen u bloku %d na poziciji %d\n", trazeniKljuc, indeksBloka, indeksTuple)
	} else {
		fmt.Printf("Kljuc '%s' nije pronadjen\n", trazeniKljuc)
	}
	trazeniKljuc = "nepostojeci"
	indeksBloka, indeksTuple = bm.PronadjiBlok(trazeniKljuc)
	if indeksBloka != -1 {
		fmt.Printf("Kljuc '%s' pronadjen u bloku %d na poziciji %d\n", trazeniKljuc, indeksBloka, indeksTuple)
	} else {
		fmt.Printf("Kljuc '%s' nije pronadjen\n", trazeniKljuc)
	}

	// dodavanje novih podataka
	noviPodaci := []bufferpool.Tuple{
		*bufferpool.NoviTuple("put", "novi1", "Vrednost1"),
		*bufferpool.NoviTuple("put", "novi2", "Vrednost2"),
	}
	err = bp.Unesi(noviPodaci)
	if err != nil {
		fmt.Printf("Greska pri upisu novih podataka: %v\n", err)
	} else {
		fmt.Printf("BufferPool sada ima %d tuple-ova\n", len(bp.Podaci))
		bm.OsveziBlokove()
		err = bm.UpisiSveBlokove()
		if err != nil {
			fmt.Printf("Greska pri upisu osvezenih blokova: %v\n", err)
		} else {
			fmt.Println("Osvezeni blokovi upisani na disk")
		}
	}

}
