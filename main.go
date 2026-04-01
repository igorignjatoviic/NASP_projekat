package main

import (
	blockmanager "NASP_projekat/BlockManager"
	bufferpool "NASP_projekat/BufferPool"
	configuration "NASP_projekat/Configuration"
	wal "NASP_projekat/WriteAheadLog"
	probabilistics "NASP_projekat/probabilistics"
	cms "NASP_projekat/probabilistics/CountMinSketch"
	"fmt"
	"os"
)

func meni() {
	for {
		fmt.Println("===== No-SQL Engine =====")
		fmt.Println()
		fmt.Println("1 - Probabilisticke strukture")
		fmt.Println("2 - Konfiguracija")
		fmt.Println("3 - Put") // testiranje
		fmt.Println("4 - Test Bufferpool")
		fmt.Println("5 - Test Block Manager")
		fmt.Println("0 - Izlaz")

		fmt.Print("Unesite jednu od ponudjenih opcija: ")
		opcija := cms.UnesiBroj()

		switch opcija {
		case 1:
			probabilistics.ProbabilistickeStruktureMeni()
		case 2:
			configuration.KonfiguracijaMeni()
		case 3:
			wal.Unesi("delete", "operation", "doomsday")
			wal.Ispisi()
		case 4:
			vrednost := bufferpool.Get("operation")
			fmt.Println("Vrednost: ", vrednost)
		case 5:
			testBlockManager()
		case 0:
			ocistiProzor()
			fmt.Println("Izasli ste iz aplikacije.")
			os.Exit(0)
		default:
			fmt.Println("Pogresan unos, pokusajte ponovo.")
		}
	}
}

func ocistiProzor() {
	fmt.Print("\033[H\033[2J")
}

func main() {
	meni()
}

func testBlockManager() {
	ocistiProzor()
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
	bm := blockmanager.NoviBlockManager(bp, "test_blocks")
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
