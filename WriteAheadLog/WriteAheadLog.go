package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

// implementirati citanje segmenata, ulepsati i testirati sve do sada
type WriteAheadLog struct {
	blok               []byte
	segmenti           []string
	maksimalnaMemorija int
	padding            int
}

func (wal *WriteAheadLog) unesi(dogadjaj, kljuc, vrednost string) {
	wal.maksimalnaMemorija = 200
	wal.padding = 50

	wal.izracunajTimestamp()
	wal.izracunajTombstone(dogadjaj)
	wal.izracunajDuzinuParametra(kljuc)
	wal.izracunajDuzinuParametra(vrednost)
	wal.konvertujParametarUBajtove(kljuc)
	wal.konvertujParametarUBajtove(vrednost)
	wal.izracunajCRC()

	wal.sacuvajBlok()
	fmt.Println("upisano")
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
	segment, ime, err := wal.ucitajSegment()
	if err != nil {
		log.Fatal("Greska.")
	}

	fajl, err := os.OpenFile(segment, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer fajl.Close()

	if err := binary.Write(fajl, binary.BigEndian, wal.blok); err != nil {
		return err
	}

	_, err = wal.izmeniSegment(ime)
	if err != nil {
		errorFajl()
	}

	wal.blok = []byte{}
	return nil
}

func (wal *WriteAheadLog) ucitajSegment() (string, string, error) {
	lokacija := ""

	fajl, err := os.Open("resources/segmenti.txt")
	if err != nil {
		return "", "", err
	}
	defer fajl.Close()

	skener := bufio.NewScanner(fajl)
	skener.Scan()

	for skener.Scan() {
		linija := skener.Text()
		podaci := strings.Split(linija, ",")
		fmt.Println(linija)

		ime := podaci[0]
		popunjenaMemorija, err := strconv.Atoi(podaci[1])
		if err != nil {
			return "", "", err
		}
		popunjen, err := strconv.Atoi(podaci[2])
		if err != nil {
			return "", "", err
		}

		if wal.proveriSegment(popunjenaMemorija, popunjen) {
			lokacija = "resources/" + ime + ".bin"
			return lokacija, ime, nil
		}
	}

	lokacija, ime := wal.kreirajSegment()
	fmt.Println(lokacija)
	noviFajl, err := os.Create(lokacija)
	if err != nil {
		return "", "", err
	}
	defer noviFajl.Close()

	return lokacija, ime, nil
}

func (wal *WriteAheadLog) proveriSegment(popunjenaMemorija int, popunjen int) bool {
	dostupnaMemorija := wal.maksimalnaMemorija - popunjenaMemorija

	if dostupnaMemorija <= wal.padding {
		fmt.Println("Preostalo je premalo memorije u vasem wal.")
		return false
	} else if popunjen == 1 {
		return false
	}

	return true
}

func (wal *WriteAheadLog) kreirajSegment() (string, string) {
	fajl, err := os.ReadFile("resources/segmenti.txt")
	if err != nil {
		errorFajl()
	}

	linije := strings.Split(string(fajl), "\n")
	noviWAL := "wal" + strconv.Itoa(len(linije)-1) + ",0,0,0"
	linije[len(linije)-1] = noviWAL

	ime := strings.Split(noviWAL, ",")[0]

	fajlUpis, err := os.Create("resources/segmenti.txt")
	if err != nil {
		errorFajl()
	}
	defer fajlUpis.Close()

	for _, linija := range linije {
		fajlUpis.WriteString(linija + "\n")
	}

	return "resources/wal" + strconv.Itoa(len(linije)-1) + ".bin", ime
}

func (wal *WriteAheadLog) izmeniSegment(ime string) (bool, error) {
	fajl, err := os.ReadFile("resources/segmenti.txt")
	if err != nil {
		return false, err
	}

	linije := strings.Split(string(fajl), "\n")
	noveLinije := make([]string, len(linije))

	for i, linija := range linije {
		if strings.Contains(linija, ime) {
			podaci := strings.Split(linija, ",")

			popunjenaMemorija, err := strconv.Atoi(podaci[1])
			if err != nil {
				return false, err
			}
			popunjenaMemorija += len(wal.blok)

			popunjen, err := strconv.Atoi(podaci[2])
			if err != nil {
				return false, err
			}

			if wal.maksimalnaMemorija-popunjenaMemorija <= wal.padding {
				popunjen = 1
			}

			linija = ime + "," + strconv.Itoa(popunjenaMemorija) + "," + strconv.Itoa(popunjen) + ",0"
		}

		noveLinije[i] = linija
	}
	noveLinije = slices.Delete(noveLinije, len(noveLinije)-1, len(noveLinije))

	noviFajl, err := os.Create("resources/segmenti.txt")
	if err != nil {
		return false, err
	}
	defer noviFajl.Close()

	for _, linija := range noveLinije {
		noviFajl.WriteString(linija + "\n")
	}

	return true, nil
}

func errorParsiranje() {
	log.Fatal("Greska prilikom parsiranja.")
}

func errorFajl() {
	log.Fatal("Greska pri otvaranju fajla.")
}
