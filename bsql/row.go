package bsql

import (
	"database/sql"
	"errors"
)

type Rows struct {
	Count    int
	Data     [][]string
	ColsMap  map[string]int
	ColsName []string
}
type Row struct {
	Data     []string
	ColsMap  map[string]int
	ColsName []string
}

func FetchRows(rows *sql.Rows) *Rows {
	if rows == nil {
		return nil
	}

	defer rows.Close() //为了安全，避免异常时后没有自动关闭

	cols, err := rows.Columns()
	var dbRows = Rows{}
	dbRows.Count = 0
	dbRows.ColsName = cols
	colsCt := len(cols)
	dbRows.ColsMap = make(map[string]int)
	for i, v := range cols {
		dbRows.ColsMap[v] = i
	}

	dbRows.Data = make([][]string, 0, 30)
	values := make([]sql.RawBytes, colsCt)
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			return nil // proper error handling instead of panic in your app
		}
		result := make([]string, colsCt)
		for i, v := range values {
			result[i] = string(v)
		}
		dbRows.Data = append(dbRows.Data, result)
		dbRows.Count++
	}
	return &dbRows
}
func FetchRowsE(rows *sql.Rows) (*Rows, error) {
	if rows == nil {
		return nil, nil
	}

	defer rows.Close() //为了安全，避免异常时后没有自动关闭

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	var dbRows = Rows{}
	dbRows.Count = 0
	dbRows.ColsName = cols
	colsCt := len(cols)
	dbRows.ColsMap = make(map[string]int)
	for i, v := range cols {
		dbRows.ColsMap[v] = i
	}

	dbRows.Data = make([][]string, 0, 30)
	values := make([]sql.RawBytes, colsCt)
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			return nil, err
		}
		result := make([]string, colsCt)
		for i, v := range values {
			result[i] = string(v)
		}
		dbRows.Data = append(dbRows.Data, result)
		dbRows.Count++
	}
	return &dbRows, nil
}
func FetchFirstRowE(rows *sql.Rows) (*Row, error) {
	if rows == nil {
		return nil, errors.New("no row")
	}
	defer rows.Close() //为了安全，避免异常时后没有自动关闭
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	var dbRow = Row{}
	dbRow.ColsName = cols
	dbRow.ColsMap = make(map[string]int)
	for i, v := range cols {
		dbRow.ColsMap[v] = i
	}
	colsCt := len(cols)
	dbRow.Data = make([]string, colsCt, colsCt)
	values := make([]sql.RawBytes, colsCt)
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	hasData := false
	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			return nil, err
		}
		for i, v := range values {
			dbRow.Data[i] = string(v)
		}
		hasData = true
		break
	}
	if hasData {
		return &dbRow, nil
	}
	return nil, errors.New("no row")
}
func ToBsqlRows(rows *sql.Rows, brows *Rows) error {
	if rows == nil {
		return nil
	}
	defer rows.Close() //为了安全，避免异常时后没有自动关闭

	cols, err := rows.Columns()
	if err != nil {
		return err
	}

	brows.Count = 0
	brows.ColsName = cols
	colsCt := len(cols)
	brows.ColsMap = make(map[string]int)
	for i, v := range cols {
		brows.ColsMap[v] = i
	}

	brows.Data = make([][]string, 0, 30)
	values := make([]sql.RawBytes, colsCt)
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			return err
		}
		result := make([]string, colsCt)
		for i, v := range values {
			result[i] = string(v)
		}
		brows.Data = append(brows.Data, result)
		brows.Count++
	}
	return nil
}

func (rows *Rows) GetRow(rowNo int) *Row {
	if rowNo >= len(rows.Data) || rowNo < 0 {
		return nil
	}
	row := Row{}
	row.ColsMap = rows.ColsMap
	row.ColsName = rows.ColsName
	row.Data = rows.Data[rowNo]
	return &row
}
func (rows *Rows) GetValue(rowNo int, colName string) string {
	if colIdx, ok := rows.ColsMap[colName]; !ok {
		return ""
	} else if rowNo >= len(rows.Data) || rowNo < 0 {
		return ""
	} else {
		val := rows.Data[rowNo][colIdx]
		return val
	}
}
func (rows *Rows) GetValueE(rowNo int, colName string) (string, error) {
	if colIdx, ok := rows.ColsMap[colName]; !ok {
		return "", errors.New("no column")
	} else {
		if rowNo >= len(rows.Data) {
			return "", errors.New("no row")
		}
		val := rows.Data[rowNo][colIdx]
		return val, nil
	}
}
