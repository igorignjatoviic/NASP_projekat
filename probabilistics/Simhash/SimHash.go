package main

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"os"
	"strings"
)

// prima string i vraca njenu hash vr kao niz bajtova
func PretvoriUBinString(data []byte) string {
	hash := md5.Sum(data) //vraca niz od 16bajtova
	res := ""
	//pretvara u bin br uz dodavanje vodecih bitova da bi sve reci bile iste duzine
	for _, b := range hash {
		res = fmt.Sprintf("%s%08b", res, b)
	}
	return res
}

// prosledjuje se dokument a vraca se jedinstvena vrednost tog dokumeta
func SimHash(podaci []string) string {
	//koristimo var jer je dinamicki niz
	var tezine []int //pravimo niz b brojeva koji predstavljaju zbir bitova svih reci

	//prolazimo kroz sve reci
	for _, rec := range podaci {
		hes := PretvoriUBinString([]byte(rec)) //dobijamo hash vrednost te reci

		if tezine == nil {
			tezine = make([]int, len(hes))
		}

		// ako je hes kraći od tezine, proširimo tezine
		if len(hes) > len(tezine) {
			razlika := len(hes) - len(tezine)
			tezine = append(tezine, make([]int, razlika)...)
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

// ucitava fajl vraca ceo tekst
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

// prima 2 reci i vraca hjihovu razliku
func HejmingovaDistanca(a, b string) int {
	dist := 0
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			dist++
		}
	}
	return dist
}
