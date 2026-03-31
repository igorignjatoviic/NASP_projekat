package wal

import (
	bufferpool "NASP_projekat/BufferPool"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"
)

type WALZapis struct {
	Kljuc     string
	Vrednost  []byte
	Timestamp int64
	Tombstone bool
}

func Unesi(dogadjaj string, kljuc string, vrednost string) {
	wal := WriteAheadLog{}
	wal.unesi(dogadjaj, kljuc, vrednost)

	bp := bufferpool.NoviBufferPool()
	podaci := bufferpool.Ucitaj()
	podaci = bufferpool.OsveziBufferPool(podaci, *bufferpool.NoviTuple(dogadjaj, kljuc, vrednost))
	bp.Unesi(podaci)
}

func Ispisi() {
	wal := WriteAheadLog{}
	wal.ucitajWriteAheadLog()

	podaci := bufferpool.Ucitaj()
	fmt.Println("\nBufferPool:")
	for _, podatak := range podaci {
		fmt.Printf("\t%s(%s, %s)\n", strings.ToUpper(podatak.Dogadjaj), podatak.Kljuc, podatak.Vrednost)
	}
}

func UcitajSveZapise() ([]WALZapis, error) { //javna funkcija koja se koristi za recovery
	w := WriteAheadLog{}
	return w.ucitajSveZapise()
}

func (wal *WriteAheadLog) ucitajSveZapise() ([]WALZapis, error) { //metoda koja vraca listu svih procitanih WAL zapisa
	fajlovi, err := os.ReadDir("WriteAheadLog/resources")
	if err != nil {
		return nil, err
	}
	segmenti := make([]string, 0)
	for _, fajl := range fajlovi {
		if strings.Contains(fajl.Name(), "wal") {
			segmenti = append(segmenti, fajl.Name())
		}
	}
	var rezultat []WALZapis
	for _, segment := range segmenti {
		fajl, err := os.Open("WriteAheadLog/resources/" + segment)
		if err != nil {
			return nil, err
		}
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
				fajl.Close()
				return nil, err
			}
			if err := binary.Read(fajl, binary.BigEndian, &timestamp); err != nil {
				fajl.Close()
				return nil, err
			}
			if err := binary.Read(fajl, binary.BigEndian, &tombstone); err != nil {
				fajl.Close()
				return nil, err
			}
			if err := binary.Read(fajl, binary.BigEndian, &keySize); err != nil {
				fajl.Close()
				return nil, err
			}
			if err := binary.Read(fajl, binary.BigEndian, &valueSize); err != nil {
				fajl.Close()
				return nil, err
			}

			header := make([]byte, 8+1+8+8)
			binary.BigEndian.PutUint64(header[0:8], timestamp)
			header[8] = tombstone
			binary.BigEndian.PutUint64(header[9:17], keySize)
			binary.BigEndian.PutUint64(header[17:25], valueSize)

			key := make([]byte, keySize)
			value := make([]byte, valueSize)

			if _, err := io.ReadFull(fajl, key); err != nil {
				fajl.Close()
				return nil, err
			}
			if _, err := io.ReadFull(fajl, value); err != nil {
				fajl.Close()
				return nil, err
			}

			wal.blok = append(wal.blok, header...)
			wal.blok = append(wal.blok, key...)
			wal.blok = append(wal.blok, value...)

			noviCRC := CRC32(wal.blok)
			if noviCRC != crc {
				fajl.Close()
				return nil, fmt.Errorf("crc mismatch u segmentu %s", segment)
			}

			rezultat = append(rezultat, WALZapis{
				Kljuc:     string(key),
				Vrednost:  append([]byte(nil), value...),
				Tombstone: tombstone == 1,
				Timestamp: int64(timestamp),
			})
		}
		fajl.Close()
	}
	return rezultat, nil
}
