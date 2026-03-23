package main

import (
	"fmt"
)

func main() {
	var opcija int
	var hash1, hash2 string

	for {
		fmt.Println("SIMHASH")
		fmt.Println("1. Napravi SimHash za dva dokumenta")
		fmt.Println("2. Proceni slicnost dokumenata")
		fmt.Println("3. Izbrisi SimHash")
		fmt.Println("0. Izadji")
		fmt.Println("Izaberite opciju:")
		fmt.Scanln(&opcija)

		switch opcija {
		case 1:
			var fajl1 string
			fmt.Println("Unesite ime prvog fajla: ")
			fmt.Scanln(&fajl1)

			tekst1, err := ucitaj(fajl1)
			if err != nil {
				fmt.Println("Greska pri ucitavanju.")
				continue
			}
			hash1 = SimHash(tekst1)
			fmt.Printf("SimHash za fajl '%s' je: %s\n", fajl1, hash1)

			var fajl2 string
			fmt.Println("Unesite ime drugog fajla: ")
			fmt.Scanln(&fajl2)

			tekst2, err2 := ucitaj(fajl2)
			if err2 != nil {
				fmt.Println("Greska pri ucitavanju.")
				continue
			}
			hash2 = SimHash(tekst2)
			fmt.Printf("SimHash za fajl '%s' je: %s\n", fajl2, hash2)

		case 2:
			if hash1 == "" || hash2 == "" {
				fmt.Println("Morate imati 2 dokumenta!")
				continue
			}

			d := HejmingovaDistanca(hash1, hash2)
			prag := len(hash1) / 10

			if d <= prag {
				fmt.Println("Dokumenti su slicni")
			} else {
				fmt.Println("Dokumenti nisu slicni")
			}

		case 3:
			hash1 = ""
			hash2 = ""
			fmt.Println("SimHash vrednosti su obrisane")

		case 0:
			fmt.Println("Izlazak iz programa...")
			return

		default:
			fmt.Println("Nepostojeca opcija")
		}
	}

}
