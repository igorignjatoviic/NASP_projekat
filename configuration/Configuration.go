package configuration

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

// dodati ostale strukture u konfiguraciju
func KonfiguracijaMeni() {
	konfiguracija := UcitajKonfiguraciju()

	fmt.Print("===== Konfiguracija =====\n\n")
	fmt.Println("1 - WriteAheadLog")
	fmt.Println("2 - BufferPool")
	fmt.Println("3 - Memtable")
	fmt.Println("4 - SSTable")
	fmt.Println("0 - Nazad")

	fmt.Print("\nUnesite jednu od ponudjenih opcija: ")
	opcija := unesiBroj()

	switch opcija {
	case 1:
		konfiguracijaStruktura(konfiguracija, "WriteAheadLog")
	case 2:
		konfiguracijaStruktura(konfiguracija, "BufferPool")
	case 3:
		konfiguracijaStruktura(konfiguracija, "Memtable")
	case 4:
		konfiguracijaStruktura(konfiguracija, "SSTable")
	case 0:
		return
	}
}

func konfiguracijaStruktura(korenskaKonfig map[string]map[string]uint64,
	struktura string) map[string]map[string]uint64 {
	ocistiProzor()
	fmt.Print("===== Konfiguracija", struktura, "=====\n\n")
	fmt.Println("1 - Trenutna konfiguracija")
	fmt.Println("2 - Izmeni konfiguraciju")

	fmt.Print("\nUnesite jednu od ponudjenih opcija: ")
	opcija := unesiBroj()

	switch opcija {
	case 1:
		trenutnaKonfiguracijaStrukture(korenskaKonfig, struktura)
	case 2:
		switch struktura {
		case "WriteAheadLog":
			korenskaKonfig = izmeniKonfiguracijuWriteAheadLoga(korenskaKonfig)
		case "BufferPool":
			korenskaKonfig = izmeniKonfiguracijuBufferPoola(korenskaKonfig)
		case "Memtable":
			korenskaKonfig = izmeniKonfiguracijuMemtablea(korenskaKonfig)
		case "SSTable":
			korenskaKonfig = izmeniKonfiguracijuSSTable(korenskaKonfig)

		}

		podaci, err := json.MarshalIndent(korenskaKonfig, "", "  ")
		if err != nil {
			errorKonverzija()
		}

		upisiKonfiguraciju(podaci)

		fmt.Println("Uspesno izmenjena konfiguracija.")
	}

	return korenskaKonfig
}

func trenutnaKonfiguracijaStrukture(korenskaKonfig map[string]map[string]uint64, struktura string) {
	ocistiProzor()
	fmt.Printf("===== Trenutna konfiguracija %s-a =====\n\n", struktura)

	konfiguracija := korenskaKonfig[struktura]
	for kljuc, vrednost := range konfiguracija {
		fmt.Printf("%s -> %d\n", kljuc, vrednost)
	}
}

func izmeniKonfiguracijuWriteAheadLoga(korenskaKonfig map[string]map[string]uint64) map[string]map[string]uint64 {
	ocistiProzor()
	fmt.Print("===== Izmena konfiguracije WriteAheadLog-a =====\n\n")
	fmt.Println("1 - Velicina segmenta")
	fmt.Println("2 - Padding")

	fmt.Print("\nUnesite jednu od ponudjenih opcija: ")
	opcija := unesiBroj()

	switch opcija {
	case 1:
		korenskaKonfig = izmeniVelicinuSegmenta(korenskaKonfig, "WriteAheadLog")
	case 2:
		korenskaKonfig = izmeniPadding(korenskaKonfig)
	}

	return korenskaKonfig
}

func izmeniKonfiguracijuBufferPoola(korenskaKonfig map[string]map[string]uint64) map[string]map[string]uint64 {
	ocistiProzor()
	fmt.Print("===== Izmena konfiguracije BufferPool-a =====\n\n")
	fmt.Println("1 - Velicina BufferPool-a")

	fmt.Print("\nUnesite jednu od ponudjenih opcija: ")
	opcija := unesiBroj()

	switch opcija {
	case 1:
		korenskaKonfig = izmeniVelicinuSegmenta(korenskaKonfig, "BufferPool")
	}

	return korenskaKonfig
}

func izmeniKonfiguracijuMemtablea(korenskaKonfig map[string]map[string]uint64) map[string]map[string]uint64 {
	ocistiProzor()
	fmt.Println("===== Izmena konfiguracije Memtable-a ====\n\n")
	fmt.Println("1 - Struktura")
	fmt.Println("2 - Broj tabela")
	fmt.Println("3 - Max elemenata")
	fmt.Println("4 - Max visina skipliste")
	fmt.Println("5 - Red B stabla")

	fmt.Print("\nUnesite jednu od ponudjenih opcija: ")
	opcija := unesiBroj()

	switch opcija {
	case 1:
		korenskaKonfig = izmeniStrukturuMemtablea(korenskaKonfig)
	case 2:
		korenskaKonfig = izmeniBrojTabela(korenskaKonfig)
	case 3:
		korenskaKonfig = izmeniMaxElemenata(korenskaKonfig)
	case 4:
		korenskaKonfig = izmeniMaxVisinuSkipliste(korenskaKonfig)
	case 5:
		korenskaKonfig = izmeniRedBStabla(korenskaKonfig)
	}
	return korenskaKonfig
}

