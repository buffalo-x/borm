package mysql

func (fs *FuncStruct) FirstRow(sqlStr string, dsType, dsVersion string) (string, string) {
	return sqlStr + " limit 0,1", ""
}
