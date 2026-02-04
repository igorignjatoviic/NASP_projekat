package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"time"
)

// napraviti da sistem sam bira segment iz segmenti.txt i upis writeaheadloga
type WriteAheadLog struct {
	blok     []byte
	segmenti []string
}

func (wal *WriteAheadLog) unesi(dogadjaj, kljuc, vrednost string) {
	wal.izracunajTimestamp()
	wal.izracunajTombstone(dogadjaj)
	wal.izracunajDuzinuParametra(kljuc)
	wal.izracunajDuzinuParametra(vrednost)
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
	bafer[0] = byte(duzina)
	copy(bafer[1:len(parametar)+1], []byte(parametar))
	wal.blok = append(wal.blok, bafer...)
}

func (wal *WriteAheadLog) sacuvajBlok() {
	wal.ucitajSegmente()
}

func (wal *WriteAheadLog) ucitajSegmente() {
	podaci, err := os.ReadFile("resources/segmenti.txt")
	if err != nil {
		fmt.Printf("Fajl nepostoji.")
		return
	}
	fmt.Println(string(podaci))
}
