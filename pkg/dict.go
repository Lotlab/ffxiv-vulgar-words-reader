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
	dict := Dict{
		charaReplace: make(map[uint16]uint16),
		charaSkip:    make(map[uint16]bool),
	}

	for i := 0; i < 0x10000/8; i++ {
		ch := data[12+i]
		for j := 0; j < 8; j++ {
			if ch&(1<<j) > 0 {
				dict.charaSkip[uint16(i*8+j)] = true
			}
		}
	}

	const offsetStart = 0x8124
	offset := offsetStart

	// 字符映射表
	for i := 0; i < 256; i++ {
		val := binary.LittleEndian.Uint16(data[offset : offset+2])
		dict.charaReplace[uint16(i)+0x0000] = val
		offset += 2
	}
	for i := 0; i < 256; i++ {
		val := binary.LittleEndian.Uint16(data[offset : offset+2])
		dict.charaReplace[uint16(i)+0xFF00] = val
		offset += 2
	}
	for i := 0; i < 256; i++ {
		val := binary.LittleEndian.Uint16(data[offset : offset+2])
		dict.charaReplace[uint16(i)+0x3000] = val
		offset += 2
	}

	const mapStart = 0x8750
	const mapSize = 0x200

	// 读取偏移量, offset = 0x8724
	blocks := make([]blockInfo, 5)
	for i := 0; i < len(blocks); i++ {
		val := binary.LittleEndian.Uint32(data[offset : offset+4])
		blocks[i].offset = val + mapStart + mapSize
		offset += 4
	}
	for i := 0; i < len(blocks); i++ {
		blocks[i].length = binary.LittleEndian.Uint32(data[offset : offset+4])
		offset += 4
	}

	offset += 4

	// 区块映射表
	dict.charaBlock = make([]uint16, 256)
	for i := 0; i < mapSize/2; i++ {
		dict.charaBlock[i] = binary.LittleEndian.Uint16(data[offset : offset+2])
		offset += 2
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

// runeMapping 映射字符表
func (d *Dict) runeMapping(r rune) rune {
	ret, ok := d.charaReplace[uint16(r)]
	if ok {
		return rune(ret)
	}
	return r
}

// rune2Index 将rune转换为index
func (d *Dict) rune2Index(r rune) uint32 {
	val := uint16(d.runeMapping(r))

	higher := val >> 8
	lower := val & 0xFF

	if int(higher) >= len(d.charaBlock) {
		return 0
	}

	higher = d.charaBlock[higher]
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
	for i := 0; i < len(d.charaBlock); i++ {
		val[d.charaBlock[i]] = uint16(i)
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

// isSkip 判断是否要跳过指定字符
func (d *Dict) isSkip(r rune) bool {
	_, ok := d.charaSkip[uint16(r)]
	return ok
}
