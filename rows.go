package excelize

import (
	"bytes"
	"encoding/xml"
	"math"
	"strconv"
	"strings"
)

// GetRows return all the rows in a sheet by given worksheet name (case
// sensitive). For example:
//
//    for _, row := range xlsx.GetRows("Sheet1") {
//        for _, colCell := range row {
//            fmt.Print(colCell, "\t")
//        }
//        fmt.Println()
//    }
//
func (f *File) GetRows(sheet string) [][]string {
	xlsx := f.workSheetReader(sheet)
	rows := [][]string{}
	name, ok := f.sheetMap[trimSheetName(sheet)]
	if !ok {
		return rows
	}
	if xlsx != nil {
		output, _ := xml.Marshal(f.Sheet[name])
		f.saveFileList(name, replaceWorkSheetsRelationshipsNameSpace(string(output)))
	}
	decoder := xml.NewDecoder(strings.NewReader(f.readXML(name)))
	d := f.sharedStringsReader()
	var inElement string
	var r xlsxRow
	var row []string
	tr, tc := f.getTotalRowsCols(name)
	for i := 0; i < tr; i++ {
		row = []string{}
		for j := 0; j <= tc; j++ {
			row = append(row, "")
		}
		rows = append(rows, row)
	}
	decoder = xml.NewDecoder(strings.NewReader(f.readXML(name)))
	for {
		token, _ := decoder.Token()
		if token == nil {
			break
		}
		switch startElement := token.(type) {
		case xml.StartElement:
			inElement = startElement.Name.Local
			if inElement == "row" {
				r = xlsxRow{}
				decoder.DecodeElement(&r, &startElement)
				cr := r.R - 1
				for _, colCell := range r.C {
					c := TitleToNumber(strings.Map(letterOnlyMapF, colCell.R))
					val, _ := colCell.getValueFrom(f, d)
					rows[cr][c] = val
				}
			}
		default:
		}
	}
	return rows
}

// getTotalRowsCols provides a function to get total columns and rows in a
// worksheet.
func (f *File) getTotalRowsCols(name string) (int, int) {
	decoder := xml.NewDecoder(strings.NewReader(f.readXML(name)))
	var inElement string
	var r xlsxRow
	var tr, tc int
	for {
		token, _ := decoder.Token()
		if token == nil {
			break
		}
		switch startElement := token.(type) {
		case xml.StartElement:
			inElement = startElement.Name.Local
			if inElement == "row" {
				r = xlsxRow{}
				decoder.DecodeElement(&r, &startElement)
				tr = r.R
				for _, colCell := range r.C {
					col := TitleToNumber(strings.Map(letterOnlyMapF, colCell.R))
					if col > tc {
						tc = col
					}
				}
			}
		default:
		}
	}
	return tr, tc
}

// SetRowHeight provides a function to set the height of a single row.
// For example:
//
//    xlsx := excelize.NewFile()
//    xlsx.SetRowHeight("Sheet1", 0, 50)
//    err := xlsx.Save()
//    if err != nil {
//        fmt.Println(err)
//    }
//
func (f *File) SetRowHeight(sheet string, rowIndex int, height float64) {
	xlsx := f.workSheetReader(sheet)
	rows := rowIndex + 1
	cells := 0
	completeRow(xlsx, rows, cells)
	tRow := xlsx.SheetData.Row[rowIndex]
	tRow.Ht = height
	tRow.CustomHeight = true

	xlsx.SheetData.Row[rowIndex] = tRow
}

// getRowHeight provides function to get row height in pixels by given sheet
// name and row index.
func (f *File) getRowHeight(sheet string, row int) int {
	xlsx := f.workSheetReader(sheet)
	for _, v := range xlsx.SheetData.Row {
		if v.R == row+1 && v.Ht != 0 {
			return int(convertRowHeightToPixels(v.Ht))
		}
	}
	// Optimisation for when the row heights haven't changed.
	return int(defaultRowHeightPixels)
}

