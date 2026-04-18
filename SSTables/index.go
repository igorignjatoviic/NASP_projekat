package SSTables

import (
	"bytes"
	"encoding/binary"
	"io"
	"os"
)

type IndexSegment struct {
	putanja   string
	blockSize uint64
}

type IndexUnos struct {
	Kljuc  string
	Offset uint64
}

func NoviIndexSegment(putanja string, blockSize uint64) *IndexSegment {
	return &IndexSegment{putanja: putanja, blockSize: blockSize}
}

// upisuje sve indekse i vraca njihove offsete-treba za summary
func (idx *IndexSegment) Upisi(offseti []OffsetInfo) ([]IndexUnos, error) {
	fajl, err := os.Create(idx.putanja)
	if err != nil {
		return nil, err
	}
	defer fajl.Close()

	indexUnosi := make([]IndexUnos, 0, len(offseti))
	blok := make([]byte, 0, idx.blockSize)
	trenutniOffset := uint64(0)

	for _, info := range offseti {
		var buf bytes.Buffer
		if err := serijalizujIndexUnos(&buf, info.Kljuc, info.Offset); err != nil {
			return nil, err
		}
		unosBytes := buf.Bytes()

		indexUnosi = append(indexUnosi, IndexUnos{Kljuc: info.Kljuc, Offset: trenutniOffset})
		trenutniOffset += uint64(len(unosBytes))
		blok = append(blok, unosBytes...)

		if uint64(len(blok)) >= idx.blockSize {
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

	return indexUnosi, nil
}

// trazi kljuc u index fajlu izmedju dva offseta (koje dobije iz summary)
func (idx *IndexSegment) NadjiUIndexu(kljuc string, startOffset uint64, endOffset uint64) (uint64, bool, error) {
	fajl, err := os.Open(idx.putanja)
	if err != nil {
		return 0, false, err
	}
	defer fajl.Close()

	fi, err := fajl.Stat()
	if err != nil {
		return 0, false, err
	}

	if endOffset == ^uint64(0) {
		endOffset = uint64(fi.Size())
	}

	return pretragaUIndexu(fajl, kljuc, startOffset, endOffset)
}

func pretragaUIndexu(fajl *os.File, kljuc string, startOffset uint64, endOffset uint64) (uint64, bool, error) {
	if startOffset >= endOffset {
		return 0, false, nil
	}

	if _, err := fajl.Seek(int64(startOffset), io.SeekStart); err != nil {
		return 0, false, err
	}

	trenutnaPozicija := startOffset
	for trenutnaPozicija < endOffset {
		unos, err := deserijalizujIndexUnos(fajl)
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, false, err
		}

		if unos.Kljuc == kljuc {
			return unos.Offset, true, nil
		}
		trenutnaPozicija += velicinaIndexUnosa(unos.Kljuc)
	}

	return 0, false, nil
}

func serijalizujIndexUnos(w io.Writer, kljuc string, offset uint64) error {
	keySize := uint64(len(kljuc))
	if err := binary.Write(w, binary.BigEndian, keySize); err != nil {
		return err
	}
	if _, err := w.Write([]byte(kljuc)); err != nil {
		return err
	}
	return binary.Write(w, binary.BigEndian, offset)
}

func deserijalizujIndexUnos(r io.Reader) (*IndexUnos, error) {
	var keySize uint64
	if err := binary.Read(r, binary.BigEndian, &keySize); err != nil {
		return nil, err
	}
	keyBytes := make([]byte, keySize)
	if _, err := io.ReadFull(r, keyBytes); err != nil {
		return nil, err
	}
	var offset uint64
	if err := binary.Read(r, binary.BigEndian, &offset); err != nil {
		return nil, err
	}
	return &IndexUnos{Kljuc: string(keyBytes), Offset: offset}, nil
}

// velicina jednog index unosa u bajtovima keySize+kljuc+offset
func velicinaIndexUnosa(kljuc string) uint64 {
	return 8 + uint64(len(kljuc)) + 8
}
