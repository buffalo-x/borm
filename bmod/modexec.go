package bmod

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/buffalo-x/borm/bds"
	"github.com/buffalo-x/borm/bsql"
	"reflect"
	"strings"
	"time"
)

func Create(structPtr interface{}) (retModDb *ModDB) {
	modDb := &ModDB{}
	retModDb = modDb
	defer func() {
		if r := recover(); r != nil {
			modDb.setErr("Create", errors.New("Panic happend:"+fmt.Sprintf("%v", r)))
		}
	}()

	dsName := "default"
	ds := bds.GetDs(dsName)
	if ds == nil {
		modDb.setErr("Create", errors.New("no ds found :"+dsName))
		return
	}
	modDb.sqlDb = ds.SqlDb
	modDb.ds = ds

	ptrType := reflect.TypeOf(structPtr)
	structTypeName := strings.Replace(ptrType.String(), "*", "", -1)

	ormModelLock.RLock()
	model, ok := ormModelMap[structTypeName]
	ormModelLock.RUnlock()
	if ok {
		modDb.model = model
		modDb.model.tableName = model.tableName
	} else {
		model, err := modelStructPtr(structPtr)
		if err != nil {
			modDb.setErr("Create", err)
			return
		}
		modDb.model = model
	}

	sqlStr, sqlValues, err := makeCreateSql(structPtr, modDb)
	if err != nil {
		modDb.setErr("Create", err)
		return
	}
	result, err := modDb.execSql(sqlStr, sqlValues...)
	if err != nil {
		modDb.setErr("Create", err)
	} else {
		modDb.LastInsertId, _ = result.LastInsertId()
		modDb.RowsAffected, _ = result.RowsAffected()
	}

	if modDb.model.primaryKey != "" {
		keyField, _ := modDb.model.fieldMap[modDb.model.primaryKey]
		if keyField.autoIncrement {
			ptrValue := reflect.ValueOf(structPtr)
			structValue := ptrValue.Elem()
			structValue.FieldByName(modDb.model.primaryKey).SetInt(modDb.LastInsertId)
		}
	}
	return
}
func Save(structPtr interface{}) (retModDb *ModDB) {
	modDb := &ModDB{}
	retModDb = modDb
	defer func() {
		if r := recover(); r != nil {
			modDb.setErr("Save", errors.New("Panic happend:"+fmt.Sprintf("%v", r)))
		}
	}()

	dsName := "default"
	ds := bds.GetDs(dsName)
	if ds == nil {
		modDb.setErr("Save", errors.New("no ds found :"+dsName))
		return
	}
	modDb.sqlDb = ds.SqlDb
	modDb.ds = ds

	ptrType := reflect.TypeOf(structPtr)
	structTypeName := strings.Replace(ptrType.String(), "*", "", -1)

	ormModelLock.RLock()
	model, ok := ormModelMap[structTypeName]
	ormModelLock.RUnlock()
	if ok {
		modDb.model = model
	} else {
		model, err := modelStructPtr(structPtr)
		if err != nil {
			modDb.setErr("Save", err)
			return
		}
		modDb.model = model
	}

	if modDb.model.primaryKey == "" {
		modDb.setErr("Save", errors.New("no primary key"))
		return
	}
	sqlStr, sqlValues, err := makeSaveSql(structPtr, modDb)
	if err != nil {
		modDb.setErr("Save", err)
		return
	}
	result, err := modDb.execSql(sqlStr, sqlValues...)
	if err != nil {
		modDb.setErr("Save", err)
	} else {
		modDb.LastInsertId, _ = result.LastInsertId()
		modDb.RowsAffected, _ = result.RowsAffected()
	}
	return
}
func Delete(structPtr interface{}) (retModDb *ModDB) {
	modDb := &ModDB{}
	retModDb = modDb
	defer func() {
		if r := recover(); r != nil {
			modDb.setErr("Delete", errors.New("Panic happend:"+fmt.Sprintf("%v", r)))
		}
	}()

	dsName := "default"
	ds := bds.GetDs(dsName)
	if ds == nil {
		modDb.setErr("Delete", errors.New("no ds found :"+dsName))
		return
	}
	modDb.sqlDb = ds.SqlDb
	modDb.ds = ds

	ptrType := reflect.TypeOf(structPtr)
	structTypeName := strings.Replace(ptrType.String(), "*", "", -1)
	ormModelLock.RLock()
	model, ok := ormModelMap[structTypeName]
	ormModelLock.RUnlock()
	if ok {
		modDb.model = model
	} else {
		model, err := modelStructPtr(structPtr)
		if err != nil {
			modDb.setErr("Delete", err)
			return
		}
		modDb.model = model
	}

	if modDb.model.primaryKey == "" {
		modDb.setErr("Delete", errors.New("no primary key"))
		return
	}

	ptrValue := reflect.ValueOf(structPtr)
	structValue := ptrValue.Elem()

	primaryKeyField, _ := modDb.model.fieldMap[modDb.model.primaryKey]
	sqlValues := make([]interface{}, 0, 1)
	sqlValues = append(sqlValues, structValue.FieldByName(primaryKeyField.name).Interface())
	execSql := "delete from " + modDb.model.tableName + " where " + primaryKeyField.column + "=?"

	result, err := modDb.execSql(execSql, sqlValues...)
	if err != nil {
		modDb.setErr("Delete", err)
	} else {
		modDb.LastInsertId, _ = result.LastInsertId()
		modDb.RowsAffected, _ = result.RowsAffected()
	}

	return
}