// GetRowHeight provides function to get row height by given worksheet name and
// row index.
func (f *File) GetRowHeight(sheet string, row int) float64 {
	xlsx := f.workSheetReader(sheet)
	for _, v := range xlsx.SheetData.Row {
		if v.R == row+1 && v.Ht != 0 {
			return v.Ht
		}
	}
	// Optimisation for when the row heights haven't changed.
	return defaultRowHeightPixels
}

// sharedStringsReader provides function to get the pointer to the structure
// after deserialization of xl/sharedStrings.xml.
func (f *File) sharedStringsReader() *xlsxSST {
	if f.SharedStrings == nil {
		var sharedStrings xlsxSST
		xml.Unmarshal([]byte(f.readXML("xl/sharedStrings.xml")), &sharedStrings)
		f.SharedStrings = &sharedStrings
	}
	return f.SharedStrings
}

// getValueFrom return a value from a column/row cell, this function is inteded
// to be used with for range on rows an argument with the xlsx opened file.
func (xlsx *xlsxC) getValueFrom(f *File, d *xlsxSST) (string, error) {
	switch xlsx.T {
	case "s":
		xlsxSI := 0
		xlsxSI, _ = strconv.Atoi(xlsx.V)
		if len(d.SI[xlsxSI].R) > 0 {
			value := ""
			for _, v := range d.SI[xlsxSI].R {
				value += v.T
			}
			return value, nil
		}
		return f.formattedValue(xlsx.S, d.SI[xlsxSI].T), nil
	case "str":
		return f.formattedValue(xlsx.S, xlsx.V), nil
	default:
		return f.formattedValue(xlsx.S, xlsx.V), nil
	}
}

// SetRowVisible provides a function to set visible of a single row by given
// worksheet name and row index. For example, hide row 3 in Sheet1:
//
//    xlsx.SetRowVisible("Sheet1", 2, false)
//
func (f *File) SetRowVisible(sheet string, rowIndex int, visible bool) {
	xlsx := f.workSheetReader(sheet)
	rows := rowIndex + 1
	cells := 0
	completeRow(xlsx, rows, cells)
	tRow := xlsx.SheetData.Row[rowIndex]
	if visible {
		tRow.Hidden = false
		return
	}
	tRow.Hidden = true
	xlsx.SheetData.Row[rowIndex] = tRow
}

// GetRowVisible provides a function to get visible of a single row by given
// worksheet name and row index. For example, get visible state of row 3 in
// Sheet1:
//
//    xlsx.GetRowVisible("Sheet1", 2)
//
func (f *File) GetRowVisible(sheet string, rowIndex int) bool {
	xlsx := f.workSheetReader(sheet)
	rows := rowIndex + 1
	cells := 0
	completeRow(xlsx, rows, cells)
	return !xlsx.SheetData.Row[rowIndex].Hidden
}

// RemoveRow provides function to remove single row by given worksheet name and
// row index. For example, remove row 3 in Sheet1:
//
//    xlsx.RemoveRow("Sheet1", 2)
//
func (f *File) RemoveRow(sheet string, row int) {
	if row < 0 {
		return
	}
	xlsx := f.workSheetReader(sheet)
	row++
	for i, r := range xlsx.SheetData.Row {
		if r.R == row {
			delete(xlsx.SheetData.Row, i)
			f.adjustHelper(sheet, -1, row, -1)
			return
		}
	}
}

// InsertRow provides function to insert a new row before given row index. For
// example, create a new row before row 3 in Sheet1:
//
//    xlsx.InsertRow("Sheet1", 2)
//
func (f *File) InsertRow(sheet string, row int) {
	if row < 0 {
		return
	}
	row++
	f.adjustHelper(sheet, -1, row, 1)
}

