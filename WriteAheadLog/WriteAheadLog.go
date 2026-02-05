package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// napraviti da sistem sam bira segment iz segmenti.txt i upis writeaheadloga
// proveriti da li je upis dobar pomocu citanja i zatim napraviti bolji odabir walova, upis u fajl da je zapis promenjen itd
type WriteAheadLog struct {
	blok                 []byte
	segmenti             []string
	maksimalanBrojZapisa int
}

func (wal *WriteAheadLog) unesi(dogadjaj, kljuc, vrednost string) {
	wal.maksimalanBrojZapisa = 10

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
	segment := wal.ucitajSegmente()
	fajl, err := os.Create(segment)
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

func (wal *WriteAheadLog) ucitajSegmente() string {
	fajl, err := os.Open("resources/segmenti.txt")
	if err != nil {
		fmt.Println("Greska pri otvaranju fajla.")
	}
	defer fajl.Close()

	skener := bufio.NewScanner(fajl)
	skener.Scan()

	for skener.Scan() {
		line := skener.Text()
		podaci := strings.Split(line, ",")
		brojZapisa, err := strconv.Atoi(podaci[1])
		if err != nil {
			fmt.Println("Greska u parsiranju.")
		}

		if brojZapisa < wal.maksimalanBrojZapisa {
			return "resources/" + podaci[0] + ".bin"
		}
	}

	return ""
}
