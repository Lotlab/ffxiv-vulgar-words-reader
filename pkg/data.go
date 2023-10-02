package pkg

type entryItem struct {
	flag    uint32
	sibling uint32
	child   uint32
	offset  uint32
}

type Dict struct {
	charaReplace map[uint16]uint16
	charaSkip    map[uint16]bool
	charaBlock   []uint16
	beginNode    []uint16
	innerNode    []uint16
	chara        []rune
	word         []rune
	entries      []entryItem
}
