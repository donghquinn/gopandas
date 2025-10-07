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
	strings   map[int]string
}

type worksheet struct {
	SheetData struct {
		Rows []struct {
			Cells []struct {
				Reference string `xml:"r,attr"`
				Type      string `xml:"t,attr"`
				Value     string `xml:"v"`
				InlineStr struct {
					Text string `xml:"t"`
				} `xml:"is"`
			} `xml:"c"`
		} `xml:"row"`
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
		strings:   make(map[int]string),
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
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open XLS file: %w", err)
	}
	defer file.Close()
	
	data, err := io.ReadAll(file)
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
	
	maxCols := 0
	for _, row := range ws.SheetData.Rows {
		if len(row.Cells) > maxCols {
			maxCols = len(row.Cells)
		}
	}
	
	columns := make([]string, maxCols)
	if len(ws.SheetData.Rows) > 0 {
		firstRow := ws.SheetData.Rows[0]
		for i, cell := range firstRow.Cells {
			if i < maxCols {
				columns[i] = er.getCellValue(cell)
			}
		}
		for i := len(firstRow.Cells); i < maxCols; i++ {
			columns[i] = fmt.Sprintf("col_%d", i)
		}
	} else {
		for i := range columns {
			columns[i] = fmt.Sprintf("col_%d", i)
		}
	}
	
	df := NewDataFrame(columns)
	
	for i := 1; i < len(ws.SheetData.Rows); i++ {
		row := make([]interface{}, maxCols)
		cells := ws.SheetData.Rows[i].Cells
		
		for j := 0; j < maxCols; j++ {
			if j < len(cells) {
				value := er.getCellValue(cells[j])
				row[j] = inferType(value)
			} else {
				row[j] = nil
			}
		}
		
		df.AddRow(row)
	}
	
	return df, nil
}

func (er *ExcelReader) getCellValue(cell struct {
	Reference string `xml:"r,attr"`
	Type      string `xml:"t,attr"`
	Value     string `xml:"v"`
	InlineStr struct {
		Text string `xml:"t"`
	} `xml:"is"`
}) string {
	if cell.Type == "s" {
		if idx, err := strconv.Atoi(cell.Value); err == nil {
			if str, exists := er.strings[idx]; exists {
				return str
			}
		}
	} else if cell.Type == "inlineStr" {
		return cell.InlineStr.Text
	}
	
	return cell.Value
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
	
	var signature uint16
	if err := binary.Read(reader, binary.LittleEndian, &signature); err != nil {
		return nil, fmt.Errorf("failed to read XLS signature: %w", err)
	}
	
	if signature != 0x0809 {
		return nil, fmt.Errorf("invalid XLS file: wrong signature")
	}
	
	reader.Seek(0, 0)
	
	var records []xlsRecord
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
		
		records = append(records, record)
		
		switch record.Type {
		case 0x00FC:
			strings = append(strings, parseSST(record.Data))
		case 0x0201:
			if row := parseRow(record.Data, strings); row != nil {
				rows = append(rows, row)
			}
		}
	}
	
	if len(rows) == 0 {
		return nil, fmt.Errorf("no data found in XLS file")
	}
	
	maxCols := 0
	for _, row := range rows {
		if len(row) > maxCols {
			maxCols = len(row)
		}
	}
	
	columns := make([]string, maxCols)
	if len(rows) > 0 {
		for i, cell := range rows[0] {
			if i < maxCols {
				columns[i] = cell
			}
		}
		for i := len(rows[0]); i < maxCols; i++ {
			columns[i] = fmt.Sprintf("col_%d", i)
		}
	} else {
		for i := range columns {
			columns[i] = fmt.Sprintf("col_%d", i)
		}
	}
	
	df := NewDataFrame(columns)
	
	for i := 1; i < len(rows); i++ {
		row := make([]interface{}, maxCols)
		for j := 0; j < maxCols; j++ {
			if j < len(rows[i]) {
				row[j] = inferType(rows[i][j])
			} else {
				row[j] = nil
			}
		}
		df.AddRow(row)
	}
	
	return df, nil
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
	
	binary.Read(reader, binary.LittleEndian, &rowIndex)
	binary.Read(reader, binary.LittleEndian, &firstCol)
	binary.Read(reader, binary.LittleEndian, &lastCol)
	
	if lastCol < firstCol {
		return nil
	}
	
	row := make([]string, lastCol-firstCol+1)
	
	for i := range row {
		if reader.Len() >= 8 {
			var cellType uint16
			var cellData [6]byte
			
			binary.Read(reader, binary.LittleEndian, &cellType)
			reader.Read(cellData[:])
			
			switch cellType {
			case 0x0204:
				if len(cellData) >= 8 {
					val := binary.LittleEndian.Uint64(cellData[:])
					row[i] = fmt.Sprintf("%.2f", float64(val))
				}
			case 0x0205:
				if len(cellData) >= 4 {
					idx := binary.LittleEndian.Uint32(cellData[:4])
					if int(idx) < len(strings) {
						row[i] = strings[idx]
					}
				}
			default:
				row[i] = string(cellData[:])
			}
		}
	}
	
	return row
}