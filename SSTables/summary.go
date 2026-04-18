package SSTables

import (
	"bytes"
	"encoding/binary"
	"io"
	"os"
)

// format summary fajla:
// prvo idu granice- min i max kljuc
// zatim svaki N-ti unos iz index fajla

type SummarySegment struct {
	putanja     string
	blockSize   uint64
	summaryStep uint64 // svaki N-ti unos iz index
}

type SummaryUnos struct {
	Kljuc         string
	OffsetUIndexu uint64
}

type SummaryGranice struct {
	MinKljuc string
	MaxKljuc string
}

func NoviSummarySegment(putanja string, blockSize uint64, summaryStep uint64) *SummarySegment {
	if summaryStep == 0 {
		summaryStep = 1
	}
	return &SummarySegment{putanja: putanja, blockSize: blockSize, summaryStep: summaryStep}
}

// upisuje summary fajl
func (s *SummarySegment) Upisi(indexUnosi []IndexUnos) error {
	if len(indexUnosi) == 0 {
		return nil
	}

	fajl, err := os.Create(s.putanja)
	if err != nil {
		return err
	}
	defer fajl.Close()

	blok := make([]byte, 0, s.blockSize)

	// prvo upisujem granice - min i max kljuc
	minKljuc := indexUnosi[0].Kljuc
	maxKljuc := indexUnosi[len(indexUnosi)-1].Kljuc
	var headerBuf bytes.Buffer
	if err := serijalizujGranice(&headerBuf, minKljuc, maxKljuc); err != nil {
		return err
	}
	blok = append(blok, headerBuf.Bytes()...)

	// upisujem svaki summarystep-ti unos iz index
	for i, unos := range indexUnosi {
		if i%int(s.summaryStep) != 0 {
			continue
		}
		var buf bytes.Buffer
		if err := serijalizujSummaryUnos(&buf, unos.Kljuc, unos.Offset); err != nil {
			return err
		}
		blok = append(blok, buf.Bytes()...)

		if uint64(len(blok)) >= s.blockSize {
			if _, err := fajl.Write(blok); err != nil {
				return err
			}
			blok = blok[:0]
		}
	}

	if len(blok) > 0 {
		if _, err := fajl.Write(blok); err != nil {
			return err
		}
	}

	return nil
}

// trazi u kom delu index fajla se nalazi kljuc
func (s *SummarySegment) NadjiOpsegUIndexu(kljuc string) (uint64, uint64, bool, error) {
	fajl, err := os.Open(s.putanja)
	if err != nil {
		return 0, 0, false, err
	}
	defer fajl.Close()

	fi, err := fajl.Stat()
	if err != nil {
		return 0, 0, false, err
	}

	return pretragaUSummary(fajl, 0, uint64(fi.Size()), kljuc)
}

func pretragaUSummary(fajl *os.File, startOffset uint64, endOffset uint64, kljuc string) (uint64, uint64, bool, error) {
	if _, err := fajl.Seek(int64(startOffset), io.SeekStart); err != nil {
		return 0, 0, false, err
	}

	// prvo proverim granice
	granice, err := deserijalizujGranice(fajl)
	if err != nil {
		return 0, 0, false, err
	}

	if kljuc < granice.MinKljuc || kljuc > granice.MaxKljuc {
		return 0, 0, false, nil
	}

	// trazim izmedju kojih summary unosa je kljuc
	var prethodni *SummaryUnos
	var trenutni *SummaryUnos
	trenutnaPozicija, _ := fajl.Seek(0, io.SeekCurrent)

	for trenutnaPozicija < int64(endOffset) {
		unos, err := deserijalizujSummaryUnos(fajl)
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, 0, false, err
		}

		if unos.Kljuc > kljuc {
			trenutni = unos
			break
		}
		prethodni = unos
		trenutnaPozicija, _ = fajl.Seek(0, io.SeekCurrent)
	}

	if prethodni == nil {
		return 0, 0, false, nil
	}

	start := prethodni.OffsetUIndexu
	if trenutni != nil {
		return start, trenutni.OffsetUIndexu, true, nil
	}

	// ako je poslednji segment, citam do kraja fajla
	return start, ^uint64(0), true, nil
}

func serijalizujGranice(w io.Writer, minKljuc string, maxKljuc string) error {
	minSize := uint64(len(minKljuc))
	if err := binary.Write(w, binary.BigEndian, minSize); err != nil {
		return err
	}
	if _, err := w.Write([]byte(minKljuc)); err != nil {
		return err
	}
	maxSize := uint64(len(maxKljuc))
	if err := binary.Write(w, binary.BigEndian, maxSize); err != nil {
		return err
	}
	if _, err := w.Write([]byte(maxKljuc)); err != nil {
		return err
	}
	return nil
}

func deserijalizujGranice(r io.Reader) (*SummaryGranice, error) {
	var minSize uint64
	if err := binary.Read(r, binary.BigEndian, &minSize); err != nil {
		return nil, err
	}
	minBytes := make([]byte, minSize)
	if _, err := io.ReadFull(r, minBytes); err != nil {
		return nil, err
	}
	var maxSize uint64
	if err := binary.Read(r, binary.BigEndian, &maxSize); err != nil {
		return nil, err
	}
	maxBytes := make([]byte, maxSize)
	if _, err := io.ReadFull(r, maxBytes); err != nil {
		return nil, err
	}
	return &SummaryGranice{MinKljuc: string(minBytes), MaxKljuc: string(maxBytes)}, nil
}

func serijalizujSummaryUnos(w io.Writer, kljuc string, offset uint64) error {
	keySize := uint64(len(kljuc))
	if err := binary.Write(w, binary.BigEndian, keySize); err != nil {
		return err
	}
	if _, err := w.Write([]byte(kljuc)); err != nil {
		return err
	}
	return binary.Write(w, binary.BigEndian, offset)
}

func deserijalizujSummaryUnos(r io.Reader) (*SummaryUnos, error) {
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
	return &SummaryUnos{Kljuc: string(keyBytes), OffsetUIndexu: offset}, nil
}
