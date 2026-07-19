package gopandas

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type ExcelReader struct {
	zipReader *zip.ReadCloser
	strings   []string
}

type xlsxCell struct {
	Reference string `xml:"r,attr"`
	Type      string `xml:"t,attr"`
	Value     string `xml:"v"`
	InlineStr struct {
		Text string `xml:"t"`
	} `xml:"is"`
}

type xlsxRow struct {
	Cells []xlsxCell `xml:"c"`
}

type worksheet struct {
	SheetData struct {
		Rows []xlsxRow `xml:"row"`
	} `xml:"sheetData"`
}

type sharedStrings struct {
	Items []struct {
		Text string `xml:"t"`
	} `xml:"si"`
}

func ReadExcel(filename string, sheetName ...string) (*DataFrame, error) {
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".xlsx":
		return readXLSX(filename, sheetName...)
	case ".xls":
		return readXLS(filename, sheetName...)
	default:
		return nil, fmt.Errorf("unsupported file format: %s (only .xlsx and .xls files are supported)", ext)
	}
}

func readXLSX(filename string, sheetName ...string) (*DataFrame, error) {
	reader, err := zip.OpenReader(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer reader.Close()

	excelReader := &ExcelReader{
		zipReader: reader,
	}

	if err := excelReader.loadSharedStrings(); err != nil {
		return nil, fmt.Errorf("failed to load shared strings: %w", err)
	}

	sheet := "sheet1.xml"
	if len(sheetName) > 0 && sheetName[0] != "" {
		sheet = strings.ToLower(sheetName[0]) + ".xml"
	}

	return excelReader.readWorksheet(sheet)
}

func readXLS(filename string, sheetName ...string) (*DataFrame, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read XLS file: %w", err)
	}

	return parseXLS(data, sheetName...)
}

func (er *ExcelReader) loadSharedStrings() error {
	for _, file := range er.zipReader.File {
		if file.Name == "xl/sharedStrings.xml" {
			rc, err := file.Open()
			if err != nil {
				return err
			}
			defer rc.Close()

			data, err := io.ReadAll(rc)
			if err != nil {
				return err
			}

			var ss sharedStrings
			if err := xml.Unmarshal(data, &ss); err != nil {
				return err
			}

			er.strings = make([]string, len(ss.Items))
			for i, item := range ss.Items {
				er.strings[i] = item.Text
			}

			return nil
		}
	}
	return nil
}

func (er *ExcelReader) readWorksheet(sheetName string) (*DataFrame, error) {
	var worksheetFile *zip.File

	for _, file := range er.zipReader.File {
		if strings.HasSuffix(file.Name, sheetName) || file.Name == "xl/worksheets/"+sheetName {
			worksheetFile = file
			break
		}
	}

	if worksheetFile == nil {
		return nil, fmt.Errorf("worksheet '%s' not found", sheetName)
	}

	rc, err := worksheetFile.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, err
	}

	var ws worksheet
	if err := xml.Unmarshal(data, &ws); err != nil {
		return nil, err
	}

	if len(ws.SheetData.Rows) == 0 {
		return nil, fmt.Errorf("worksheet is empty")
	}

	rows := make([][]string, len(ws.SheetData.Rows))
	for i, wsRow := range ws.SheetData.Rows {
		row := make([]string, len(wsRow.Cells))
		for j, cell := range wsRow.Cells {
			row[j] = er.getCellValue(cell)
		}
		rows[i] = row
	}

	return dataFrameFromStringRows(rows), nil
}

func (er *ExcelReader) getCellValue(cell xlsxCell) string {
	if cell.Type == "s" {
		if idx, err := strconv.Atoi(cell.Value); err == nil && idx >= 0 && idx < len(er.strings) {
			return er.strings[idx]
		}
	} else if cell.Type == "inlineStr" {
		return cell.InlineStr.Text
	}

	return cell.Value
}

// dataFrameFromStringRows builds a DataFrame from raw string rows. The first
// row provides column names; ragged rows are padded with generated names/nils.
func dataFrameFromStringRows(rows [][]string) *DataFrame {
	maxCols := 0
	for _, row := range rows {
		if len(row) > maxCols {
			maxCols = len(row)
		}
	}

	columns := make([]string, maxCols)
	if len(rows) > 0 {
		copy(columns, rows[0])
		for i := len(rows[0]); i < maxCols; i++ {
			columns[i] = fmt.Sprintf("col_%d", i)
		}
	}

	df := NewDataFrame(columns)
	if len(rows) > 1 {
		df.data = make([][]interface{}, 0, len(rows)-1)
		df.index = make([]interface{}, 0, len(rows)-1)
	}

	for i := 1; i < len(rows); i++ {
		row := make([]interface{}, maxCols)
		for j, val := range rows[i] {
			row[j] = inferType(val)
		}
		df.AddRow(row)
	}

	return df
}

