package bgen

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/buffalo-x/borm/bds"
	"github.com/buffalo-x/borm/bsql"
	"github.com/buffalo-x/borm/mdb"
	"reflect"
	"strconv"
	"strings"
)

func (genDb *GenDB) Exec() (res *bsql.Result, err error) {

	defer func() {
		if r := recover(); r != nil {
			genDb.setErr("Exec", errors.New("Panic happend:"+fmt.Sprintf("%v", r)))
			err = genDb.err
			res = &bsql.Result{LastInsertId: genDb.LastInsertId, RowsAffected: genDb.RowsAffected}
		}
	}()

	if genDb.err != nil {
		return &bsql.Result{LastInsertId: genDb.LastInsertId, RowsAffected: genDb.RowsAffected}, genDb.err
	}
	if genDb.sqlStr == "" {
		genDb.setErr("Exec", errors.New("no sql"))
		return &bsql.Result{LastInsertId: genDb.LastInsertId, RowsAffected: genDb.RowsAffected}, genDb.err
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
		return &bsql.Result{LastInsertId: genDb.LastInsertId, RowsAffected: genDb.RowsAffected}, genDb.err
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
		return &bsql.Result{LastInsertId: genDb.LastInsertId, RowsAffected: genDb.RowsAffected}, genDb.err
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
		return &bsql.Result{LastInsertId: genDb.LastInsertId, RowsAffected: genDb.RowsAffected}, genDb.err
	}

	if strings.Index(argType, "[][]interface") != 0 {
		genDb.setErr("Exec", errors.New("parameter error"))
		return &bsql.Result{LastInsertId: genDb.LastInsertId, RowsAffected: genDb.RowsAffected}, genDb.err
	}
	//if the only parameter is  a 2-d slice

	var stmt *sql.Stmt
	if genDb.sqlDb != nil {
		stmt, err = genDb.sqlDb.Prepare(genDb.sqlStr)
	} else {
		stmt, err = genDb.sqlTx.Prepare(genDb.sqlStr)
	}
	if err != nil {
		genDb.setErr("Stmt", err)
		return &bsql.Result{LastInsertId: genDb.LastInsertId, RowsAffected: genDb.RowsAffected}, genDb.err
	}
	defer stmt.Close()
	for i, vArg := range arg.([][]interface{}) {
		result, err := stmt.Exec(vArg...)
		if err != nil {
			genDb.setErr("Stmt", errors.New("line "+strconv.Itoa(i+1)+" Error:"+err.Error()))
			return &bsql.Result{LastInsertId: genDb.LastInsertId, RowsAffected: genDb.RowsAffected}, genDb.err
		} else {
			affected, err := result.RowsAffected()
			if err != nil {
				genDb.setErr("Stmt", errors.New("line "+strconv.Itoa(i+1)+" Error:"+err.Error()))
				return nil, genDb.err
			}
			genDb.RowsAffected = genDb.RowsAffected + affected
			genDb.LastInsertId, _ = result.LastInsertId()
		}
	}
	return &bsql.Result{LastInsertId: genDb.LastInsertId, RowsAffected: genDb.RowsAffected}, genDb.err
}
func (genDb *GenDB) Batch(sqls []bsql.BatchSql) (res *bsql.Result, err error) {

	defer func() {
		if r := recover(); r != nil {
			genDb.setErr("Batch", errors.New("Panic happend:"+fmt.Sprintf("%v", r)))
			err = genDb.err
			res = &bsql.Result{LastInsertId: genDb.LastInsertId, RowsAffected: genDb.RowsAffected}
		}
	}()

	if genDb.err != nil {
		return &bsql.Result{LastInsertId: genDb.LastInsertId, RowsAffected: genDb.RowsAffected}, nil
	}
	for i, batSql := range sqls {
		sqlStr, err := mdb.SqlOption(genDb.ds, batSql.Sql)
		if err != nil {
			genDb.setErr("Batch", errors.New("line "+strconv.Itoa(i+1)+" Error:"+err.Error()))
			return &bsql.Result{LastInsertId: genDb.LastInsertId, RowsAffected: genDb.RowsAffected}, nil
		}

		result, err := genDb.execSql(sqlStr, batSql.Args...)
		if err != nil {
			genDb.setErr("Batch", errors.New("line "+strconv.Itoa(i+1)+" Error:"+err.Error()))
			return &bsql.Result{LastInsertId: genDb.LastInsertId, RowsAffected: genDb.RowsAffected}, nil
		} else {
			affected, err := result.RowsAffected()
			if err != nil {
				genDb.setErr("Stmt", errors.New("line "+strconv.Itoa(i+1)+" Error:"+err.Error()))
				return nil, genDb.err
			}
			genDb.RowsAffected = genDb.RowsAffected + affected
			genDb.LastInsertId, _ = result.LastInsertId()
		}
	}
	return &bsql.Result{LastInsertId: genDb.LastInsertId, RowsAffected: genDb.RowsAffected}, genDb.err
}
func (genDb *GenDB) execSql(sqlStr string, args ...interface{}) (sql.Result, error) {
	if genDb.sqlDb != nil {
		return genDb.sqlDb.Exec(sqlStr, args...)
	} else {
		return genDb.sqlTx.Exec(sqlStr, args...)
	}
}
func Batch(sqls []bsql.BatchSql) (*bsql.Result, error) {
	genDb := &GenDB{}
	name := "default"
	ds := bds.GetDs(name)
	if ds == nil {
		genDb.setErr("Batch", errors.New("no ds found :"+name))
	}
	genDb.sqlDb = ds.SqlDb
	genDb.ds = ds
	return genDb.Batch(sqls)
}

