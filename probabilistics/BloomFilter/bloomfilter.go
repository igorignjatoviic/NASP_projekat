package main

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"os"
)

type HashSeed struct {
	Seed []byte
}

func (hs HashSeed) Hash(data []byte) uint64 {
	h := md5.New()
	h.Write(append(data, hs.Seed...)) //ove tri tackice je da rabije na bajtove seed posto je 4 bajtni a to ne mogu da smestim u slice
	return binary.BigEndian.Uint64(h.Sum(nil))
}

func generisiHashFunkcije(k uint32, seed uint32) []HashSeed {
	hashes := make([]HashSeed, k)
	for i := uint32(0); i < k; i++ {
		s := make([]byte, 4)
		binary.BigEndian.PutUint32(s, seed+i)
		hashes[i] = HashSeed{Seed: s}
	}
	return hashes
}

func Greska(e error, poruka string) {
	if e != nil {
		log.Fatalf("%s: %v", poruka, e)
	}
}

func izracunajM(numElements uint32, fpRate float64) uint {
	return uint(math.Ceil(float64(numElements) * math.Abs(math.Log(fpRate)) / math.Pow(math.Log(2), 2)))
}

func izracunajK(numElements uint32, m uint) uint {
	return uint(math.Ceil((float64(m) / float64(numElements)) * math.Log(2)))
}

type BloomFilter struct {
	BitArray     []byte
	HashFunkcije []HashSeed
}

func NewBloomFilter(numElements uint32, fpRate float64, seed uint32) *BloomFilter {
	m := izracunajM(numElements, fpRate)
	k := izracunajK(numElements, m)
	bajtDuzina := (m + 7) / 8
	return &BloomFilter{
		BitArray:     make([]byte, bajtDuzina),
		HashFunkcije: generisiHashFunkcije(uint32(k), seed),
	}
}

func (bf *BloomFilter) postaviBit(index uint64) {
	bajtIndex := index / 8
	bitIndex := index % 8
	bf.BitArray[bajtIndex] |= 1 << bitIndex
}

func (bf *BloomFilter) procitajBit(index uint64) bool {
	bajtIndex := index / 8
	bitIndex := index % 8
	return (bf.BitArray[bajtIndex] & (1 << bitIndex)) != 0
}

// Dodavanje elementa u Bloom Filter
func (bf *BloomFilter) Dodaj(data []byte) {
	for _, hf := range bf.HashFunkcije {
		index := hf.Hash(data) % (uint64(len(bf.BitArray)) * 8)
		bf.postaviBit(index)
	}
}

func (bf *BloomFilter) Proveri(data []byte) bool {
	for _, hf := range bf.HashFunkcije {
		index := hf.Hash(data) % (uint64(len(bf.BitArray)) * 8)
		if !bf.procitajBit(index) {
			return false
		}
	}
	return true
}

func (bf *BloomFilter) SerijalizacijaUFajl(path string) error {
	buffer := new(bytes.Buffer)
	bitDuzina := uint32(len(bf.BitArray) * 8)
	if er := binary.Write(buffer, binary.BigEndian, bitDuzina); er != nil {
		return er
	}
	if _, er := buffer.Write(bf.BitArray); er != nil {
		return er
	}
	if er := binary.Write(buffer, binary.BigEndian, uint32(len(bf.HashFunkcije))); er != nil {
		return er
	}
	for _, hf := range bf.HashFunkcije {
		if _, er := buffer.Write(hf.Seed); er != nil {
			return er
		}
	}
	return os.WriteFile(path, buffer.Bytes(), 0666)
}

func DeserijalizacijaIzFajla(path string) (*BloomFilter, error) {
	data, er := os.ReadFile(path)
	if er != nil {
		return nil, er
	}

	reader := bytes.NewReader(data)
	bf := &BloomFilter{}
	var bitDuzina uint32
	if er := binary.Read(reader, binary.BigEndian, &bitDuzina); er != nil {
		return nil, er
	}
	bajtDuzina := (bitDuzina + 7) / 8
	bf.BitArray = make([]byte, bajtDuzina)
	if _, er := reader.Read(bf.BitArray); er != nil {
		return nil, er
	}
	var brojhf uint32
	if er := binary.Read(reader, binary.BigEndian, &brojhf); er != nil {
		return nil, er
	}
	bf.HashFunkcije = make([]HashSeed, brojhf)
	for i := uint32(0); i < brojhf; i++ {
		seed := make([]byte, 4)
		if _, er := reader.Read(seed); er != nil {
			return nil, er
		}
		bf.HashFunkcije[i] = HashSeed{Seed: seed}
	}
	return bf, nil
}

// spaja dva bloom filtera
func (bf *BloomFilter) Sproji(other *BloomFilter) error {
	if len(bf.BitArray) != len(other.BitArray) {
		return fmt.Errorf("ne moze, razlicite su dimenzije")
	}
	if len(bf.HashFunkcije) != len(other.HashFunkcije) {
		return fmt.Errorf("ne moze; razliciti broj hash funkcija")
	}
	for i := range bf.HashFunkcije {
		if !bytes.Equal(bf.HashFunkcije[i].Seed, other.HashFunkcije[i].Seed) {
			return fmt.Errorf("ne moze; razlicite hash funkcije")
		}
	}
	for i := range bf.BitArray {
		bf.BitArray[i] |= other.BitArray[i] //samo radimo or operaciju da spojimo
	}
	return nil
}
