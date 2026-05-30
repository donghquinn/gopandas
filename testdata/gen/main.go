// Command gen creates binary test fixtures (xlsx, xls) in the testdata directory.
// Run from the project root: go run testdata/gen/main.go
package main

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
)

// ── XLSX ────────────────────────────────────────────────────────────────────

func createXLSX(filename, sheetXML string) error {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	fw, err := w.Create("xl/worksheets/sheet1.xml")
	if err != nil {
		return fmt.Errorf("create sheet entry: %w", err)
	}
	if _, err := fw.Write([]byte(sheetXML)); err != nil {
		return fmt.Errorf("write sheet xml: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("close zip: %w", err)
	}
	return os.WriteFile(filename, buf.Bytes(), 0644)
}

// basicSheetXML: 5 data rows, columns name/age/city.
// Strings use inlineStr so no sharedStrings.xml is needed.
const basicSheetXML = `<?xml version="1.0" encoding="UTF-8"?>
<worksheet>
  <sheetData>
    <row><c t="inlineStr"><is><t>name</t></is></c><c t="inlineStr"><is><t>age</t></is></c><c t="inlineStr"><is><t>city</t></is></c></row>
    <row><c t="inlineStr"><is><t>Alice</t></is></c><c><v>25</v></c><c t="inlineStr"><is><t>New York</t></is></c></row>
    <row><c t="inlineStr"><is><t>Bob</t></is></c><c><v>30</v></c><c t="inlineStr"><is><t>London</t></is></c></row>
    <row><c t="inlineStr"><is><t>Charlie</t></is></c><c><v>35</v></c><c t="inlineStr"><is><t>Paris</t></is></c></row>
    <row><c t="inlineStr"><is><t>Dave</t></is></c><c><v>28</v></c><c t="inlineStr"><is><t>Tokyo</t></is></c></row>
    <row><c t="inlineStr"><is><t>Eve</t></is></c><c><v>22</v></c><c t="inlineStr"><is><t>Berlin</t></is></c></row>
  </sheetData>
</worksheet>`

// nullsSheetXML: 4 data rows, columns name/score/grade; some cells are empty.
// An empty <c/> produces Value="" → inferType("") = nil in the parser.
const nullsSheetXML = `<?xml version="1.0" encoding="UTF-8"?>
<worksheet>
  <sheetData>
    <row><c t="inlineStr"><is><t>name</t></is></c><c t="inlineStr"><is><t>score</t></is></c><c t="inlineStr"><is><t>grade</t></is></c></row>
    <row><c t="inlineStr"><is><t>Alice</t></is></c><c/><c t="inlineStr"><is><t>A</t></is></c></row>
    <row><c t="inlineStr"><is><t>Bob</t></is></c><c><v>85</v></c><c/></row>
    <row><c t="inlineStr"><is><t>Charlie</t></is></c><c><v>90</v></c><c t="inlineStr"><is><t>B</t></is></c></row>
    <row><c/><c><v>70</v></c><c t="inlineStr"><is><t>C</t></is></c></row>
  </sheetData>
</worksheet>`

// ── XLS (custom BIFF8-style matching the project's XLS parser) ─────────────
//
// The project's parseXLS reads records in the form: type(uint16) + size(uint16) + data.
//   - Record 0x00FC (SST): data = length(uint16) + ASCII string bytes.
//     parseSST adds one string per record to the shared strings slice.
//   - Record 0x0201 (Row): data = rowIdx(uint16) + firstCol(uint16) + lastCol(uint16)
//     followed by (lastCol-firstCol+1) cells, each 8 bytes:
//       cellType(uint16) + cellData[6]byte
//     Cell type 0x0205: cellData[0:4] = uint32 index into shared strings.
//     Any other type: printable ASCII bytes from cellData become the cell string.
//
// The first row becomes column headers; subsequent rows become data rows.
// Values are passed through inferType (empty→nil, numeric strings→int/float64, etc.).

func writeU16(b *bytes.Buffer, v uint16) {
	_ = binary.Write(b, binary.LittleEndian, v)
}

func writeRecord(b *bytes.Buffer, typ uint16, data []byte) {
	writeU16(b, typ)
	writeU16(b, uint16(len(data)))
	b.Write(data)
}

// writeSST appends one SST record (one shared string).
func writeSST(b *bytes.Buffer, s string) {
	var d bytes.Buffer
	writeU16(&d, uint16(len(s)))
	d.WriteString(s)
	writeRecord(b, 0x00FC, d.Bytes())
}

type xlsCell struct {
	strIdx int    // ≥0: shared string reference; <0: use ascii
	ascii  string // up to 6 printable ASCII bytes
}

func strCell(idx int) xlsCell    { return xlsCell{strIdx: idx} }
func asciiCell(s string) xlsCell { return xlsCell{strIdx: -1, ascii: s} }
func emptyCell() xlsCell         { return xlsCell{strIdx: -1, ascii: ""} }

func writeRow(b *bytes.Buffer, rowIdx uint16, cells []xlsCell) {
	nCols := uint16(len(cells))
	var d bytes.Buffer
	writeU16(&d, rowIdx)
	writeU16(&d, 0)       // firstCol
	writeU16(&d, nCols-1) // lastCol
	for _, c := range cells {
		if c.strIdx >= 0 {
			writeU16(&d, 0x0205)
			var data [6]byte
			binary.LittleEndian.PutUint32(data[0:4], uint32(c.strIdx))
			d.Write(data[:])
		} else {
			writeU16(&d, 0x0000) // default → printable ASCII
			var data [6]byte
			for i, ch := range []byte(c.ascii) {
				if i < 6 {
					data[i] = ch
				}
			}
			d.Write(data[:])
		}
	}
	writeRecord(b, 0x0201, d.Bytes())
}

func createXLS(filename string, sharedStrings []string, rows [][]xlsCell) error {
	var buf bytes.Buffer
	// BOF record with BIFF8 signature (0x0809).
	writeRecord(&buf, 0x0809, []byte{0x00, 0x06, 0x10, 0x00, 0xE4, 0x00, 0xCC, 0x07})
	for _, s := range sharedStrings {
		writeSST(&buf, s)
	}
	for i, row := range rows {
		writeRow(&buf, uint16(i), row)
	}
	return os.WriteFile(filename, buf.Bytes(), 0644)
}

// ── main ────────────────────────────────────────────────────────────────────

func must(label string, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "gen %s: %v\n", label, err)
		os.Exit(1)
	}
}

