package main

import (
	probabilistics "NASP_projekat/Probabilistics"
	cms "NASP_projekat/Probabilistics/CountMinSketch"
	"fmt"
	"os"
)

func meni() {
	for {
		fmt.Println("===== No-SQL Engine =====")
		fmt.Println("1 - Probabilisticke strukture")
		fmt.Println("2 - Konfiguracija")
		fmt.Println("0 - Izlaz")

		fmt.Print("Unesite jednu od ponudjenih opcija: ")
		opcija := cms.UnesiBroj()

		switch opcija {
		case 1:
			probabilistics.ProbabilistickeStruktureMeni()
		case 2:
			// konfiguracija
		case 0:
			ocistiProzor()
			fmt.Println("Izasli ste iz aplikacije.")
			os.Exit(0)
		default:
			fmt.Println("Pogresan unos, pokusajte ponovo.")
		}
	}
}

func ocistiProzor() {
	fmt.Print("\033[H\033[2J")
}

func main() {
	meni()
}
