package main

import (
	"slices"
)

type CountMinSketch struct {
	m            uint
	k            uint
	hashFunkcije []HashWithSeed
	tabela       [][]uint32
}

func (cms *CountMinSketch) izracunajK(delta float64) {
	cms.k = CalculateK(delta)
}

func (cms CountMinSketch) getK() uint {
	return cms.k
}

func (cms *CountMinSketch) setK(k uint) {
	cms.k = k
}

func (cms *CountMinSketch) izracunajM(epsilon float64) {
	cms.m = CalculateM(epsilon)
}

func (cms CountMinSketch) getM() uint {
	return cms.m
}

func (cms *CountMinSketch) setM(m uint) {
	cms.m = m
}

func (cms *CountMinSketch) napraviHashFunkcije() {
	cms.hashFunkcije = CreateHashFunctions(cms.k)
}

func (cms CountMinSketch) getHashFunkcije() []HashWithSeed {
	return cms.hashFunkcije
}

func (cms *CountMinSketch) setHashFunkcije(hashFunkcije []HashWithSeed) {
	cms.hashFunkcije = hashFunkcije
}

func (cms *CountMinSketch) napraviTabelu() [][]uint32 {
	cms.tabela = make([][]uint32, cms.k)
	for i := range cms.tabela {
		cms.tabela[i] = make([]uint32, cms.m)
	}

	return cms.tabela
}

func (cms CountMinSketch) getTabela() [][]uint32 {
	return cms.tabela
}

func (cms *CountMinSketch) setTabela(tabela [][]uint32) {
	cms.tabela = tabela
}

func (cms *CountMinSketch) unesi(data []byte) {
	for i := uint(0); i < cms.k; i++ {
		index := cms.hashFunkcije[i].Hash(data) % uint64(cms.m)
		cms.tabela[i][index]++
	}
}

func (cms CountMinSketch) pretrazi(data []byte) uint32 {
	mins := make([]uint32, cms.k)
	for i := uint(0); i < cms.k; i++ {
		index := cms.hashFunkcije[i].Hash(data) % uint64(cms.m)
		mins[i] = cms.tabela[i][index]
	}

	return slices.Min(mins)
}
