package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

// dodati ostale strukture u konfiguraciju
func meni() {
	konfiguracija := ucitajKonfiguraciju()

	fmt.Print("===== Konfiguracija =====\n\n")
	fmt.Println("1 - WriteAheadLog")

	fmt.Print("\nUnesite jednu od ponudjenih opcija: ")
	opcija := unesiBroj()

	switch opcija {
	case 1:
		konfiguracijaWriteAheadLoga(konfiguracija)
	}
}

func konfiguracijaWriteAheadLoga(korenskaKonfig map[string]map[string]uint64) map[string]map[string]uint64 {
	ocistiProzor()
	fmt.Print("===== Konfiguracija WriteAheadLog-a =====\n\n")
	fmt.Println("1 - Trenutna konfiguracija")
	fmt.Println("2 - Izmeni konfiguraciju")

	fmt.Print("\nUnesite jednu od ponudjenih opcija: ")
	opcija := unesiBroj()

	switch opcija {
	case 1:
		trenutnaKonfiguracijaWriteAheadLoga(korenskaKonfig)
	case 2:
		korenskaKonfig = izmeniKonfiguracijuWriteAheadLoga(korenskaKonfig)

		podaci, err := json.MarshalIndent(korenskaKonfig, "", "  ")
		if err != nil {
			errorKonverzija()
		}

		upisiKonfiguraciju(podaci)

		fmt.Println("Uspesno izmenjena konfiguracija WriteAheadLog-a.")
	}

	return korenskaKonfig
}

func trenutnaKonfiguracijaWriteAheadLoga(korenskaKonfig map[string]map[string]uint64) {
	ocistiProzor()
	fmt.Print("===== Trenutna konfiguracija WriteAheadLog-a =====\n\n")

	wal := korenskaKonfig["WriteAheadLog"]
	for kljuc, vrednost := range wal {
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
		korenskaKonfig = izmeniVelicinuSegmenta(korenskaKonfig)
	case 2:
		korenskaKonfig = izmeniPadding(korenskaKonfig)
	}

	return korenskaKonfig
}

func izmeniVelicinuSegmenta(korenskaKonfig map[string]map[string]uint64) map[string]map[string]uint64 {
	fmt.Print("Unesite novu velicinu segmenta: ")
	velicinaSegmenta := unesiBroj()

	korenskaKonfig["WriteAheadLog"]["Velicina segmenta"] = uint64(velicinaSegmenta)

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

func ucitajKonfiguraciju() map[string]map[string]uint64 {
	konfiguracija := make(map[string]map[string]uint64)

	podaci, err := os.ReadFile("resources/configuration_file.json")
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
	err := os.WriteFile("resources/configuration_file.json", podaci, 0644)
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

func main() {
	meni()
}