func (modDb *ModDB) Create(structPtr interface{}) (retModDb *ModDB) {
	retModDb = modDb
	defer func() {
		if r := recover(); r != nil {
			modDb.setErr("Create", errors.New("Panic happend:"+fmt.Sprintf("%v", r)))
		}
	}()
	if modDb.err != nil {
		return
	}
	if modDb.ds == nil {
		dsName := "default"
		ds := bds.GetDs(dsName)
		if ds == nil {
			modDb.setErr("Create", errors.New("no ds found :"+dsName))
			return
		}
		modDb.sqlDb = ds.SqlDb
		modDb.ds = ds
	}

	ptrType := reflect.TypeOf(structPtr)
	structTypeName := strings.Replace(ptrType.String(), "*", "", -1)

	if modDb.model != nil {
		if modDb.model.name != structTypeName {
			modDb.setErr("Create", errors.New("model inconsistency "))
			return
		}
	} else {
		ormModelLock.RLock()
		model, ok := ormModelMap[structTypeName]
		ormModelLock.RUnlock()
		if ok {
			modDb.model = model
			modDb.model.tableName = model.tableName
		} else {
			model, err := modelStructPtr(structPtr)
			if err != nil {
				modDb.setErr("Create", err)
				return
			}
			modDb.model = model
			modDb.model.tableName = model.tableName
		}
	}

	sqlStr, sqlValues, err := makeCreateSql(structPtr, modDb)
	if err != nil {
		modDb.setErr("Create", err)
		return
	}
	result, err := modDb.execSql(sqlStr, sqlValues...)
	if err != nil {
		modDb.setErr("Create", err)
	} else {
		modDb.LastInsertId, _ = result.LastInsertId()
		modDb.RowsAffected, _ = result.RowsAffected()
	}

	if modDb.model.primaryKey != "" {
		keyField, _ := modDb.model.fieldMap[modDb.model.primaryKey]
		if keyField.autoIncrement {
			ptrValue := reflect.ValueOf(structPtr)
			structValue := ptrValue.Elem()
			structValue.FieldByName(modDb.model.primaryKey).SetInt(modDb.LastInsertId)
		}
	}
	return
}
func (modDb *ModDB) Save(structPtr interface{}) (retModDb *ModDB) {
	retModDb = modDb
	defer func() {
		if r := recover(); r != nil {
			modDb.setErr("Save", errors.New("Panic happend:"+fmt.Sprintf("%v", r)))
		}
	}()
	if modDb.err != nil {
		return
	}
	if modDb.ds == nil {
		dsName := "default"
		ds := bds.GetDs(dsName)
		if ds == nil {
			modDb.setErr("Save", errors.New("no ds found :"+dsName))
			return
		}
		modDb.sqlDb = ds.SqlDb
		modDb.ds = ds
	}

	ptrType := reflect.TypeOf(structPtr)
	structTypeName := strings.Replace(ptrType.String(), "*", "", -1)

	if modDb.model != nil {
		if modDb.model.name != structTypeName {
			modDb.setErr("Save", errors.New("model inconsistency "))
			return
		}
	} else {
		ormModelLock.RLock()
		model, ok := ormModelMap[structTypeName]
		ormModelLock.RUnlock()
		if ok {
			modDb.model = model
			modDb.model.tableName = model.tableName
		} else {
			model, err := modelStructPtr(structPtr)
			if err != nil {
				modDb.setErr("Save", err)
				return
			}
			modDb.model = model
			modDb.model.tableName = model.tableName
		}
	}

	if modDb.model.primaryKey == "" {
		modDb.setErr("Save", errors.New("no primary key"))
		return
	}
	sqlStr, sqlValues, err := makeSaveSql(structPtr, modDb)
	if err != nil {
		modDb.setErr("Save", err)
		return
	}
	result, err := modDb.execSql(sqlStr, sqlValues...)
	if err != nil {
		modDb.setErr("Save", err)
	} else {
		modDb.LastInsertId, _ = result.LastInsertId()
		modDb.RowsAffected, _ = result.RowsAffected()
	}
	return
}
func (modDb *ModDB) Delete(structPtr interface{}) (retModDb *ModDB) {
	retModDb = modDb
	defer func() {
		if r := recover(); r != nil {
			modDb.setErr("Delete", errors.New("Panic happend:"+fmt.Sprintf("%v", r)))
		}
	}()
	if modDb.err != nil {
		return
	}
	if modDb.ds == nil {
		dsName := "default"
		ds := bds.GetDs(dsName)
		if ds == nil {
			modDb.setErr("Delete", errors.New("no ds found :"+dsName))
			return
		}
		modDb.sqlDb = ds.SqlDb
		modDb.ds = ds
	}

	ptrType := reflect.TypeOf(structPtr)
	structTypeName := strings.Replace(ptrType.String(), "*", "", -1)

	if modDb.model != nil {
		if modDb.model.name != structTypeName {
			modDb.setErr("Delete", errors.New("model inconsistency "))
			return
		}
	} else {
		ormModelLock.RLock()
		model, ok := ormModelMap[structTypeName]
		ormModelLock.RUnlock()
		if ok {
			modDb.model = model
		} else {
			model, err := modelStructPtr(structPtr)
			if err != nil {
				modDb.setErr("Delete", err)
				return
			}
			modDb.model = model
		}
	}

	if modDb.model.primaryKey == "" {
		modDb.setErr("Delete", errors.New("no primary key"))
		return
	}

	ptrValue := reflect.ValueOf(structPtr)
	structValue := ptrValue.Elem()

	primaryKeyField, _ := modDb.model.fieldMap[modDb.model.primaryKey]
	sqlValues := make([]interface{}, 0, 1)
	sqlValues = append(sqlValues, structValue.FieldByName(primaryKeyField.name).Interface())
	sqlStr := "delete from " + modDb.model.tableName + " where " + primaryKeyField.column + "=?"

	result, err := modDb.execSql(sqlStr, sqlValues...)
	if err != nil {
		modDb.setErr("Delete", err)
	} else {
		modDb.LastInsertId, _ = result.LastInsertId()
		modDb.RowsAffected, _ = result.RowsAffected()
	}
	return
}
func (modDb *ModDB) Update(column string, value interface{}) (retModDb *ModDB) {
	retModDb = modDb
	defer func() {
		if r := recover(); r != nil {
			modDb.setErr("Update", errors.New("Panic happend:"+fmt.Sprintf("%v", r)))
		}
	}()

	if modDb.err != nil {
		return
	}
	if modDb.ds == nil {
		dsName := "default"
		ds := bds.GetDs(dsName)
		if ds == nil {
			modDb.setErr("Update", errors.New("no ds found :"+dsName))
			return
		}
		modDb.sqlDb = ds.SqlDb
		modDb.ds = ds
	}
	if modDb.model == nil {
		modDb.setErr("Update", errors.New("no model"))
		return
	}
	if modDb.where == "" {
		if modDb.primaryKeyValue == nil {
			modDb.setErr("Update", errors.New("no where and no modelValue"))
			return
		} else if modDb.model.primaryKey == "" {
			modDb.setErr("Update", errors.New("no primary key"))
			return
		}
	}
	sqlValues := make([]interface{}, 0, 5)
	sqlStr := "update " + modDb.model.tableName + " set " + column + "="

	if value == nil {
		sqlStr = sqlStr + column + "=" + "?"
		sqlValues = append(sqlValues, nil)
	} else if reflect.TypeOf(value).String() == "*bsql.SqlExpr" {
		exprValue := value.(*bsql.SqlExpr)
		if exprValue.Err != nil {
			modDb.setErr("Update", errors.New("parameter err :"+exprValue.Err.Error()))
			return
		}
		sqlStr = sqlStr + value.(*bsql.SqlExpr).Result
	} else {
		sqlStr = sqlStr + "?"
		sqlValues = append(sqlValues, value)
	}

	if modDb.where != "" {
		sqlStr = sqlStr + " where " + modDb.where
		sqlValues = append(sqlValues, modDb.whereVars...)
	} else {
		primaryKeyField, _ := modDb.model.fieldMap[modDb.model.primaryKey]
		sqlStr = sqlStr + " where " + primaryKeyField.column + " = ?"
		sqlValues = append(sqlValues, modDb.primaryKeyValue)
	}
	result, err := modDb.execSql(sqlStr, sqlValues...)
	if err != nil {
		modDb.setErr("Update", err)
	} else {
		modDb.LastInsertId, _ = result.LastInsertId()
		modDb.RowsAffected, _ = result.RowsAffected()
	}
	return
}
func (modDb *ModDB) UpdateColumns(columnValueMap bsql.CV) (retModDb *ModDB) {
	retModDb = modDb
	defer func() {
		if r := recover(); r != nil {
			modDb.setErr("Update", errors.New("Panic happend:"+fmt.Sprintf("%v", r)))
		}
	}()

	if len(columnValueMap) == 0 {
		modDb.setErr("Update", errors.New("no update data "))
		return
	}

	if modDb.err != nil {
		return
	}
	if modDb.ds == nil {
		dsName := "default"
		ds := bds.GetDs(dsName)
		if ds == nil {
			modDb.setErr("Delete", errors.New("no ds found :"+dsName))
			return
		}
		modDb.sqlDb = ds.SqlDb
		modDb.ds = ds
	}
	if modDb.model == nil {
		modDb.setErr("Update", errors.New("no model"))
		return
	}
	if modDb.where == "" {
		if modDb.primaryKeyValue == nil {
			modDb.setErr("Update", errors.New("no where and no modelValue"))
			return
		} else if modDb.model.primaryKey == "" {
			modDb.setErr("Update", errors.New("no primary key"))
			return
		}
	}
	sqlValues := make([]interface{}, 0, 5)
	sqlStr := "update " + modDb.model.tableName + " set "
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
				modDb.setErr("Update", errors.New("parameter err :"+exprValue.Err.Error()))
				return
			}
			sqlStr = sqlStr + column + "=" + value.(*bsql.SqlExpr).Result
		} else {
			sqlStr = sqlStr + column + "=" + "?"
			sqlValues = append(sqlValues, value)
		}
		i++
	}

	if modDb.where != "" {
		sqlStr = sqlStr + " where " + modDb.where
		sqlValues = append(sqlValues, modDb.whereVars...)
	} else {
		primaryKeyField, _ := modDb.model.fieldMap[modDb.model.primaryKey]
		sqlStr = sqlStr + " where " + primaryKeyField.column + " = ?"
		sqlValues = append(sqlValues, modDb.primaryKeyValue)
	}
	result, err := modDb.execSql(sqlStr, sqlValues...)
	if err != nil {
		modDb.setErr("Update", err)
	} else {
		modDb.LastInsertId, _ = result.LastInsertId()
		modDb.RowsAffected, _ = result.RowsAffected()
	}
	return
}

