package pkg

type entryItem struct {
	flag    uint32
	sibling uint32
	child   uint32
	offset  uint32
}

type Dict struct {
	charaMap  []uint16
	beginNode []uint16
	innerNode []uint16
	chara     []rune
	word      []rune
	entries   []entryItem
}