func main() {
	// basic.xlsx — 5 rows, columns: name / age / city
	must("basic.xlsx", createXLSX("testdata/basic.xlsx", basicSheetXML))

	// nulls.xlsx — 4 rows with some empty cells, columns: name / score / grade
	must("nulls.xlsx", createXLSX("testdata/nulls.xlsx", nullsSheetXML))

	// basic.xls — same logical content as basic.xlsx
	//
	// Shared strings table (index → value):
	//   0=name  1=age  2=city
	//   3=Alice  4=New York
	//   5=Bob    6=London
	//   7=Charlie 8=Paris
	//   9=Dave  10=Tokyo
	//  11=Eve  12=Berlin
	xlsStrings := []string{
		"name", "age", "city",
		"Alice", "New York",
		"Bob", "London",
		"Charlie", "Paris",
		"Dave", "Tokyo",
		"Eve", "Berlin",
	}
	xlsRows := [][]xlsCell{
		{strCell(0), strCell(1), strCell(2)}, // header
		{strCell(3), asciiCell("25"), strCell(4)},
		{strCell(5), asciiCell("30"), strCell(6)},
		{strCell(7), asciiCell("35"), strCell(8)},
		{strCell(9), asciiCell("28"), strCell(10)},
		{strCell(11), asciiCell("22"), strCell(12)},
	}
	must("basic.xls", createXLS("testdata/basic.xls", xlsStrings, xlsRows))

	fmt.Println("generated: testdata/basic.xlsx  testdata/nulls.xlsx  testdata/basic.xls")
}
