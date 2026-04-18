package SSTables

import (
	"encoding/binary"
	"hash/crc32"
	"io"
)

// Format zapisa: CRC (4B) | Timestamp (8B) | Tombstone (1B) | KeySize (8B) | ValueSize (8B) | Key | Value

type SSTableRecord struct {
	CRC       uint32
	Timestamp int64
	Tombstone bool
	KeySize   uint64
	ValueSize uint64
	Key       string
	Value     []byte
}

// pravi novi zapis i racuna mu CRC
func NoviZapis(kljuc string, vrednost []byte, timestamp int64, tombstone bool) *SSTableRecord {
	r := &SSTableRecord{
		Timestamp: timestamp,
		Tombstone: tombstone,
		KeySize:   uint64(len(kljuc)),
		ValueSize: uint64(len(vrednost)),
		Key:       kljuc,
		Value:     vrednost,
	}
	r.CRC = izracunajCRC(r)
	return r
}

func izracunajCRC(r *SSTableRecord) uint32 {
	buf := make([]byte, 0)
	tsBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(tsBytes, uint64(r.Timestamp))
	buf = append(buf, tsBytes...)
	if r.Tombstone {
		buf = append(buf, 1)
	} else {
		buf = append(buf, 0)
	}
	buf = append(buf, []byte(r.Key)...)
	buf = append(buf, r.Value...)
	return crc32.ChecksumIEEE(buf)
}

// upisuje zapis u fajl
func (r *SSTableRecord) Serijalizuj(w io.Writer) error {
	if err := binary.Write(w, binary.BigEndian, r.CRC); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, r.Timestamp); err != nil {
		return err
	}
	tombstoneByte := byte(0)
	if r.Tombstone {
		tombstoneByte = 1
	}
	if err := binary.Write(w, binary.BigEndian, tombstoneByte); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, r.KeySize); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, r.ValueSize); err != nil {
		return err
	}
	if _, err := w.Write([]byte(r.Key)); err != nil {
		return err
	}
	if _, err := w.Write(r.Value); err != nil {
		return err
	}
	return nil
}

// cita zapis iz fajla
func DeserijalizujZapis(r io.Reader) (*SSTableRecord, error) {
	rec := &SSTableRecord{}
	if err := binary.Read(r, binary.BigEndian, &rec.CRC); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.BigEndian, &rec.Timestamp); err != nil {
		return nil, err
	}
	var tombstoneByte byte
	if err := binary.Read(r, binary.BigEndian, &tombstoneByte); err != nil {
		return nil, err
	}
	rec.Tombstone = tombstoneByte == 1
	if err := binary.Read(r, binary.BigEndian, &rec.KeySize); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.BigEndian, &rec.ValueSize); err != nil {
		return nil, err
	}
	keyBytes := make([]byte, rec.KeySize)
	if _, err := io.ReadFull(r, keyBytes); err != nil {
		return nil, err
	}
	rec.Key = string(keyBytes)
	rec.Value = make([]byte, rec.ValueSize)
	if _, err := io.ReadFull(r, rec.Value); err != nil {
		return nil, err
	}
	return rec, nil
}

// proverava da li je CRC ispravan
func (r *SSTableRecord) ValidanCRC() bool {
	return r.CRC == izracunajCRC(r)
}

// vraca velicinu zapisa u bajtovima
func (r *SSTableRecord) Velicina() uint64 {
	return 4 + 8 + 1 + 8 + 8 + r.KeySize + r.ValueSize
}
