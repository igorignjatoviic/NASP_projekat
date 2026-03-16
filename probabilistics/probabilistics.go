package probabilistics

import (
	bf "NASP_projekat/Probabilistics/BloomFilter"
	cms "NASP_projekat/Probabilistics/CountMinSketch"
	hll "NASP_projekat/Probabilistics/HyperLogLog"
	sh "NASP_projekat/Probabilistics/Simhash"
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
