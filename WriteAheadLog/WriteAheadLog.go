package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// testirati sistem za biranje segmenata i modifikovanje segment.txt prilikom upisa zapisa
type WriteAheadLog struct {
	blok                 []byte
	segmenti             []string
	maksimalanBrojZapisa int
	padding              int
}

func (wal *WriteAheadLog) unesi(dogadjaj, kljuc, vrednost string) {
	wal.maksimalanBrojZapisa = 200
	wal.padding = 50

	wal.izracunajTimestamp()
	wal.izracunajTombstone(dogadjaj)
	wal.izracunajDuzinuParametra(kljuc)
	wal.izracunajDuzinuParametra(vrednost)
	wal.konvertujParametarUBajtove(kljuc)
	wal.konvertujParametarUBajtove(vrednost)
	wal.izracunajCRC()

	wal.sacuvajBlok()
}

func (wal *WriteAheadLog) izracunajCRC() {
	bafer := make([]byte, 4)
	binary.BigEndian.PutUint32(bafer, CRC32(wal.blok))
	wal.blok = append(bafer, wal.blok...)
}

func (wal *WriteAheadLog) izracunajTimestamp() {
	timestamp := time.Now().Unix()
	bafer := make([]byte, 8)
	binary.BigEndian.PutUint64(bafer, uint64(timestamp))
	wal.blok = append(wal.blok, bafer...)
}

func (wal *WriteAheadLog) izracunajTombstone(dogadjaj string) {
	if dogadjaj == "delete" {
		wal.blok = append(wal.blok, byte(1))
	} else {
		wal.blok = append(wal.blok, byte(0))
	}
}

func (wal *WriteAheadLog) izracunajDuzinuParametra(parametar string) {
	duzina := len(parametar)
	bafer := make([]byte, 8)
	binary.BigEndian.PutUint64(bafer, uint64(duzina))
	wal.blok = append(wal.blok, bafer...)
}

func (wal *WriteAheadLog) konvertujParametarUBajtove(parametar string) {
	bafer := []byte{}
	bafer = append(bafer, []byte(parametar)...)
	wal.blok = append(wal.blok, bafer...)
}

func (wal *WriteAheadLog) sacuvajBlok() error {
	segment, err := wal.ucitajSegment()
	if err != nil {
		log.Fatal("Greska.")
	}
	fajl, err := os.Open(segment)
	if err != nil {
		return err
	}
	defer fajl.Close()

	if err := binary.Write(fajl, binary.BigEndian, wal.blok); err != nil {
		return err
	}
	fmt.Println("upisano")

	wal.blok = []byte{}
	return nil
}

func (wal *WriteAheadLog) ucitajSegment() (string, error) {
	lokacija := ""
	fajl, err := os.Open("resources/segmenti.txt")
	if err != nil {
		return lokacija, err
	}
	defer fajl.Close()

	skener := bufio.NewScanner(fajl)
	skener.Scan()

	for skener.Scan() {
		linija := skener.Text()
		podaci := strings.Split(linija, ",")

		popunjen, err := strconv.Atoi(podaci[2])
		if err != nil {
			return lokacija, err
		}

		popunjenaMemorija, err := strconv.Atoi(podaci[1])
		if err != nil {
			return lokacija, err
		}
		if wal.maksimalanBrojZapisa-popunjenaMemorija < wal.padding {
			fmt.Println("Preostalo je premalo memorije u vasem walu.")
		} else if popunjen == 0 {
			lokacija = "resources/" + podaci[0] + ".bin"
		}
	}

	if lokacija == "" {
		wal.kreirajSegment()
	}

	return lokacija, nil
}

func (wal *WriteAheadLog) kreirajSegment() {
	fajl, err := os.ReadFile("resources/segmenti.txt")
	if err != nil {
		errorFajl()
	}
	linije := strings.Split(string(fajl), "\n")
	noviWAL := "wal" + strconv.Itoa(len(linije)-1) + ",0,0,0"
	linije = append(linije, noviWAL)

	noviFajl, err := os.Create("resources/segmenti.txt")
	if err != nil {
		errorFajl()
	}
	defer noviFajl.Close()

	for i := range len(linije) {
		noviFajl.WriteString(linije[i])
	}
	noviFajl.WriteString("\n")
	fmt.Println("gotovo")
}

func errorParsiranje() {
	log.Fatal("Greska prilikom parsiranja.")
}

func errorFajl() {
	log.Fatal("Greska pri otvaranju fajla.")
}
