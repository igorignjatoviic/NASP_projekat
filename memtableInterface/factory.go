package memtableinterface

import (
	"NASP_projekat/memtables"
	btree "NASP_projekat/memtables/BTreeMemtable"
	hashmap "NASP_projekat/memtables/HashMemtable"
	skiplist "NASP_projekat/probabilistics/SkipList"
)

func NovaStruktura(tip string, maxElemenata int, maxVisina int, redBStabla int) memtables.Struktura {
	switch tip {
	case "hashmap":
		return hashmap.NapraviHashMemtable(maxElemenata)
	case "skiplist":
		return skiplist.NovaSkipLista(maxVisina, maxElemenata)
	case "btree":
		return btree.NovoBStablo(redBStabla, maxElemenata)
	default:
		return hashmap.NapraviHashMemtable(maxElemenata)
	}
}
