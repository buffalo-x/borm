package bgen

import (
	"borm/bds"
	"borm/bsql"
	"borm/mdb"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func (genDb *GenDB) Exec() (retGenDb *GenDB) {
	retGenDb = genDb
	if genDb.err != nil {
		return
	}
	if genDb.sqlStr == "" {
		genDb.setErr("Exec", errors.New("no sql"))
		return
	}

	args := genDb.sqlVars
	//no parameter or multi parameters
	if len(args) == 0 || len(args) > 1 {
		result, err := genDb.execSql(genDb.sqlStr, args...)
		if err != nil {
			genDb.setErr("Exec", err)
		} else {
			genDb.RowsAffected, _ = result.RowsAffected()
			genDb.LastInsertId, _ = result.LastInsertId()
		}
		return
	}

	arg := args[0]
	argType := reflect.TypeOf(arg).String()
	//if the only parameter is not a slice
	if strings.Index(argType, "[]interface") < 0 {
		result, err := genDb.execSql(genDb.sqlStr, arg)
		if err != nil {
			genDb.setErr("Exec", err)
		} else {
			genDb.RowsAffected, _ = result.RowsAffected()
			genDb.LastInsertId, _ = result.LastInsertId()
		}
		return
	}

	//if the only parameter is  a 1-d slice
	if strings.Index(argType, "[]interface") == 0 {
		result, err := genDb.execSql(genDb.sqlStr, arg.([]interface{})...)
		if err != nil {
			genDb.setErr("Exec", err)
		} else {
			genDb.RowsAffected, _ = result.RowsAffected()
			genDb.LastInsertId, _ = result.LastInsertId()
		}
		return
	}

	if strings.Index(argType, "[][]interface") != 0 {
		genDb.setErr("Exec", errors.New("parameter error"))
		return
	}
	//if the only parameter is  a 2-d slice
	var err error
	var stmt *sql.Stmt
	if genDb.sqlDb != nil {
		stmt, err = genDb.sqlDb.Prepare(genDb.sqlStr)
	} else {
		stmt, err = genDb.sqlTx.Prepare(genDb.sqlStr)
	}
	if err != nil {
		genDb.setErr("Stmt", err)
		return
	}
	defer stmt.Close()
	for i, vArg := range arg.([][]interface{}) {
		result, err := stmt.Exec(vArg...)
		if err != nil {
			genDb.setErr("Stmt", errors.New("line "+strconv.Itoa(i+1)+" Error:"+err.Error()))
			return
		} else {
			genDb.RowsAffected, _ = result.RowsAffected()
			genDb.LastInsertId, _ = result.LastInsertId()
		}
	}

	return
}
func (genDb *GenDB) Batch(sqls []bsql.BatchSql) (retGenDb *GenDB) {
	retGenDb = genDb
	if genDb.err != nil {
		return
	}
	for i, batSql := range sqls {

		sqlStr, err := mdb.SqlOption(genDb.ds, batSql.Sql)
		if err != nil {
			genDb.setErr("Batch", errors.New("line "+strconv.Itoa(i+1)+" Error:"+err.Error()))
			return
		}

		result, err := genDb.execSql(sqlStr, batSql.Args...)
		if err != nil {
			genDb.setErr("Batch", errors.New("line "+strconv.Itoa(i+1)+" Error:"+err.Error()))
			return
		} else {
			genDb.RowsAffected, _ = result.RowsAffected()
			genDb.LastInsertId, _ = result.LastInsertId()
		}
	}
	return
}
func (genDb *GenDB) execSql(sqlStr string, args ...interface{}) (sql.Result, error) {
	if genDb.sqlDb != nil {
		return genDb.sqlDb.Exec(sqlStr, args...)
	} else {
		return genDb.sqlTx.Exec(sqlStr, args...)
	}
}
func Batch(sqls []bsql.BatchSql) (genDb *GenDB) {
	genDb = &GenDB{}
	name := "default"
	ds := bds.GetDs(name)
	if ds == nil {
		genDb.setErr("Batch", errors.New("no ds found :"+name))
		return
	}
	genDb.sqlDb = ds.SqlDb
	genDb.ds = ds
	genDb.Batch(sqls)
	return
}

func (genDb *GenDB) Create(tableName string, columnValueMap bsql.CV) (retGenDb *GenDB) {
	retGenDb = genDb
	if genDb.err != nil {
		return
	}
	name := "default"
	ds := bds.GetDs(name)
	if ds == nil {
		genDb.setErr("Create", errors.New("no ds found :"+name))
		return
	}
	genDb.sqlDb = ds.SqlDb
	genDb.ds = ds

	sqlValues := make([]interface{}, 0, 5)
	sqlStr := "insert into " + tableName + " ( "
	sqlExprs := " values("
	i := 0
	for column, value := range columnValueMap {
		if i != 0 {
			sqlStr = sqlStr + ","
			sqlExprs = sqlExprs + ","
		}
		sqlStr = sqlStr + column
		sqlExprs = sqlExprs + "?"
		if value == nil {
			sqlValues = append(sqlValues, nil)
		} else {
			sqlValues = append(sqlValues, value)
		}
		i++
	}
	sqlStr = sqlStr + ")" + sqlExprs + ")"
	result, err := genDb.execSql(sqlStr, sqlValues...)
	if err != nil {
		genDb.setErr("Create", err)
	} else {
		genDb.LastInsertId, _ = result.LastInsertId()
		genDb.RowsAffected, _ = result.RowsAffected()
	}
	return
}
func Create(tableName string, columnValueMap bsql.CV) (genDb *GenDB) {
	genDb = &GenDB{}
	genDb.Create(tableName, columnValueMap)
	return
}
func (genDb *GenDB) Update(tableName, column string, value interface{}) (retGenDb *GenDB) {
	retGenDb = genDb
	defer func() {
		if r := recover(); r != nil {
			genDb.setErr("Update", errors.New("Panic happend:"+fmt.Sprintf("%v", r)))
		}
	}()

	if genDb.err != nil {
		return
	}
	if genDb.ds == nil {
		dsName := "default"
		ds := bds.GetDs(dsName)
		if ds == nil {
			genDb.setErr("Update", errors.New("no ds found :"+dsName))
			return
		}
		genDb.sqlDb = ds.SqlDb
		genDb.ds = ds
	}

	if genDb.sqlStr == "" {
		genDb.setErr("Update", errors.New("no sql"))
		return
	}
	sqlValues := make([]interface{}, 0, 5)
	sqlStr := "update " + tableName + " set " + column + "="

	if value == nil {
		sqlStr = sqlStr + column + "=" + "?"
		sqlValues = append(sqlValues, nil)
	} else if reflect.TypeOf(value).String() == "*bsql.SqlExpr" {
		exprValue := value.(*bsql.SqlExpr)
		if exprValue.Err != nil {
			genDb.setErr("Update", errors.New("parameter err :"+exprValue.Err.Error()))
			return
		}
		sqlStr = sqlStr + exprValue.Result
	} else {
		sqlStr = sqlStr + "?"
		sqlValues = append(sqlValues, value)
	}

	sqlStr = sqlStr + " where " + genDb.sqlStr
	sqlValues = append(sqlValues, genDb.sqlVars...)

	result, err := genDb.execSql(sqlStr, sqlValues...)
	if err != nil {
		genDb.setErr("Update", err)
	} else {
		genDb.LastInsertId, _ = result.LastInsertId()
		genDb.RowsAffected, _ = result.RowsAffected()
	}
	return
}

func (genDb *GenDB) UpdateColumns(tableName string, columnValueMap bsql.CV) (retGenDb *GenDB) {
	retGenDb = genDb
	defer func() {
		if r := recover(); r != nil {
			genDb.setErr("Update", errors.New("Panic happend:"+fmt.Sprintf("%v", r)))
		}
	}()

	if genDb.err != nil {
		return
	}
	if genDb.ds == nil {
		dsName := "default"
		ds := bds.GetDs(dsName)
		if ds == nil {
			genDb.setErr("Update", errors.New("no ds found :"+dsName))
			return
		}
		genDb.sqlDb = ds.SqlDb
		genDb.ds = ds
	}

	if genDb.sqlStr == "" {
		genDb.setErr("Update", errors.New("no sql"))
		return
	}

	sqlValues := make([]interface{}, 0, 5)
	sqlStr := "update " + tableName + " set "
	i := 0
	for column, value := range columnValueMap {
		if i != 0 {
			sqlStr = sqlStr + ","
		}
		if value == nil {
			sqlStr = sqlStr + column + "=" + "?"
			sqlValues = append(sqlValues, nil)
		} else if reflect.TypeOf(value).String() == "*bsql.SqlExpr" {
			exprValue := value.(*bsql.SqlExpr)
			if exprValue.Err != nil {
				genDb.setErr("Update", errors.New("parameter err :"+exprValue.Err.Error()))
				return
			}
			sqlStr = sqlStr + column + "=" + exprValue.Result
		} else {
			sqlStr = sqlStr + column + "=" + "?"
			sqlValues = append(sqlValues, value)
		}
		i++
	}

	sqlStr = sqlStr + " where " + genDb.sqlStr
	sqlValues = append(sqlValues, genDb.sqlVars...)

	result, err := genDb.execSql(sqlStr, sqlValues...)
	if err != nil {
		genDb.setErr("Update", err)
	} else {
		genDb.LastInsertId, _ = result.LastInsertId()
		genDb.RowsAffected, _ = result.RowsAffected()
	}
	return
}
