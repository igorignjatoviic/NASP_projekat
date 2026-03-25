package bufferpool

import "fmt"

func test() {
	GenerisiPocetniFajl()
	niz := Ucitaj()

	niz = OsveziBufferPool(niz, *NoviTuple("put", "z", "z"))
	niz = OsveziBufferPool(niz, *NoviTuple("put", "z", "y"))

	bp := NoviBufferPool()
	bp.Unesi(niz)

	niz = Ucitaj()

	niz = OsveziBufferPool(niz, *NoviTuple("delete", "operation", "doomsday"))
	bp.Unesi(niz)
	niz = Ucitaj()

	fmt.Println("krajnji niz")
	for _, podatak := range niz {
		fmt.Println(podatak)
	}

}