func izmeniStrukturuMemtablea(korenskaKonfig map[string]map[string]uint64) map[string]map[string]uint64 {
	fmt.Println("Izaberite strukturu za memtable:")
	fmt.Println("1 - HashMap")
	fmt.Println("2 - Skiplista")
	fmt.Println("3 - B stablo")

	fmt.Print("Izaberite jednu od opcija: ")
	struktura := unesiBroj()

	if struktura >= 1 && struktura <= 3 {
		korenskaKonfig["Memtable"]["Struktura"] = uint64(struktura)
	} else {
		fmt.Println("Pogresan unos, konfiuracija nije izmenjena.")
	}
	return korenskaKonfig
}

func izmeniBrojTabela(korenskaKonfig map[string]map[string]uint64) map[string]map[string]uint64 {
	fmt.Print("Unesite novi broj tabela: ")
	brojTabela := unesiBroj()

	if brojTabela > 0 {
		korenskaKonfig["Memtable"]["Broj tabela"] = uint64(brojTabela)
	} else {
		fmt.Println("Broj tabela mora da bude veci od nule.")
	}
	return korenskaKonfig
}

func izmeniMaxElemenata(korenskaKonfig map[string]map[string]uint64) map[string]map[string]uint64 {
	fmt.Print("Unesite novi maksimalan broj elemenata po tabeli: ")
	maxElemenata := unesiBroj()

	if maxElemenata > 0 {
		korenskaKonfig["Memtable"]["Maksimalan broj elemenata"] = uint64(maxElemenata)
	} else {
		fmt.Println("Maksimalan broj elemenata mora da bude broj veci od nule.")
	}
	return korenskaKonfig
}

func izmeniMaxVisinuSkipliste(korenskaKonfig map[string]map[string]uint64) map[string]map[string]uint64 {
	fmt.Print("Unesite novu maksimalnu visinu skipliste: ")
	maxVisina := unesiBroj()

	if maxVisina > 0 {
		korenskaKonfig["Memtable"]["Max visina skipliste"] = uint64(maxVisina)
	} else {
		fmt.Println("Maksimalna visina skipliste mora da bude broj veci od nule.")
	}
	return korenskaKonfig
}

func izmeniRedBStabla(korenskaKonfig map[string]map[string]uint64) map[string]map[string]uint64 {
	fmt.Print("Unesite nov red B stabla: ")
	red := unesiBroj()

	if red > 0 {
		korenskaKonfig["Memtable"]["Red B stabla"] = uint64(red)
	} else {
		fmt.Println("Red B stabla mora da bude broj veci od nule.")
	}
	return korenskaKonfig
}

func izmeniKonfiguracijuSSTable(korenskaKonfig map[string]map[string]uint64) map[string]map[string]uint64 {
	ocistiProzor()
	fmt.Print("===== Izmena konfiguracije SSTable =====\n\n")
	fmt.Println("1 - SummaryStep")
	fmt.Println("2 - Velicina bloka u bajtovima)")

	fmt.Print("\nUnesite jednu od ponudjenih opcija: ")
	opcija := unesiBroj()

	switch opcija {
	case 1:
		fmt.Print("Unesite novi summaryStep: ")
		summaryStep := unesiBroj()
		korenskaKonfig["SSTable"]["SummaryStep"] = uint64(summaryStep)
	case 2:
		fmt.Print("Unesite novu velicinu bloka u bajtovima: ")
		blockSize := unesiBroj()
		korenskaKonfig["SSTable"]["BlockSize"] = uint64(blockSize)
	}

	return korenskaKonfig
}

func izmeniMaksimalanBrojElemenata(korenskaKonfig map[string]map[string]uint64) map[string]map[string]uint64 {
	fmt.Print("Unesite novi maksimalni broj elemenata: ")
	velicinaSegmenta := unesiBroj()

	korenskaKonfig["SkipList"]["Maksimalan broj elemenata"] = uint64(velicinaSegmenta)

	return korenskaKonfig
}

func izmeniVelicinuSegmenta(korenskaKonfig map[string]map[string]uint64, struktura string) map[string]map[string]uint64 {
	fmt.Print("Unesite novu velicinu segmenta: ")
	velicinaSegmenta := unesiBroj()

	korenskaKonfig[struktura]["Velicina segmenta"] = uint64(velicinaSegmenta)

	return korenskaKonfig
}

func izmeniPadding(korenskaKonfig map[string]map[string]uint64) map[string]map[string]uint64 {
	fmt.Print("Unesite novu vrednost paddinga: ")
	padding := unesiBroj()

	korenskaKonfig["WriteAheadLog"]["Padding"] = uint64(padding)

	return korenskaKonfig
}

func unesiBroj() float32 {
	reader := bufio.NewReader(os.Stdin)
	for {
		var option float32
		input, _ := reader.ReadString('\n')

		_, err := fmt.Sscan(input, &option)
		if err != nil {
			errorPoruka()
		} else {
			return option
		}
	}
}

func UcitajKonfiguraciju() map[string]map[string]uint64 {
	konfiguracija := make(map[string]map[string]uint64)

	podaci, err := os.ReadFile("Configuration/resources/configuration_file.json")
	if err != nil {
		errorFajl()
	}

	err = json.Unmarshal(podaci, &konfiguracija)
	if err != nil {
		errorKonverzija()
	}

	return konfiguracija
}

func upisiKonfiguraciju(podaci []byte) error {
	err := os.WriteFile("Configuration/resources/configuration_file.json", podaci, 0644)
	if err != nil {
		return err
	}
	return nil
}

func ocistiProzor() {
	fmt.Print("\033[H\033[2J")
}

func errorPoruka() {
	fmt.Println("Pogresan unos, pokusajte ponovo.")
}

func errorKonverzija() {
	fmt.Println("Greska prilikom konverzije.")
}

func errorFajl() {
	fmt.Println("Greska prilikom otvaranja fajla.")
}
