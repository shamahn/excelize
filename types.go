package excelize

import (
	"encoding/xml"
	"sort"
)

type ERows map[int]xlsxRow

func (m ERows) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if len(m) == 0 {
		return nil
	}

	var keys []int
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	out := []interface{}{}
	for _, k := range keys {
		out = append(out, m[k])
	}

	return e.EncodeElement(out, start)
}

type ECols map[int]xlsxC

func (m ECols) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if len(m) == 0 {
		return nil
	}

	var keys []int
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	out := []interface{}{}
	for _, k := range keys {
		out = append(out, m[k])
	}

	return e.EncodeElement(out, start)
}
