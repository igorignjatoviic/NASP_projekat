package bufferpool

import (
	"fmt"
)

func BufferPoolMeni() {
	var bp *BufferPool
	var opcija int

	for {
		fmt.Println("\nBuffer pool")
		fmt.Println("1. Napravi BufferPool")
		fmt.Println("2. Upisi blok")
		fmt.Println("3. Procitaj blok")
		fmt.Println("4. Sacuvaj u fajl")
		fmt.Println("5. Ucitaj iz fajla")
		fmt.Println("6. Obrisi BufferPool")
		fmt.Println("7. Prikazi stanje")
		fmt.Println("0. Izlaz")
		fmt.Println("Izaberite opciju:")
		fmt.Scan(&opcija)

		switch opcija {
		case 1:
			var vel int
			fmt.Print("Unesite velicinu:")
			fmt.Scan(&vel)
			bp = Napravi(vel)
			fmt.Println("Buffer pool je napravljen.")
		case 2:
			if bp == nil {
				fmt.Println("Buffer pool nije napravljen.")
				continue
			}
			var id string
			var blok string
			fmt.Println("Unesite id bloka:")
			fmt.Scan(&id)
			fmt.Println("Unesite blok:")
			fmt.Scan(&blok)

			bp.Upisi(id, []byte(blok))
		case 3:
			if bp == nil {
				fmt.Println("Buffer pool nije napravljen.")
				continue
			}
			var id string
			fmt.Println("Unesite id bloka:")
			fmt.Scan(&id)

			pod := bp.Citaj(id)
			if pod != nil {
				fmt.Println("Trazeni blok: ", string(pod))
			}
		case 4:
			/*
				if bp == nil {
					fmt.Println("Buffer pool nije napravljen.")
					continue
				}
				var fajl string
				fmt.Println("Unesi ime fajla: ")
				fmt.Scan(&fajl)*/

			err := bp.SacuvajUFajl("BufferPool/fajl.txt")
			if err != nil {
				fmt.Println("Greska pri otvaranju fajla.")
			} else {
				fmt.Println("Sacuvano")
			}
		case 5:
			/*
				var fajl string
				fmt.Println("Unesi ime fajla: ")
				fmt.Scan(&fajl)*/

			var vel int
			fmt.Println("Unesi velicinu buffer poola: ")
			fmt.Scan(&vel) // konfiguracija

			ucitanbp, err := UcitajIzFajla("BufferPool/fajl.txt", vel)
			if err != nil {
				fmt.Println("Greska pri otvaranju fajla.")
			} else {
				fmt.Println("Ucitano")
				bp = ucitanbp
			}
		case 6:
			bp = nil
			fmt.Println("Obrisano")
		case 7:
			if bp == nil {
				fmt.Println("Ne postoji.")
			} else {
				fmt.Println("Redosled: ", bp.redosled)
				fmt.Println("Podaci:")
				for k, v := range bp.podaci {
					fmt.Printf("%s:%s\n", k, string(v))
				}
			}
		case 0:
			return
		default:
			fmt.Println("Nepostojeca opcija.")
		}
	}
}