func (genDb *GenDB) Create(tableName string, columnValueMap bsql.CV) (res *bsql.Result, err error) {

	defer func() {
		if r := recover(); r != nil {
			genDb.setErr("Create", errors.New("Panic happend:"+fmt.Sprintf("%v", r)))
			err = genDb.err
			res = &bsql.Result{LastInsertId: genDb.LastInsertId, RowsAffected: genDb.RowsAffected}
		}
	}()

	if genDb.err != nil {
		return &bsql.Result{LastInsertId: genDb.LastInsertId, RowsAffected: genDb.RowsAffected}, nil
	}
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
	result, e := genDb.execSql(sqlStr, sqlValues...)
	if e != nil {
		genDb.setErr("Create", e)
	} else {
		genDb.LastInsertId, _ = result.LastInsertId()
		genDb.RowsAffected, _ = result.RowsAffected()
	}
	return &bsql.Result{LastInsertId: genDb.LastInsertId, RowsAffected: genDb.RowsAffected}, genDb.err
}

func Create(tableName string, columnValueMap bsql.CV) (*bsql.Result, error) {
	genDb := &GenDB{}
	name := "default"
	ds := bds.GetDs(name)
	if ds == nil {
		genDb.setErr("Batch", errors.New("no ds found :"+name))
	}
	genDb.sqlDb = ds.SqlDb
	genDb.ds = ds
	return genDb.Create(tableName, columnValueMap)
}

func (genDb *GenDB) Update(tableName, column string, value interface{}) (res *bsql.Result, err error) {

	defer func() {
		if r := recover(); r != nil {
			genDb.setErr("Update", errors.New("Panic happend:"+fmt.Sprintf("%v", r)))
			err = genDb.err
			res = &bsql.Result{LastInsertId: genDb.LastInsertId, RowsAffected: genDb.RowsAffected}
		}
	}()

	if genDb.err != nil {
		return &bsql.Result{LastInsertId: genDb.LastInsertId, RowsAffected: genDb.RowsAffected}, nil
	}

	if genDb.ds == nil {
		dsName := "default"
		ds := bds.GetDs(dsName)
		if ds == nil {
			genDb.setErr("Update", errors.New("no ds found :"+dsName))
			return &bsql.Result{LastInsertId: genDb.LastInsertId, RowsAffected: genDb.RowsAffected}, nil
		}
		genDb.sqlDb = ds.SqlDb
		genDb.ds = ds
	}

	if genDb.sqlStr == "" {
		genDb.setErr("Update", errors.New("no where sql"))
		return &bsql.Result{LastInsertId: genDb.LastInsertId, RowsAffected: genDb.RowsAffected}, nil
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
			return &bsql.Result{LastInsertId: genDb.LastInsertId, RowsAffected: genDb.RowsAffected}, nil
		}
		sqlStr = sqlStr + exprValue.Result
	} else {
		sqlStr = sqlStr + "?"
		sqlValues = append(sqlValues, value)
	}

	sqlStr = sqlStr + " where " + genDb.sqlStr
	sqlValues = append(sqlValues, genDb.sqlVars...)

	result, e := genDb.execSql(sqlStr, sqlValues...)
	if e != nil {
		genDb.setErr("Update", e)
	} else {
		genDb.LastInsertId, _ = result.LastInsertId()
		genDb.RowsAffected, _ = result.RowsAffected()
	}
	return &bsql.Result{LastInsertId: genDb.LastInsertId, RowsAffected: genDb.RowsAffected}, genDb.err
}

func (genDb *GenDB) UpdateColumns(tableName string, columnValueMap bsql.CV) (res *bsql.Result, err error) {

	defer func() {
		if r := recover(); r != nil {
			genDb.setErr("Update", errors.New("Panic happend:"+fmt.Sprintf("%v", r)))
			err = genDb.err
			res = &bsql.Result{LastInsertId: genDb.LastInsertId, RowsAffected: genDb.RowsAffected}
		}
	}()

	if genDb.err != nil {
		return &bsql.Result{LastInsertId: genDb.LastInsertId, RowsAffected: genDb.RowsAffected}, nil
	}

	if genDb.sqlStr == "" {
		genDb.setErr("Update", errors.New("no where sql"))
		return &bsql.Result{LastInsertId: genDb.LastInsertId, RowsAffected: genDb.RowsAffected}, nil
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
				return &bsql.Result{LastInsertId: genDb.LastInsertId, RowsAffected: genDb.RowsAffected}, nil
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

	result, e := genDb.execSql(sqlStr, sqlValues...)
	if e != nil {
		genDb.setErr("Update", e)
	} else {
		genDb.LastInsertId, _ = result.LastInsertId()
		genDb.RowsAffected, _ = result.RowsAffected()
	}
	return &bsql.Result{LastInsertId: genDb.LastInsertId, RowsAffected: genDb.RowsAffected}, genDb.err
}
