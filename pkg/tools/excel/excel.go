package excel

import "github.com/xuri/excelize/v2"



// ReadExcel 读取Excel文件
func ReadExcel(filename string, sheet string) ([][]string, error) {
	f, err := excelize.OpenFile(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

// CreateExcel 创建Excel文件
func CreateExcel(filename string, sheet string, data [][]interface{}) error {
	f := excelize.NewFile()

	for i, row := range data {
		for j, cell := range row {
			cell := excelize.Cell{Value: cell}
			axis, _ := excelize.CoordinatesToCellName(j+1, i+1)
			f.SetCellValue(sheet, axis, cell.Value)
		}
	}

	return f.SaveAs(filename)
}
