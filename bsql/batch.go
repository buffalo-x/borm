package bsql

type BatchSql struct {
	Sql  string
	Args []interface{}
}

func NewBtchSql(sql string, args ...interface{}) BatchSql {
	return BatchSql{Sql: sql, Args: args}
}