type xlsRecord struct {
	Type uint16
	Size uint16
	Data []byte
}

func parseXLS(data []byte, sheetName ...string) (*DataFrame, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("invalid XLS file: too small")
	}

	reader := bytes.NewReader(data)

	// Check for various XLS signatures
	var signature uint16
	if err := binary.Read(reader, binary.LittleEndian, &signature); err != nil {
		return nil, fmt.Errorf("failed to read XLS signature: %w", err)
	}

	// Valid XLS signatures: BIFF5 (0x0805), BIFF8 (0x0809), or OLE compound document (0xD0CF)
	validSignature := false
	switch signature {
	case 0x0809: // BIFF8
		validSignature = true
	case 0x0805: // BIFF5
		validSignature = true
	case 0xD0CF: // OLE compound document (little endian)
		validSignature = true
		// For OLE files, we need to find the actual workbook stream
		return parseOLEXLS(data, sheetName...)
	case 0xCFD0: // OLE compound document (big endian read)
		validSignature = true
		// For OLE files, we need to find the actual workbook stream
		return parseOLEXLS(data, sheetName...)
	}

	if !validSignature {
		return nil, fmt.Errorf("invalid XLS file: unsupported signature 0x%04X", signature)
	}

	reader.Seek(0, 0)

	var strings []string
	var rows [][]string

	for reader.Len() > 4 {
		var record xlsRecord
		if err := binary.Read(reader, binary.LittleEndian, &record.Type); err != nil {
			break
		}
		if err := binary.Read(reader, binary.LittleEndian, &record.Size); err != nil {
			break
		}

		if record.Size > 0 {
			record.Data = make([]byte, record.Size)
			if n, err := reader.Read(record.Data); err != nil || n != int(record.Size) {
				break
			}
		}

		switch record.Type {
		case 0x00FC:
			if str := parseSST(record.Data); str != "" {
				strings = append(strings, str)
			}
		case 0x0201:
			if row := parseRow(record.Data, strings); len(row) > 0 {
				rows = append(rows, row)
			}
		}
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("no data found in XLS file")
	}

	return dataFrameFromStringRows(rows), nil
}

func parseSST(data []byte) string {
	if len(data) < 2 {
		return ""
	}

	reader := bytes.NewReader(data)
	var length uint16
	binary.Read(reader, binary.LittleEndian, &length)

	if int(length) > reader.Len() {
		return ""
	}

	strData := make([]byte, length)
	reader.Read(strData)

	return string(strData)
}

func parseRow(data []byte, strings []string) []string {
	if len(data) < 6 {
		return nil
	}

	reader := bytes.NewReader(data)
	var rowIndex, firstCol, lastCol uint16

	if err := binary.Read(reader, binary.LittleEndian, &rowIndex); err != nil {
		return nil
	}
	if err := binary.Read(reader, binary.LittleEndian, &firstCol); err != nil {
		return nil
	}
	if err := binary.Read(reader, binary.LittleEndian, &lastCol); err != nil {
		return nil
	}

	if lastCol < firstCol || lastCol-firstCol > 1000 { // sanity check
		return nil
	}

	row := make([]string, lastCol-firstCol+1)

	for i := range row {
		if reader.Len() >= 8 {
			var cellType uint16
			var cellData [6]byte

			if err := binary.Read(reader, binary.LittleEndian, &cellType); err != nil {
				break
			}
			if n, err := reader.Read(cellData[:]); err != nil || n != 6 {
				break
			}

			switch cellType {
			case 0x0204:
				if len(cellData) >= 8 {
					val := binary.LittleEndian.Uint64(cellData[:])
					row[i] = fmt.Sprintf("%.2f", float64(val))
				}
			case 0x0205:
				if len(cellData) >= 4 {
					idx := binary.LittleEndian.Uint32(cellData[:4])
					if int(idx) < len(strings) && strings != nil {
						row[i] = strings[idx]
					}
				}
			default:
				// Clean the string data
				cleaned := make([]byte, 0, len(cellData))
				for _, b := range cellData {
					if b != 0 && b >= 32 && b < 127 { // printable ASCII
						cleaned = append(cleaned, b)
					}
				}
				row[i] = string(cleaned)
			}
		}
	}

	return row
}

