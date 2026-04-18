package SSTables

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

type DataSegment struct {
	putanja   string
	blockSize uint64
}

type OffsetInfo struct {
	Kljuc  string
	Offset uint64
}

func NoviDataSegment(putanja string, blockSize uint64) *DataSegment {
	return &DataSegment{putanja: putanja, blockSize: blockSize}
}

// upisuje sve zapise u data fajl i vraca njihove offsete
func (d *DataSegment) Upisi(zapisi []*SSTableRecord) ([]OffsetInfo, error) {
	fajl, err := os.Create(d.putanja)
	if err != nil {
		return nil, err
	}
	defer fajl.Close()

	offseti := make([]OffsetInfo, 0, len(zapisi))
	blok := make([]byte, 0, d.blockSize)
	trenutniOffset := uint64(0)

	for _, zapis := range zapisi {
		var buf bytes.Buffer
		if err := zapis.Serijalizuj(&buf); err != nil {
			return nil, err
		}
		zapisBytes := buf.Bytes()

		offseti = append(offseti, OffsetInfo{Kljuc: zapis.Key, Offset: trenutniOffset})
		trenutniOffset += uint64(len(zapisBytes))
		blok = append(blok, zapisBytes...)

		if uint64(len(blok)) >= d.blockSize {
			if _, err := fajl.Write(blok); err != nil {
				return nil, err
			}
			blok = blok[:0]
		}
	}

	if len(blok) > 0 {
		if _, err := fajl.Write(blok); err != nil {
			return nil, err
		}
	}

	return offseti, nil
}

// cita jedan zapis sa zadate pozicije
func (d *DataSegment) CitajNaOffsetu(offset uint64) (*SSTableRecord, error) {
	fajl, err := os.Open(d.putanja)
	if err != nil {
		return nil, err
	}
	defer fajl.Close()

	if _, err := fajl.Seek(int64(offset), io.SeekStart); err != nil {
		return nil, err
	}

	zapis, err := DeserijalizujZapis(fajl)
	if err != nil {
		return nil, err
	}
	if !zapis.ValidanCRC() {
		return nil, fmt.Errorf("crc neispravan na %d", offset)
	}
	return zapis, nil
}

// koristim samo za Merkle validaciju
func (d *DataSegment) CitajSve() ([]*SSTableRecord, error) {
	fajl, err := os.Open(d.putanja)
	if err != nil {
		return nil, err
	}
	defer fajl.Close()

	var zapisi []*SSTableRecord
	for {
		zapis, err := DeserijalizujZapis(fajl)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if !zapis.ValidanCRC() {
			return nil, fmt.Errorf("crc neipravan za kljuc %s", zapis.Key)
		}
		zapisi = append(zapisi, zapis)
	}
	return zapisi, nil
}
