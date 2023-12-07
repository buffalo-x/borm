package postgresql

func (fs *FuncStruct) FirstRow(sqlStr string, dsType, dsVersion string) (string, string) {
	return sqlStr + " OFFSET 0 LIMIT 1", ""
}
