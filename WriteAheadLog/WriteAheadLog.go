package main

import (
	"encoding/binary"
	"fmt"
	"time"
)

type WriteAheadLog struct {
	blok   []byte
	logovi []string
}

func (wal *WriteAheadLog) unesi(dogadjaj, kljuc, vrednost string) {
	timestamp := time.Now().Unix()
	buffer := make([]byte, 8)
	binary.BigEndian.PutUint64(buffer, uint64(timestamp))

	buffer2 := make([]byte, 1)
	buffer2[0] = 0
	if dogadjaj == "delete" {
		buffer2[0] = 1
	}

	duzinaKljuca := len(kljuc)
	buffer3 := make([]byte, 8)
	binary.BigEndian.PutUint64(buffer3, uint64(duzinaKljuca))

	wal.blok = append(wal.blok, buffer3...)

	fmt.Print(wal.blok)
}
