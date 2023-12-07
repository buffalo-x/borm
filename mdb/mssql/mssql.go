package mssql

import "strings"

func (fs *FuncStruct) FirstRow(sqlStr string, dsType, dsVersion string) (string, string) {
	sqlStr = strings.TrimLeft(sqlStr, " ")
	if len(sqlStr) < 7 {
		return sqlStr, ""
	}
	sqlStr = "select top 1 " + sqlStr[6:]
	return sqlStr, ""
}
