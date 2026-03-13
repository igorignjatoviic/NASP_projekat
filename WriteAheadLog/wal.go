package main

func main() {
	wal := WriteAheadLog{}

	wal.unesi("put", "cao", "igor")
	wal.unesi("put", "hejjj", "neda")
	wal.unesi("put", "diskretna", "matematika")
	wal.unesi("delete", "cao", "")

	wal.ucitajWriteAheadLog()
}