func (modDb *ModDB) execSql(sqlStr string, args ...interface{}) (sql.Result, error) {
	if modDb.sqlDb != nil {
		return modDb.sqlDb.Exec(sqlStr, args...)
	} else {
		return modDb.sqlTx.Exec(sqlStr, args...)
	}
}

func makeSaveSql(structPtr interface{}, modDb *ModDB) (string, []interface{}, error) {

	ptrValue := reflect.ValueOf(structPtr)
	structValue := ptrValue.Elem()

	sqlExprs := ""
	sqlValues := make([]interface{}, 0, 10)

	for _, ormField := range modDb.model.columnMap {
		if ormField.excluded {
			continue
		}
		if ormField.readonly {
			continue
		}
		if ormField.autoIncrement {
			continue
		}
		if ormField.primaryKey {
			continue
		}
		if ormField.autoCreateTime && !ormField.autoUpdateTime {
			continue
		}
		expr := "?"
		if ormField.autoUpdateTime {
			structValue.FieldByName(ormField.name).Set(reflect.ValueOf(time.Now()))
		}
		if sqlExprs != "" {
			sqlExprs = sqlExprs + ","
		}
		sqlExprs = sqlExprs + ormField.column + "=" + expr
		if expr == "?" {
			sqlValues = append(sqlValues, structValue.FieldByName(ormField.name).Interface())
		}
	}

	primaryKeyField, _ := modDb.model.fieldMap[modDb.model.primaryKey]

	sqlValues = append(sqlValues, structValue.FieldByName(primaryKeyField.name).Interface())
	execSql := "update " + modDb.model.tableName + " set " + sqlExprs + " where " + primaryKeyField.column + "=?"
	return execSql, sqlValues, nil
}
func makeCreateSql(structPtr interface{}, modDb *ModDB) (string, []interface{}, error) {
	ptrValue := reflect.ValueOf(structPtr)
	structValue := ptrValue.Elem()

	sqlCols := ""
	sqlExprs := ""
	sqlValues := make([]interface{}, 0, 10)

	for _, ormField := range modDb.model.columnMap {
		if ormField.excluded {
			continue
		}
		if ormField.readonly {
			continue
		}
		if ormField.autoIncrement {
			continue
		}
		if ormField.autoUpdateTime && !ormField.autoCreateTime {
			continue
		}

		if sqlCols == "" {
			sqlCols = ormField.column
		} else {
			sqlCols = sqlCols + "," + ormField.column
		}
		expr := "?"
		if ormField.autoCreateTime {
			structValue.FieldByName(ormField.name).Set(reflect.ValueOf(time.Now()))
		}
		if sqlExprs == "" {
			sqlExprs = expr
		} else {
			sqlExprs = sqlExprs + "," + expr
		}
		if expr == "?" {
			sqlValues = append(sqlValues, structValue.FieldByName(ormField.name).Interface())
		}
	}
	sqlStr := "insert into " + modDb.model.tableName + "(" + sqlCols + ") values(" + sqlExprs + ")"
	return sqlStr, sqlValues, nil
}
