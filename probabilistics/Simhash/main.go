package main

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"os"
	"strings"
)

func GetHashAsString(data []byte) string {
	hash := md5.Sum(data)
	res := ""
	for _, b := range hash {
		res = fmt.Sprintf("%s%b", res, b)
	}
	return res
}

func SimHash(podaci []string) string {
	//const b = 128 //jer md5 ima 128bita i kad je const ne mora :
	//koristimo var jer je dinamicki niz
	var tezine []int //pravimo niz b brojeva koji predstavljaju zbir bitova svih reci
	//prolazimo kroz sve reci
	for _, rec := range podaci { //ovde mora _ jer ovo uvek vraca i gresku
		hes := GetHashAsString([]byte(rec)) //dobijamo hash vrednost te reci

		if tezine == nil {
			tezine = make([]int, len(hes))
		}

		// ako je hes kraći od tezine, proširimo tezine
		if len(hes) > len(tezine) {
			diff := len(hes) - len(tezine)
			tezine = append(tezine, make([]int, diff)...)
		}

		for i := 0; i < len(hes); i++ {
			if hes[i] == '1' { //pazi ovde je hes vrednost string
				tezine[i]++
			} else {
				tezine[i]--
			}
		}

	}
	//kad saberemo bitove svih reci dobijeni broj pretvaramo u bajtove
	rezultat := ""
	for _, t := range tezine {
		if t > 0 {
			rezultat += "1"
		} else {
			rezultat += "0"
		}
	}
	return rezultat

}

func ucitaj(fajl string) ([]string, error) {
	f, err := os.Open(fajl)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reci := []string{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		linija := scanner.Text() //reci odvojene razmakom
		reci_u_liniji := strings.Fields(linija)
		reci = append(reci, reci_u_liniji...)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return reci, nil
}

func HammingDistance(a, b string) int {
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}

	dist := 0
	for i := 0; i < minLen; i++ {
		if a[i] != b[i] {
			dist++
		}
	}

	// opciono dodamo razliku dužina ako nizovi nisu iste dužine
	dist += abs(len(a) - len(b))
	return dist
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func main() {
	f1, err := ucitaj("tekst1.txt")
	if err != nil {
		fmt.Println("Greksa pri otvaranju")
		return
	}
	f2, err := ucitaj("tekst2.txt")
	if err != nil {
		fmt.Println("Greksa pri otvaranju")
		return
	}
	br1 := SimHash(f1)
	br2 := SimHash(f2)

	d := HammingDistance(br1, br2)
	prag := len(br1) / 10

	if d <= prag {
		fmt.Println("Dokumenti su slični")
	} else {
		fmt.Println("Dokumenti nisu slični")
	}

}

//PROVERI DA LI SMO NA CASU RADILI LAKSE UCITAVANJE
//I VRATI SE NA ONO SA VELICINOM TEZINA I ZASTO NE MOZE SA B KONST
//I PROUCI HEJMINGOVU DISTANCU
