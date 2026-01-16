package main

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"math"
	"math/bits"
	"os"
)

// HELPER FUNKCIJE KOJE SMO DOBILI NA VEZBAMA

const (
	HLL_MIN_PRECISION = 4
	HLL_MAX_PRECISION = 16
)

// vraca prvih k bitova iz 64bitne vrednosti
func firstKbits(value, k uint64) uint64 {
	return value >> (64 - k)
}

// vraca broj nula na kraju broja
func trailingZeroBits(value uint64) int {
	return bits.TrailingZeros64(value)
}

type HLL struct {
	m   uint64
	p   uint8
	reg []uint8
}

// procena kardinalnosti
func (hll *HLL) Estimate() float64 {
	sum := 0.0

	for _, val := range hll.reg { // racuna sumu
		sum += math.Pow(math.Pow(2.0, float64(val)), -1)
	}

	// harmonijska sredina
	alpha := 0.7213 / (1.0 + 1.079/float64(hll.m))
	estimation := alpha * math.Pow(float64(hll.m), 2.0) / sum

	// korekcija za male i velike vrednosti
	emptyRegs := hll.emptyCount()
	if estimation <= 2.5*float64(hll.m) { // do small range correction
		if emptyRegs > 0 {
			estimation = float64(hll.m) * math.Log(float64(hll.m)/float64(emptyRegs))
		}
	} else if estimation > 1/30.0*math.Pow(2.0, 32.0) { // do large range correction
		estimation = -math.Pow(2.0, 32.0) * math.Log(1.0-estimation/math.Pow(2.0, 32.0))
	}
	return estimation
}

func (hll *HLL) emptyCount() int {
	sum := 0
	for _, val := range hll.reg {
		if val == 0 {
			sum++
		}
	}
	return sum
}

// KRAJ HELPER FUNKCIJA SA VEZBI

func napraviHyperLogLog(_p uint8) (*HLL, error) {
	if _p < HLL_MIN_PRECISION || _p > HLL_MAX_PRECISION {
		return nil, fmt.Errorf("Greska: vrednost p mora biti iz opsega [4,16]! ")
	}

	_m := uint64(1) << _p // siftujem bitove od 1 ulevo jer je m=2 na stepen p
	hll := HLL{
		m:   _m,
		p:   _p,
		reg: make([]uint8, _m),
	}
	return &hll, nil
}

func MojHash(vrednost []byte) uint64 { // moja hes funkcija
	// koristim fnv hesiranje; ne treba mi seed jer fnv uvek koristi istu inicijalnu vrednost
	h := fnv.New64()
	h.Write(vrednost)
	hash_vrednost := h.Sum64()
	return hash_vrednost
}

func (hll *HLL) Dodaj(niz []byte) {
	hash_vrednost := MojHash(niz)

	baket := firstKbits(hash_vrednost, uint64(hll.p))      // uint64
	vrednost := uint8(trailingZeroBits(hash_vrednost) + 1) // vrednost koju upisujem u baket; maks je 64+1=65 (ako su sve 0), a taj broj moze stati u uint8
	if vrednost > hll.reg[baket] {
		hll.reg[baket] = vrednost // azuriram vrednost u baketu ako je nova veca od vec upisane
	}
}

func (hll *HLL) DodajString(s string) {
	hll.Dodaj([]byte(s))
}

func (hll *HLL) Resetuj() {
	for i := range hll.reg {
		hll.reg[i] = 0
	}
}

func ucitajHyperLogLog(nazivFajla string) (*HLL, error) {
	fajl, err := os.Open(nazivFajla)
	if err != nil {
		return nil, fmt.Errorf("Greska pri ucitavanju HLL fajla: %v", err)
	}
	defer fajl.Close()
	var _p uint8
	err = binary.Read(fajl, binary.BigEndian, &_p)
	if err != nil {
		return nil, fmt.Errorf("Greska pri citanju preciznosti: %v", err)
	}
	if _p < HLL_MIN_PRECISION || _p > HLL_MAX_PRECISION {
		return nil, fmt.Errorf("Greska, nevalidna preciznost pri citanju: %d", _p)
	}
	_m := uint64(1) << _p // m=2 na stepen p
	_reg := make([]uint8, _m)
	err = binary.Read(fajl, binary.BigEndian, &_reg)
	if err != nil {
		return nil, fmt.Errorf("Greska pri citanju registara: %v", err)
	}
	hll := &HLL{
		m:   _m,
		p:   _p,
		reg: _reg,
	}
	return hll, nil
}

func (hll *HLL) Sacuvaj(nazivFajla string) error {
	fajl, err := os.Create(nazivFajla)
	if err != nil {
		return fmt.Errorf("Greska pri kreiranju fajla za cuvanje HLL: %v", err)
	}
	defer fajl.Close()
	err = binary.Write(fajl, binary.BigEndian, hll.p)
	if err != nil {
		return fmt.Errorf("Greska pri cuvanju preciznosti: %v", err)
	}
	err = binary.Write(fajl, binary.BigEndian, hll.reg)
	if err != nil {
		return fmt.Errorf("Greska pri cuvanju registara: %v", err)
	}
	return nil
}
