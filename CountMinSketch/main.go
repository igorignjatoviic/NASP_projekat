package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
)

func meni() {
	// isprobavanje
	fmt.Println("===== Count-Min Sketch =====")
	fmt.Println()
	fmt.Println("1 - Kreirajte novu instancu")
	fmt.Println("2 - Brisanje postojece instance")
	fmt.Println("3 - Dodavanje novog dogadjaja u neku instancu")
	fmt.Println("4 - Provera ucestalosti dogadjaja u nekoj instanci")
	fmt.Println("0 - Izlazak")

	fmt.Print("Unesite jednu od ponudjenih opcija: ")
	option := unesiBroj()

	switch option {
	case 1:
		kreirajCountMinSketch()
	case 2:
		izbrisiCountMinSketch()
	case 3:
		dodavanjeDogadjaja()
	case 4:
		proveraUcestalostiDogadjaja()
	case 0:
		izlazakIzAplikacije()
	default:
		fmt.Println("Pogresan unos")
	}

}

func kreirajCountMinSketch() {
	ocistiProzor()
	fmt.Println("===== Novi Count-Min Sketch =====")
	fmt.Println()

	ime, lokacija := unesiteNazivCountMinSketcha(true)
	cms := CountMinSketch{}

	var epsilon float32
	for {
		fmt.Print("Unesite vrednost parametra epsilon (izmedju 0 i 1): ")
		epsilon = unesiBroj()
		if 0 < epsilon && epsilon < 1 {
			break
		}
	}
	cms.izracunajM(float64(epsilon))

	var delta float32
	for {
		fmt.Print("Unesite vrednost parametra delta (izmedju 0 i 1): ")
		delta = unesiBroj()
		if 0 < delta && delta < 1 {
			break
		}
	}
	cms.izracunajK(float64(delta))

	cms.napraviHashFunkcije()
	cms.napraviTabelu()

	sacuvajCountMinSketch(lokacija, &cms)
	fmt.Println()
	fmt.Printf("Uspesno ste kreirali Count-Min Sketch objekat pod imenom '%s'.\n", ime)

	meni()
}

func izbrisiCountMinSketch() {
	ocistiProzor()
	fmt.Println("===== Brisanje Count-Min Sketcha =====")
	fmt.Println()

	ime, lokacija := unesiteNazivCountMinSketcha(false)
	os.Remove(lokacija)

	fmt.Println()
	fmt.Printf("Uspesno ste izbrisali Count-Min Sketch objekat pod nazivom '%s'.\n", ime)

	meni()
}

func dodavanjeDogadjaja() {
	ocistiProzor()
	fmt.Println("===== Dodavanje dogadjaja =====")
	fmt.Println()

	ime, lokacija := unesiteNazivCountMinSketcha(false)

	cms, err := ucitajCountMinSketch(lokacija)
	if err != nil {
		errorPoruka()
	}

	fmt.Print("Unesite koliko elemenata zelite da unesete u CountMin-Sketch: ")
	n := int(unesiBroj())
	podaci := make([]string, n)
	for i := 0; i < n; i++ {
		fmt.Printf("Unesite %d. podatak: ", i+1)
		var podatak string
		_, err := fmt.Scan(&podatak)
		if err != nil {
			errorPoruka()
		}
		podaci[i] = podatak
	}
	fmt.Scanln()

	for _, podatak := range podaci {
		cms.unesi([]byte(podatak))
	}

	sacuvajCountMinSketch(lokacija, &cms)
	fmt.Println()
	fmt.Printf("Podaci su uspesno uneti u Count-Min Sketch pod imenom '%s'.\n", ime)

	meni()
}

func proveraUcestalostiDogadjaja() {
	ocistiProzor()
	fmt.Println("===== Provera ucestalosti dogadjaja =====")
	fmt.Println()

	ime, lokacija := unesiteNazivCountMinSketcha(false)
	cms, err := ucitajCountMinSketch(lokacija)
	if err != nil {
		errorPoruka()
	}

	fmt.Print("Unesite podatak koji zelite da proverite: ")
	podatak := unesiString()
	rezultat := cms.pretrazi([]byte(podatak))
	fmt.Printf("Minimalan broj ponavljanja podatka '%s' u Count-Min Sketchu '%s': %d.\n", podatak, ime, rezultat)
	fmt.Scanln()

	meni()
}

func izlazakIzAplikacije() {
	ocistiProzor()
	fmt.Println("Izasli ste iz aplikacije.")
	os.Exit(0)
}

func sacuvajCountMinSketch(lok string, cms *CountMinSketch) error {
	f, err := os.Create(lok)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := binary.Write(f, binary.BigEndian, uint32(cms.getK())); err != nil {
		return err
	}
	if err := binary.Write(f, binary.BigEndian, uint32(cms.getM())); err != nil {
		return err
	}

	for _, funkcija := range cms.getHashFunkcije() {
		if err := binary.Write(f, binary.BigEndian, funkcija.Seed); err != nil {
			return err
		}
	}

	for i := uint32(0); i < uint32(cms.getK()); i++ {
		if err := binary.Write(f, binary.BigEndian, cms.getTabela()[i]); err != nil {
			return err
		}
	}

	return nil
}

func ucitajCountMinSketch(lok string) (CountMinSketch, error) {
	cms := CountMinSketch{}
	f, err := os.Open(lok)
	if err != nil {
		return cms, err
	}
	defer f.Close()

	var k, m uint32
	binary.Read(f, binary.BigEndian, &k)
	binary.Read(f, binary.BigEndian, &m)

	hashFunkcije := make([]HashWithSeed, k)
	for i := uint32(0); i < k; i++ {
		if err := binary.Read(f, binary.BigEndian, &hashFunkcije[i].Seed); err != nil {
			return cms, err
		}
	}

	tabela := make([][]uint32, k)
	for i := uint32(0); i < k; i++ {
		tabela[i] = make([]uint32, m)
		if err := binary.Read(f, binary.BigEndian, tabela[i]); err != nil {
			return cms, err
		}
	}

	cms.setK(uint(k))
	cms.setM(uint(m))
	cms.setTabela(tabela)
	cms.setHashFunkcije(hashFunkcije)

	return cms, nil
}

func unesiteNazivCountMinSketcha(kopija bool) (string, string) {
	lokacija := ""
	ime := ""
	for {
		fmt.Print("Unesite naziv Count-Min Sketcha: ")
		ime = unesiString()
		fmt.Scanln()
		lokacija = "resources/" + ime + ".bin"
		if proveriPutanju(lokacija) && kopija {
			errorPoruka()
		} else {
			break
		}
	}

	return ime, lokacija
}

func proveriPutanju(lok string) bool {
	_, err := os.Stat(lok)
	if err == nil {
		return true
	} else {
		return false
	}
}

func unesiString() string {
	for {
		var input string
		_, err := fmt.Scan(&input)
		if err != nil {
			errorPoruka()
		} else {
			return input
		}
	}
}

func unesiBroj() float32 {
	reader := bufio.NewReader(os.Stdin)
	for {
		var option float32
		input, _ := reader.ReadString('\n')

		_, err := fmt.Sscan(input, &option)
		if err != nil {
			errorPoruka()
		} else {
			return option
		}
	}
}

func ocistiProzor() {
	fmt.Print("\033[H\033[2J")
}

func errorPoruka() {
	fmt.Println("Pogresan unos, pokusajte ponovo.")
}

func main() {
	meni()
}
