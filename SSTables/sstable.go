package SSTables

import (
	configuration "NASP_projekat/Configuration"
	"NASP_projekat/memtables"
	bf "NASP_projekat/probabilistics/BloomFilter"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type SSTableConfig struct {
	SummaryStep uint64 // svaki N-ti unos iz index-a ide u summary
	BlockSize   uint64 // velicina bloka za upis/ispis
}

type SSTable struct {
	putanja string
	config  SSTableConfig
}

// ovo su putanje do fajlova da bude organizovanije
func (s *SSTable) dataPutanja() string {
	return filepath.Join(s.putanja, "data.db")
}
func (s *SSTable) indexPutanja() string {
	return filepath.Join(s.putanja, "index.db")
}
func (s *SSTable) summaryPutanja() string {
	return filepath.Join(s.putanja, "summary.db")
}
func (s *SSTable) filterPutanja() string {
	return filepath.Join(s.putanja, "filter.db")
}
func (s *SSTable) metadataPutanja() string {
	return filepath.Join(s.putanja, "metadata.db")
}

// ucitava iz konfig fajla podesavanja
func UcitajSSTableConfig() SSTableConfig {
	konf := configuration.UcitajKonfiguraciju()
	sstConf, postoji := konf["SSTable"]
	if !postoji {
		return defaultConfig()
	}
	summaryStep := sstConf["SummaryStep"]
	blockSize := sstConf["BlockSize"]
	if summaryStep == 0 {
		summaryStep = 3
	}
	if blockSize == 0 {
		blockSize = 4096
	}
	return SSTableConfig{SummaryStep: summaryStep, BlockSize: blockSize}
}

// podrazumevane vrednosti
func defaultConfig() SSTableConfig {
	return SSTableConfig{SummaryStep: 3, BlockSize: 4096}
}

func NovaSSTable(putanja string, config SSTableConfig) *SSTable {
	os.MkdirAll(putanja, 0755)
	return &SSTable{putanja: putanja, config: config}
}


//ovo mi je bila funkcija samo za test da vidim da li radi bloom filter posto mi je filterPutanja bila privatna i ne vidi se u drugim paketima
//func (s *SSTable) FilterPutanja() string {
//	return s.filterPutanja()
//}

// upisuje podatke iz memtable u sstable fajlove
func (s *SSTable) Flush(zapisi []memtables.Unos) error {
	// konvertujem memtable zapise u sstable zapise
	sstZapisi := make([]*SSTableRecord, 0, len(zapisi))
	for _, unos := range zapisi {
		zapis := NoviZapis(unos.Kljuc, unos.Vrednost, unos.Timestamp, unos.Tombstone)
		sstZapisi = append(sstZapisi, zapis)
	}

	// bloomfilter
	seed := uint32(time.Now().Unix())
	bloomFilter := bf.NewBloomFilter(uint32(len(sstZapisi)), 0.01, seed)
	for _, zapis := range sstZapisi {
		bloomFilter.Dodaj([]byte(zapis.Key))
	}

	// data
	data := NoviDataSegment(s.dataPutanja(), s.config.BlockSize)
	offseti, err := data.Upisi(sstZapisi)
	if err != nil {
		return fmt.Errorf("greska pri upisu Data: %w", err)
	}

	// index
	index := NoviIndexSegment(s.indexPutanja(), s.config.BlockSize)
	indexUnosi, err := index.Upisi(offseti)
	if err != nil {
		return fmt.Errorf("greska pri upisu Index: %w", err)
	}

	// summary
	summary := NoviSummarySegment(s.summaryPutanja(), s.config.BlockSize, s.config.SummaryStep)
	if err := summary.Upisi(indexUnosi); err != nil {
		return fmt.Errorf("greska pri upisu Summary: %w", err)
	}

	// filter
	if err := bloomFilter.SerijalizacijaUFajl(s.filterPutanja()); err != nil {
		return fmt.Errorf("greska pri upisu Filter: %w", err)
	}

	// metadata
	metadata := NoviMetadataSegment(s.metadataPutanja(), s.config.BlockSize)
	if err := metadata.Upisi(sstZapisi); err != nil {
		return fmt.Errorf("greska pri upisu Metadata: %w", err)
	}

	return nil
}

// trazi kljuc u ovoj sstable tabeli
func (s *SSTable) Get(kljuc string) (*SSTableRecord, bool, error) {
	//  bloom filter brzo proverava da li kljuc mozda postoji
	bloomFilter, err := bf.DeserijalizacijaIzFajla(s.filterPutanja())
	if err != nil {
		return nil, false, fmt.Errorf("greska pri citanju Filter: %w", err)
	}
	if !bloomFilter.Proveri([]byte(kljuc)) {
		return nil, false, nil
	}

	// summary -provera opsega i gde da trazim u Index-u
	summary := NoviSummarySegment(s.summaryPutanja(), s.config.BlockSize, s.config.SummaryStep)
	startOffset, endOffset, pronadjen, err := summary.NadjiOpsegUIndexu(kljuc)
	if err != nil {
		return nil, false, fmt.Errorf("greska pri citanju Summary: %w", err)
	}
	if !pronadjen {
		return nil, false, nil
	}

	//  index- trazenje offseta u Data fajlu
	index := NoviIndexSegment(s.indexPutanja(), s.config.BlockSize)
	dataOffset, pronadjen, err := index.NadjiUIndexu(kljuc, startOffset, endOffset)
	if err != nil {
		return nil, false, fmt.Errorf("greska pri citanju Index: %w", err)
	}
	if !pronadjen {
		return nil, false, nil
	}

	//data - citanje zapisa sa tog offseta
	data := NoviDataSegment(s.dataPutanja(), s.config.BlockSize)
	zapis, err := data.CitajNaOffsetu(dataOffset)
	if err != nil {
		return nil, false, fmt.Errorf("greska pri citanju data: %w", err)
	}

	return zapis, true, nil
}

// Merkle validacija

// proverava da li su podaci osteceni
func (s *SSTable) ValidirajMerkle() ([]string, error) {
	data := NoviDataSegment(s.dataPutanja(), s.config.BlockSize)
	zapisi, err := data.CitajSve()
	if err != nil {
		return nil, fmt.Errorf("greska pri citanju podataka: %w", err)
	}

	metadata := NoviMetadataSegment(s.metadataPutanja(), s.config.BlockSize)
	razlike, err := metadata.Validiraj(zapisi)
	if err != nil {
		return nil, err
	}

	return razlike, nil
}

// pomocne funkcije
// najmanji kljuc u ovoj tabeli
func (s *SSTable) MinKljuc() (string, error) {
	fajl, err := os.Open(s.summaryPutanja())
	if err != nil {
		return "", err
	}
	defer fajl.Close()

	granice, err := deserijalizujGranice(fajl)
	if err != nil {
		return "", err
	}
	return granice.MinKljuc, nil
}

// najveci kljuc u ovoj tabeli
func (s *SSTable) MaxKljuc() (string, error) {
	fajl, err := os.Open(s.summaryPutanja())
	if err != nil {
		return "", err
	}
	defer fajl.Close()

	granice, err := deserijalizujGranice(fajl)
	if err != nil {
		return "", err
	}
	return granice.MaxKljuc, nil
}

func (s *SSTable) Putanja() string {
	return s.putanja
}
