package main

import (
	blockmanager "NASP_projekat/BlockManager"
	bufferpool "NASP_projekat/BufferPool"
	configuration "NASP_projekat/Configuration"
	wal "NASP_projekat/WriteAheadLog"
	probabilistics "NASP_projekat/probabilistics"
	cms "NASP_projekat/probabilistics/CountMinSketch"
	"fmt"
	"os"
)

func meni() {
	for {
		fmt.Println("===== No-SQL Engine =====")
		fmt.Println()
		fmt.Println("1 - Probabilisticke strukture")
		fmt.Println("2 - Konfiguracija")
		fmt.Println("3 - Put") // testiranje
		fmt.Println("4 - Test Bufferpool")
		fmt.Println("5 - Test Block Manager")
		fmt.Println("0 - Izlaz")

		fmt.Print("Unesite jednu od ponudjenih opcija: ")
		opcija := cms.UnesiBroj()

		switch opcija {
		case 1:
			probabilistics.ProbabilistickeStruktureMeni()
		case 2:
			configuration.KonfiguracijaMeni()
		case 3:
			wal.Unesi("delete", "operation", "doomsday")
			// wal.Ispisi()
		case 4:
			vrednost := bufferpool.Get("operation")
			fmt.Println("Vrednost: ", vrednost)
		case 5:
			blockmanager.TestBlockManager()
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
