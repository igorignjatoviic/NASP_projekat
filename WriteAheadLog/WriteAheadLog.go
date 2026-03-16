package wal

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

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

	fmt.Println("Upisano na lokaciji: ", segment)
	wal.blok = []byte{}

	return nil
}

func (wal *WriteAheadLog) ucitajSegment() (string, string, error) {
	lokacija := ""

	fajl, err := os.Open("WriteAheadLog/resources/segmenti.txt")
	if err != nil {
		return "", "", err
	}
	defer fajl.Close()

	skener := bufio.NewScanner(fajl)
	skener.Scan()

	for skener.Scan() {
		linija := skener.Text()
		podaci := strings.Split(linija, ",")

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
			lokacija = "WriteAheadLog/resources/" + ime + ".bin"
			return lokacija, ime, nil
		}
	}

	lokacija, ime := wal.kreirajSegment()
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
		return false
	} else if popunjen == 1 {
		return false
	}

	return true
}

func (wal *WriteAheadLog) kreirajSegment() (string, string) {
	fajl, err := os.ReadFile("WriteAheadLog/resources/segmenti.txt")
	if err != nil {
		errorFajl()
	}

	linije := strings.Split(string(fajl), "\n")
	noviWAL := "wal" + strconv.Itoa(len(linije)-1) + ",0,0,0"
	linije[len(linije)-1] = noviWAL

	ime := strings.Split(noviWAL, ",")[0]

	fajlUpis, err := os.Create("WriteAheadLog/resources/segmenti.txt")
	if err != nil {
		errorFajl()
	}
	defer fajlUpis.Close()

	for _, linija := range linije {
		fajlUpis.WriteString(linija + "\n")
	}

	return "WriteAheadLog/resources/wal" + strconv.Itoa(len(linije)-1) + ".bin", ime
}

func (wal *WriteAheadLog) izmeniSegment(ime string) (bool, error) {
	fajl, err := os.ReadFile("WriteAheadLog/resources/segmenti.txt")
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

	noviFajl, err := os.Create("WriteAheadLog/resources/segmenti.txt")
	if err != nil {
		return false, err
	}
	defer noviFajl.Close()

	for _, linija := range noveLinije {
		noviFajl.WriteString(linija + "\n")
	}

	return true, nil
}

func (wal *WriteAheadLog) ucitajWriteAheadLog() {
	fajlovi, err := os.ReadDir("WriteAheadLog/resources")
	if err != nil {
		errorFajl()
	}

	segmenti := make([]string, 0)
	for _, fajl := range fajlovi {
		if strings.Contains(fajl.Name(), "wal") {
			segmenti = append(segmenti, fajl.Name())
		}
	}

	for _, segment := range segmenti {
		fajl, err := os.Open("WriteAheadLog/resources/" + segment)
		if err != nil {
			errorFajl()
		}
		defer fajl.Close()

		fmt.Println("\nSegment: ", segment)
		for {
			wal.blok = []byte{}

			var crc uint32
			var timestamp uint64
			var tombstone uint8
			var keySize uint64
			var valueSize uint64

			err := binary.Read(fajl, binary.BigEndian, &crc)
			if err == io.EOF {
				break
			}
			if err != nil {
				errorFajl()
			}

			binary.Read(fajl, binary.BigEndian, &timestamp)
			binary.Read(fajl, binary.BigEndian, &tombstone)
			binary.Read(fajl, binary.BigEndian, &keySize)
			binary.Read(fajl, binary.BigEndian, &valueSize)

			header := make([]byte, 8+1+8+8)
			binary.BigEndian.PutUint64(header[0:8], timestamp)
			header[8] = tombstone
			binary.BigEndian.PutUint64(header[9:17], keySize)
			binary.BigEndian.PutUint64(header[17:25], valueSize)

			key := make([]byte, keySize)
			value := make([]byte, valueSize)

			binary.Read(fajl, binary.BigEndian, &key)
			binary.Read(fajl, binary.BigEndian, &value)

			wal.blok = append(wal.blok, header...)
			wal.blok = append(wal.blok, key...)
			wal.blok = append(wal.blok, value...)

			noviCRC := CRC32(wal.blok)

			dogadjaj := ""
			if tombstone == 0 {
				dogadjaj = "PUT"
			} else {
				dogadjaj = "DELETE"
			}

			fmt.Printf("\t%s(%s, %s)", dogadjaj, string(key), string(value))
			fmt.Println("\tUpisani CRC: ", crc, ", Novi CRC: ", noviCRC)
		}
	}
}

func errorParsiranje() {
	log.Fatal("Greska prilikom parsiranja.")
}

func errorFajl() {
	log.Fatal("Greska pri otvaranju fajla.")
}
