package pkg

import (
	"encoding/binary"
	"fmt"
)

type blockInfo struct {
	offset uint32
	length uint32
}

func NewDict(data []byte) (*Dict, error) {
	dict := Dict{}

	const offsetStart = 0x8724
	const mapStart = 0x8750
	const mapSize = 0x200

	// 读取偏移量
	blocks := make([]blockInfo, 5)
	offset := offsetStart
	for i := 0; i < len(blocks); i++ {
		val := binary.LittleEndian.Uint32(data[offset : offset+4])
		blocks[i].offset = val + mapStart + mapSize
		offset += 4
	}
	for i := 0; i < len(blocks); i++ {
		blocks[i].length = binary.LittleEndian.Uint32(data[offset : offset+4])
		offset += 4
	}

	// 字符映射表
	dict.charaMap = make([]uint16, 256)
	for i := 0; i < mapSize/2; i++ {
		ofs := mapStart + i*2
		dict.charaMap[i] = binary.LittleEndian.Uint16(data[ofs : ofs+2])
	}

	// 首个节点表
	dict.beginNode = make([]uint16, blocks[0].length/2)
	for i := 0; i < int(blocks[0].length)/2; i++ {
		ofs := int(blocks[0].offset) + i*2
		dict.beginNode[i] = binary.LittleEndian.Uint16(data[ofs : ofs+2])
	}

	// 内部节点表
	dict.innerNode = make([]uint16, blocks[1].length/2)
	for i := 0; i < int(blocks[1].length)/2; i++ {
		ofs := int(blocks[1].offset) + i*2
		dict.innerNode[i] = binary.LittleEndian.Uint16(data[ofs : ofs+2])
	}

	// 字符表
	dict.chara = make([]rune, blocks[2].length/2)
	for i := 0; i < int(blocks[2].length)/2; i++ {
		ofs := int(blocks[2].offset) + i*2
		dict.chara[i] = rune(binary.LittleEndian.Uint16(data[ofs : ofs+2]))
	}

	// 词表
	dict.word = make([]rune, blocks[3].length/2)
	for i := 0; i < int(blocks[3].length)/2; i++ {
		ofs := int(blocks[3].offset) + i*2
		dict.word[i] = rune(binary.LittleEndian.Uint16(data[ofs : ofs+2]))
	}

	// 边表
	dict.entries = make([]entryItem, blocks[4].length/16)
	for i := 0; i < int(blocks[4].length)/16; i++ {
		ofs := int(blocks[4].offset) + i*16
		dict.entries[i].flag = binary.LittleEndian.Uint32(data[ofs : ofs+4])
		dict.entries[i].sibling = binary.LittleEndian.Uint32(data[ofs+4 : ofs+8])
		dict.entries[i].child = binary.LittleEndian.Uint32(data[ofs+8 : ofs+12])
		dict.entries[i].offset = binary.LittleEndian.Uint32(data[ofs+12 : ofs+16])
	}

	return &dict, nil
}

// mapRune 对输入字符做映射
func (d *Dict) rune2Index(r rune) uint32 {
	val := int(r)
	higher := val >> 8
	lower := val & 0xFF

	if higher >= len(d.charaMap) {
		return 0
	}

	higher = int(d.charaMap[higher])
	if higher == 0 {
		return 0
	}

	return (uint32(higher) << 8) + uint32(lower)
}

// index2Rune 从index反映射到字符
func (d *Dict) index2Rune(lut map[uint16]uint16, index uint32) rune {
	higher := index >> 8
	lower := index & 0xFF

	if higher == 0 {
		return 0
	}

	newVal, ok := lut[uint16(higher)]
	if !ok {
		return 0
	}
	return rune((uint32(newVal) << 8) + lower)
}

// generateIndexRuneLut 生成反查表
func (d *Dict) generateIndexRuneLut() map[uint16]uint16 {
	val := make(map[uint16]uint16)
	for i := 0; i < len(d.charaMap); i++ {
		val[d.charaMap[i]] = uint16(i)
	}
	return val
}

// getString 获取指定节点对应的字符串
func (d *Dict) getString(entryID int, sibilID int) (string, error) {
	runes, err := d.getStringRunes(entryID, sibilID)
	return string(runes), err
}

// getStringRunes 获取指定节点对应的rune
func (d *Dict) getStringRunes(entryID int, sibilID int) ([]rune, error) {
	if entryID >= len(d.entries) {
		return make([]rune, 0), fmt.Errorf("entry ID %d overflow", entryID)
	}
	entry := d.entries[entryID]

	if entry.flag == 0 {
		pos := int(entry.offset/2) + sibilID
		if pos >= len(d.chara) {
			return make([]rune, 0), fmt.Errorf("character position %d overflow", pos)
		}
		if d.chara[pos] == 0 {
			return make([]rune, 0), nil
		}
		return d.chara[pos : pos+1], nil
	}

	begin := int(entry.offset / 2)
	end := begin + 1
	for end < len(d.word) && d.word[end] != 0 {
		end++
	}
	return d.word[begin:end], nil
}
