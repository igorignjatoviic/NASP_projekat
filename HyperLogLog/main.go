package main

import (
	"bufio"
	"fmt"
	"os"
)

func unesiBroj() uint8 { // uint8 jer ce mi to trebati za p, a svakako je validno i za odabir opcije x
	reader := bufio.NewReader(os.Stdin)
	for {
		var x uint8
		unos, _ := reader.ReadString('\n')
		_, err := fmt.Sscan(unos, &x)
		if err != nil {
			fmt.Println("Pogresan unos! Probajte ponovo.")
		} else {
			return x
		}
	}
}

func unesiString() string {
	reader := bufio.NewReader(os.Stdin)
	for {
		var s string
		unos, _ := reader.ReadString('\n')
		_, err := fmt.Sscan(unos, &s)
		if err != nil {
			fmt.Println("Pogresan unos! Probajte ponovo.")
		} else {
			return s
		}
	}
}

func proveriPutanju(lokacija string) bool {
	_, err := os.Stat(lokacija)
	if err != nil {
		return false
	} else {
		return true
	}
}

func unesiNazivHLL(novi bool) (string, string) {
	lokacija := ""
	naziv := ""
	for {
		fmt.Print("Unesite naziv: ")
		naziv = unesiString()
		//fmt.Scanln()
		lokacija = naziv + ".bin"
		if novi {
			if !proveriPutanju(lokacija) {
				break
			} else {
				fmt.Println("Greska, putanja ne postoji.")
			}
		} else {
			if proveriPutanju(lokacija) {
				break
			} else {
				fmt.Println("Greska, putanja ne postoji.")
			}
		}

	}

	return naziv, lokacija
}

func napraviHLLmeni() {
	naziv, lokacija := unesiNazivHLL(true)
	fmt.Println("Unesite preciznost: ")
	p := unesiBroj()
	hll, err := napraviHyperLogLog(p)
	if err != nil {
		fmt.Println(err)
	} else {
		err = hll.Sacuvaj(lokacija)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf("HLL '%s' je uspеšno kreiran!\n", naziv)
		}
	}
}

func dodajUHLLmeni() {
	naziv, lokacija := unesiNazivHLL(false)
	hll, err := ucitajHyperLogLog(lokacija)
	if err != nil {
		fmt.Println("Ne postoji HyperLogLog sa unetim nazivom!")
	}
	fmt.Println("Unesite koliko elemenata zelite da unesete: ")
	n := int(unesiBroj())
	podaci := make([]string, n)
	for i := 0; i < n; i++ {
		fmt.Printf("Unesite %d. podatak: ", i+1)
		var podatak string
		_, err := fmt.Scan(&podatak)
		if err != nil {
			fmt.Println("Pogresan unos! Probajte ponovo.")
		}
		podaci[i] = podatak
	}
	fmt.Scanln()
	for _, podatak := range podaci {
		hll.DodajString(podatak) // DodajString mi pretvara podatak u niz bajtova i dodaje u hll
	}
	hll.Sacuvaj(lokacija)
	fmt.Printf("Podaci su uneti u HyperLogLog '%s'.\n", naziv)
}

func proceniKardinalnostHLLmeni() {
	naziv, lokacija := unesiNazivHLL(false)
	hll, err := ucitajHyperLogLog(lokacija)
	if err != nil {
		fmt.Println("Greska pri ucitavanju HyperLogLog!")
		return
	}

	procena := hll.Estimate()
	fmt.Printf("Procijenjena kardinalnost HyperLogLog '%s' je: %.2f\n", naziv, procena)
	fmt.Printf("Zaokruzeno: %.0f\n", procena)
}

func izbrisiHLL() {
	naziv, lokacija := unesiNazivHLL(false)
	os.Remove(lokacija)
	fmt.Printf("HyperLogLog '%s' je uspesno obrisan!\n", naziv)
}

func izlazak() {
	fmt.Println("\nDa li ste sigurni da zelite da izadjete iz programa? (da/ne)")
	fmt.Print(">> ")
	odgovor := unesiString()

	if odgovor == "da" || odgovor == "DA" || odgovor == "Da" {
		fmt.Println("Dovidjenja!")
		os.Exit(0)
	}
}

func main() {

	for {
		fmt.Println("\n--- HYPER LOG LOG ---")

		fmt.Println("Izaberite opciju:")
		fmt.Println("1. Napravi HyperLogLog")
		fmt.Println("2. Dodaj element u HyperLogLog")
		fmt.Println("3. Proceni kardinalnost HyperLogLog strukture")
		fmt.Println("4. Izbrisi HyperLogLog")
		fmt.Println("0. Izadji")

		fmt.Println(">> ")
		x := unesiBroj()

		switch x {
		case 1:
			napraviHLLmeni()
		case 2:
			dodajUHLLmeni()
		case 3:
			proceniKardinalnostHLLmeni()
		case 4:
			izbrisiHLL()
		case 0:
			izlazak()
		default:
			fmt.Println("Pogresan unos! Probajte ponovo.")
		}
	}

}
