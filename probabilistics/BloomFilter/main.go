package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

func unesiString(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	text, _ := reader.ReadString('\n') //uklanjam n

	if len(text) > 0 && text[len(text)-1] == '\n' {
		text = text[:len(text)-1]
	}
	text = strings.TrimSpace(text)
	return text
}

func unesiBroj(prompt string) int {
	for {
		var x int
		fmt.Print(prompt)
		_, err := fmt.Scan(&x)
		if err != nil {
			fmt.Println("Pogrešan unos, pokusaj ponovo.")
			fmt.Scanln() // cistimo ulaz
		} else {
			fmt.Scanln()
			return x
		}
	}
}

func main() {
	const filename = "bloomfilter.dat"
	seed := uint32(time.Now().Unix()) // moze nesto randomm!
	bf := NewBloomFilter(1000, 0.01, seed)

	for {
		fmt.Println("\n--- BLOOM FILTER ---")
		fmt.Println("1. Dodaj element")
		fmt.Println("2. Proveri element")
		fmt.Println("3. Serijalizuj Bloom Filter")
		fmt.Println("4. Učitaj Bloom Filter")
		fmt.Println("0. Izlaz")
		choice := unesiBroj("Izaberite opciju: ")

		switch choice {
		case 1:
			elem := unesiString("Unesite element: ")
			bf.Dodaj([]byte(elem))
			fmt.Printf("Dodato: %s\n", elem)
		case 2:
			elem := unesiString("Unesite element za proveru: ")
			if bf.Proveri([]byte(elem)) {
				fmt.Printf("%s je moguće prisutan u filteru\n", elem)
			} else {
				fmt.Printf("%s definitivno nije u filteru\n", elem)
			}
		case 3:
			err := bf.SerijalizacijaUFajl(filename)
			if err != nil {
				fmt.Println("Greška pri serijalizaciji:", err)
			} else {
				fmt.Println("Bloom Filter sačuvan u", filename)
			}
		case 4:

			bf2, err := DeserijalizacijaIzFajla(filename)
			if err != nil {
				fmt.Println("Greška pri učitavanju:", err)
			} else {
				bf = bf2
				fmt.Println("Bloom Filter učitan iz", filename)
			}
		case 0:
			fmt.Println("Izlaz")
			return
		default:
			fmt.Println("Pogrešan unos, probajte ponovo.")
		}
	}
}
