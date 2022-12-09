package pkg

// DumpDict 将内部文本导出
func (d *Dict) DumpDict() ([]string, error) {
	result := make([]string, 0)
	lut := d.generateIndexRuneLut()

	for id, v := range d.beginNode {
		if v == 0 {
			continue
		}
		chara := d.index2Rune(lut, uint32(id))
		err := d.dumpDictNode(int(v), &result, string(chara))
		if err != nil {
			return result, err
		}
	}
	return result, nil
}

func (d *Dict) dumpDictNode(entryID int, vec *[]string, prev string) error {
	node := d.entries[entryID]
	for i := 0; i < int(node.sibling); i++ {
		current, err := d.getString(entryID, i)
		if err != nil {
			return err
		}

		if node.child == 0 {
			*vec = append(*vec, prev+current)
			continue
		}

		val := d.innerNode[int(node.child)+i]
		if val == 0 {
			*vec = append(*vec, prev+current)
			continue
		}

		d.dumpDictNode(int(val), vec, prev+current)
	}
	return nil
}
