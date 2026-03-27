package probabilistics

import (
	bf "NASP_projekat/probabilistics/BloomFilter"
	cms "NASP_projekat/probabilistics/CountMinSketch"
	hll "NASP_projekat/probabilistics/HyperLogLog"
	sh "NASP_projekat/probabilistics/Simhash"
	"fmt"
)

func ProbabilistickeStruktureMeni() {
	for {
		ocistiProzor()

		fmt.Println("===== Probabilisticke strukture =====")
		fmt.Println("1 - BloomFilter")
		fmt.Println("2 - CountMinSketch")
		fmt.Println("3 - HyperLogLog")
		fmt.Println("4 - SimHash")
		fmt.Println("0 - Nazad")

		fmt.Println()
		fmt.Print("Unesite jednu od ponudjenih opcija: ")
		opcija := cms.UnesiBroj()

		switch opcija {
		case 1:
			bf.BloomFilterMeni()
		case 2:
			cms.CountMinSketchMeni()
		case 3:
			hll.HyperLogLogMeni()
		case 4:
			sh.SimHashMeni()
		case 0:
			return
		default:
			errorPoruka()
		}
	}
}

func ocistiProzor() {
	fmt.Print("\033[H\033[2J")
}

func errorPoruka() {
	fmt.Println("Pogresan unos, pokusajte ponovo.")
}
