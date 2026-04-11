package main

import (
	"fmt"
)

func main() {
	var blokovi []string
	var stablo *MerkleStablo
	var stablo2 *MerkleStablo
	for {
		fmt.Println("\nMerkle stablo")
		fmt.Println("1. Napravi stablo")
		fmt.Println("2. Dodaj blok")
		fmt.Println("3. Obrisi blok")
		fmt.Println("4. Uporedi sa drugim stablom")
		fmt.Println("5. Prikazi root hash")
		fmt.Println("6. Ispisi stablo")
		fmt.Println("7. Sacuvaj serijalizovano stablo")
		fmt.Println("8. Ucitaj stablo iz fajla")
		fmt.Println("0. Izlaz")

		var opcija int
		fmt.Println("Izbaerite opciju:")
		fmt.Scan(&opcija)

		switch opcija {
		case 1:
			var n int
			fmt.Println("Koliko blokova unosite:")
			fmt.Scan(&n)

			blokovi = []string{}
			for i := 0; i < n; i++ {
				var b string
				fmt.Println("Uneiste blok:")
				fmt.Scan(&b)
				blokovi = append(blokovi, b)
			}
			stablo = Napravi(blokovi)
			fmt.Println("Stablo je napravljeno.")
		case 2:
			var b string
			fmt.Println("Uneiste blok:")
			fmt.Scan(&b)
			blokovi = append(blokovi, b)
			stablo = Napravi(blokovi)
			fmt.Println("Dodato.")
		case 3:
			if len(blokovi) == 0 {
				fmt.Println("Nema blokova!")
				continue
			}

			var i int
			fmt.Println("Uneiste indeks bloka:")
			fmt.Scan(&i)
			if i < 0 || i >= len(blokovi) {
				fmt.Println("Nepostojeci id")
				continue
			}

			blokovi = append(blokovi[:i], blokovi[i+1:]...)
			stablo = Napravi(blokovi)
			fmt.Println("Obrisano.")
		case 4:
			var drugi []string

			var n int
			fmt.Println("Koliko blokova unosite:")
			fmt.Scan(&n)

			for i := 0; i < n; i++ {
				var b string
				fmt.Println("Uneiste blok:")
				fmt.Scan(&b)
				drugi = append(drugi, b)
			}
			stablo2 = Napravi(drugi)
			if stablo.koren.hes == stablo2.koren.hes {
				fmt.Printf("Isti su.")
			} else {
				Uporedi(stablo.koren, stablo2.koren)
			}
		case 5:
			if stablo == nil || stablo.koren == nil {
				fmt.Println("Stablo ne postoji!")
				continue
			}
			fmt.Println("Root hash:", stablo.koren.hes)
		case 6:
			if stablo == nil {
				fmt.Println("Stablo ne postoji!")
				continue
			}
			Ispisi(stablo.koren, 0)
		case 7:
			err := Sacuvaj(stablo.koren, "stablo.txt")
			if err != nil {
				fmt.Println("Greska:", err)
			} else {
				fmt.Println("Sacuvano u stablo.txt")
			}

		case 8:
			root, err := Ucitaj("stablo.txt")
			if err != nil {
				fmt.Println("Greska:", err)
			} else {
				stablo = &MerkleStablo{koren: root}
				fmt.Println("Ucitan TXT:")
				Ispisi(stablo.koren, 0)
			}
		case 0:
			return
		default:
			fmt.Println("Nepostojeca opcija.")
		}

	}
}
