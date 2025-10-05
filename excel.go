package gopandas

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
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