package pkg

// CheckRunes 判断文本中是否含有敏感词
// 返回值中，begin和end为敏感词出现的起始和结束位置。通过 runes[begin:end] 即可获取到实际的敏感词
func (d *Dict) CheckRunes(runes []rune) (begin int, end int, err error) {
	for i, r := range runes {
		begin = i
		mapIndex := d.rune2Index(r)
		if mapIndex == 0 || d.isSkip(r) || int(mapIndex) >= len(d.beginNode) {
			continue
		}

		nodeIndex := d.beginNode[mapIndex]
		result, e := d.nodeComp(runes[i+1:], int(nodeIndex))
		if e != nil {
			end = -1
			err = e
			return
		}
		if result >= 0 {
			end = begin + result + 1
			return
		}
	}

	begin = -1
	end = -1
	err = nil
	return
}

// CheckString 判断文本中是否含有敏感词
func (d *Dict) CheckString(str string) (begin int, end int, err error) {
	runes := []rune(str)
	return d.CheckRunes(runes)
}

func (d *Dict) nodeComp(str []rune, entryID int) (int, error) {
	node := d.entries[entryID]
	for i := 0; i < int(node.sibling); i++ {
		// 获得当前节点的值
		current, err := d.getStringRunes(entryID, i)
		if err != nil {
			return -1, err
		}

		// 判断节点与目标文本是否完全匹配
		if len(current) > len(str) {
			continue
		}
		match := true
		strOffset := 0
		for i := 0; i < len(current); i++ {
			for d.isSkip(str[strOffset]) {
				strOffset++
			}

			if current[i] != d.runeMapping(str[strOffset]) {
				match = false
				break
			}
			strOffset++
		}
		if !match {
			continue
		}

		// 完全匹配，没有子节点
		if node.child == 0 {
			return strOffset, nil
		}

		val := d.innerNode[int(node.child)+i]
		if val == 0 {
			// 最后一个节点，完全匹配
			return strOffset, nil
		}

		// 匹配子节点
		result, err := d.nodeComp(str[strOffset:], int(val))
		if err != nil {
			return -1, err
		}
		if result >= 0 {
			return strOffset + result, nil
		}
	}
	return -1, nil
}