// checkRow provides function to check and fill each column element for all rows
// and make that is continuous in a worksheet of XML. For example:
//
//    <row r="15" spans="1:22" x14ac:dyDescent="0.2">
//        <c r="A15" s="2" />
//        <c r="B15" s="2" />
//        <c r="F15" s="1" />
//        <c r="G15" s="1" />
//    </row>
//
// in this case, we should to change it to
//
//    <row r="15" spans="1:22" x14ac:dyDescent="0.2">
//        <c r="A15" s="2" />
//        <c r="B15" s="2" />
//        <c r="C15" s="2" />
//        <c r="D15" s="2" />
//        <c r="E15" s="2" />
//        <c r="F15" s="1" />
//        <c r="G15" s="1" />
//    </row>
//
// Noteice: this method could be very slow for large spreadsheets (more than
// 3000 rows one sheet).
func checkRow(xlsx *xlsxWorksheet) {
	buffer := bytes.Buffer{}
	for k := range xlsx.SheetData.Row {
		lenCol := len(xlsx.SheetData.Row[k].C)
		if lenCol > 0 {
			endR := string(strings.Map(letterOnlyMapF, xlsx.SheetData.Row[k].C[lenCol-1].R))
			endRow, _ := strconv.Atoi(strings.Map(intOnlyMapF, xlsx.SheetData.Row[k].C[lenCol-1].R))
			endCol := TitleToNumber(endR) + 1
			if lenCol < endCol {
				oldRow := xlsx.SheetData.Row[k].C
				for c := range xlsx.SheetData.Row[k].C {
					delete(xlsx.SheetData.Row[k].C, c)
				}
				tmp := make(ECols)
				for i := 0; i <= endCol; i++ {
					buffer.WriteString(ToAlphaString(i))
					buffer.WriteString(strconv.Itoa(endRow))
					tmp[i] = xlsxC{
						R: buffer.String(),
					}
					buffer.Reset()
				}
				tRow := xlsx.SheetData.Row[k]
				tRow.C = tmp
				xlsx.SheetData.Row[k] = tRow

				for _, y := range oldRow {
					colAxis := TitleToNumber(string(strings.Map(letterOnlyMapF, y.R)))
					xlsx.SheetData.Row[k].C[colAxis] = y
				}
			}
		}
	}
}

// completeRow provides function to check and fill each column element for a
// single row and make that is continuous in a worksheet of XML by given row
// index and axis.
func completeRow(xlsx *xlsxWorksheet, row, cell int) {
	if xlsx.SheetData.Row == nil {
		xlsx.SheetData.Row = make(ERows)
	}

	i := row - 1
	{
		if xlsx.SheetData.Row[i].R > 0 {
			return
		}
		xlsx.SheetData.Row[i] = xlsxRow{
			R: i + 1,
		}
	}

	ii := row - 1
	{
		if xlsx.SheetData.Row[ii].C == nil {
			tRow := xlsx.SheetData.Row[ii]
			tRow.C = make(ECols)
			xlsx.SheetData.Row[ii] = tRow
		}
		start := len(xlsx.SheetData.Row[ii].C)
		if start == 0 {
			iii := cell - 1
			{
				if xlsx.SheetData.Row[ii].C[iii].R == "" {
					buffer := bytes.Buffer{}
					buffer.WriteString(ToAlphaString(iii))
					buffer.WriteString(strconv.Itoa(ii + 1))
					xlsx.SheetData.Row[ii].C[iii] = xlsxC{
						R: buffer.String(),
					}
					buffer.Reset()
				}
			}
		}
	}
}

// convertRowHeightToPixels provides function to convert the height of a cell
// from user's units to pixels. If the height hasn't been set by the user we use
// the default value. If the row is hidden it has a value of zero.
func convertRowHeightToPixels(height float64) float64 {
	var pixels float64
	if height == 0 {
		return pixels
	}
	pixels = math.Ceil(4.0 / 3.0 * height)
	return pixels
}