func parseOLEXLS(data []byte, sheetName ...string) (*DataFrame, error) {
	if len(data) < 512 {
		return nil, fmt.Errorf("invalid OLE file: too small")
	}

	// Simple OLE parsing - look for workbook stream data
	// Most XLS files store the actual Excel data after the OLE header

	// Try to find BIFF records starting from different offsets
	offsets := []int{512, 1024, 2048, 4096}

	for _, offset := range offsets {
		if offset >= len(data) {
			continue
		}

		// Check if we can find a BIFF signature at this offset
		if offset+4 < len(data) {
			sig := binary.LittleEndian.Uint16(data[offset:])
			if sig == 0x0809 || sig == 0x0805 {
				// Found BIFF data, parse from this offset
				return parseBIFFData(data[offset:], sheetName...)
			}
		}
	}

	// If no BIFF data found, try a more aggressive search
	for i := 0; i < len(data)-4; i += 512 {
		if i+4 < len(data) {
			sig := binary.LittleEndian.Uint16(data[i:])
			if sig == 0x0809 || sig == 0x0805 {
				return parseBIFFData(data[i:], sheetName...)
			}
		}
	}

	return nil, fmt.Errorf("no valid Excel data found in OLE file")
}

func parseBIFFData(data []byte, sheetName ...string) (*DataFrame, error) {
	reader := bytes.NewReader(data)

	var strings []string
	var rows [][]string

	for reader.Len() > 4 {
		var record xlsRecord
		if err := binary.Read(reader, binary.LittleEndian, &record.Type); err != nil {
			break
		}
		if err := binary.Read(reader, binary.LittleEndian, &record.Size); err != nil {
			break
		}

		if record.Size > 0 && int(record.Size) <= reader.Len() {
			record.Data = make([]byte, record.Size)
			if n, err := reader.Read(record.Data); err != nil || n != int(record.Size) {
				break
			}
		}

		switch record.Type {
		case 0x00FC: // SST
			if str := parseSST(record.Data); str != "" {
				strings = append(strings, str)
			}
		case 0x0201: // BLANK
			if row := parseRow(record.Data, strings); len(row) > 0 {
				rows = append(rows, row)
			}
		case 0x0203: // NUMBER
			if row := parseNumberRecord(record.Data); len(row) > 0 {
				rows = append(rows, row)
			}
		case 0x0204: // LABEL
			if row := parseLabelRecord(record.Data, strings); len(row) > 0 {
				rows = append(rows, row)
			}
		}
	}

	return dataFrameFromStringRows(rows), nil
}

func parseNumberRecord(data []byte) []string {
	if len(data) < 14 {
		return nil
	}

	reader := bytes.NewReader(data)
	var row, col uint16
	var value float64

	if err := binary.Read(reader, binary.LittleEndian, &row); err != nil {
		return nil
	}
	if err := binary.Read(reader, binary.LittleEndian, &col); err != nil {
		return nil
	}
	if col > 255 { // sanity check
		return nil
	}
	reader.Seek(4, 1) // skip XF index
	if err := binary.Read(reader, binary.LittleEndian, &value); err != nil {
		return nil
	}

	result := make([]string, int(col)+1)
	result[col] = fmt.Sprintf("%.2f", value)

	return result
}

func parseLabelRecord(data []byte, strings []string) []string {
	if len(data) < 8 {
		return nil
	}

	reader := bytes.NewReader(data)
	var row, col, length uint16

	if err := binary.Read(reader, binary.LittleEndian, &row); err != nil {
		return nil
	}
	if err := binary.Read(reader, binary.LittleEndian, &col); err != nil {
		return nil
	}
	if col > 255 { // sanity check
		return nil
	}
	reader.Seek(2, 1) // skip XF index
	if err := binary.Read(reader, binary.LittleEndian, &length); err != nil {
		return nil
	}

	if int(length) > reader.Len() || length > 1000 { // sanity check
		return nil
	}

	strData := make([]byte, length)
	if n, err := reader.Read(strData); err != nil || n != int(length) {
		return nil
	}

	// Clean the string data
	cleaned := make([]byte, 0, len(strData))
	for _, b := range strData {
		if b != 0 && (b >= 32 || b == 9 || b == 10 || b == 13) { // printable chars + tab/newline
			cleaned = append(cleaned, b)
		}
	}

	result := make([]string, int(col)+1)
	result[col] = string(cleaned)

	return result
}
