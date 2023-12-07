package bmod

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/buffalo-x/borm/bds"
	"github.com/buffalo-x/borm/mdb"
	"reflect"
	"strconv"
	"time"
)

func (modDb *ModDB) Query(slicePtr interface{}) (retModDb *ModDB) {
	retModDb = modDb
	defer func() {
		if r := recover(); r != nil {
			modDb.setErr("Query", errors.New("Panic happend:"+fmt.Sprintf("%v", r)))
		}
	}()
	if modDb.err != nil {
		return
	}
	slicePtrValue := reflect.ValueOf(slicePtr)
	if slicePtrValue.Kind() != reflect.Ptr {
		modDb.setErr("Query", errors.New("parameter is not a pointer"))
		return
	}

	sliceType := reflect.Indirect(slicePtrValue).Type()
	if sliceType.Kind() != reflect.Slice {
		modDb.setErr("Query", errors.New("parameter should be a slice pointer"))
		return
	}

	structType := sliceType.Elem()
	if structType.Kind() != reflect.Struct {
		modDb.setErr("Query", errors.New("parameter should be like &[]struct"))
		return
	}
	structTypeName := structType.String()
	if modDb.model != nil {
		if modDb.model.name != structTypeName {
			modDb.setErr("Query", errors.New("model inconsistency "))
			return
		}
	} else {
		ormModelLock.RLock()
		model, ok := ormModelMap[structTypeName]
		ormModelLock.RUnlock()
		if ok {
			modDb.model = model
		} else {
			model, err := modelStructType(structType)
			if err != nil {
				modDb.setErr("Query", err)
				return
			}
			modDb.model = model
		}
	}

	sqlCols := getModelQueryColumnList(modDb.model)
	sqlStr := "select " + sqlCols + " from " + modDb.model.tableName
	if modDb.where != "" {
		sqlStr = sqlStr + " where " + modDb.where
	}
	rows, err := modDb.execQry(sqlStr, modDb.whereVars...)
	if err != nil {
		modDb.setErr("Query", err)
		return
	}
	defer rows.Close()

	retSliceType := reflect.SliceOf(structType) //reflect.PtrTo(modelType)
	retSlice := reflect.MakeSlice(retSliceType, 0, 10)

	columns, err := rows.Columns()
	if err != nil {
		modDb.setErr("Query", err)
		return
	}
	values := make([]interface{}, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	for rows.Next() {
		structValue := reflect.New(structType)
		err = rows.Scan(scanArgs...)
		if err != nil {
			modDb.setErr("Query", err)
			return
		}
		for idx, col := range columns {
			field, ok := modDb.model.columnMap[col]
			if !ok {
				modDb.setErr("Query", errors.New("no column:"+col))
				return
			}
			fieldValue := reflect.New(field.goType)
			err := convertData(fieldValue, values[idx])
			if err != nil {
				modDb.setErr("Query", err)
				return
			}
			structValue.Elem().FieldByName(field.name).Set(fieldValue.Elem())
		}
		retSlice = reflect.Append(retSlice, structValue.Elem())
	}
	slicePtrValue.Elem().Set(retSlice)
	return
}
func (modDb *ModDB) First(structPtr interface{}) (retModDb *ModDB) {
	retModDb = modDb
	if modDb.err != nil {
		return
	}
	structPtrValue := reflect.ValueOf(structPtr)
	if structPtrValue.Kind() != reflect.Ptr {
		modDb.setErr("First", errors.New("parameter is not a pointer"))
		return
	}
	structValue := structPtrValue.Elem()
	if structValue.Kind() != reflect.Struct {
		modDb.setErr("Query", errors.New("parameter should be like &struct"))
		return
	}
	structType := structValue.Type()

	structTypeName := structType.String()
	if modDb.model != nil {
		if modDb.model.name != structTypeName {
			modDb.setErr("Query", errors.New("model inconsistency "))
			return
		}
	} else {
		ormModelLock.RLock()
		model, ok := ormModelMap[structTypeName]
		ormModelLock.RUnlock()
		if ok {
			modDb.model = model
		} else {
			model, err := modelStructType(structType)
			if err != nil {
				modDb.setErr("Query", err)
				return
			}
			modDb.model = model
		}
	}

	sqlCols := getModelQueryColumnList(modDb.model)
	sqlStr := "select " + sqlCols + " from " + modDb.model.tableName
	if modDb.where != "" {
		sqlStr = sqlStr + " where " + modDb.where
	}

	newSql, info := mdb.FirstRow(sqlStr, modDb.ds)
	if info != "" {
		modDb.setErr("First", errors.New(info))
		return
	}

	rows, err := modDb.execQry(newSql, modDb.whereVars...)
	if err != nil {
		modDb.setErr("Query", err)
		return
	}
	rows.Next()
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		modDb.setErr("First", err)
		return
	}

	values := make([]interface{}, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	err = rows.Scan(scanArgs...)
	if err != nil {
		modDb.setErr("First", err)
		return
	}
	for idx, col := range columns {
		field, ok := modDb.model.columnMap[col]
		if !ok {
			modDb.setErr("First", errors.New("no column:"+col))
			return
		}
		fieldValue := reflect.New(field.goType)
		err := convertData(fieldValue, values[idx])
		if err != nil {
			modDb.setErr("First", err)
			return
		}
		structValue.FieldByName(field.name).Set(fieldValue.Elem())
	}
	return
}
func (modDb *ModDB) Count(count *int) (retModDb *ModDB) {
	retModDb = modDb
	if modDb.err != nil {
		return
	}

	if modDb.model == nil {
		retModDb.setErr("Count", errors.New("no model"))
		return
	}
	if modDb.model.tableName == "" {
		retModDb.setErr("Count", errors.New("no table"))
		return
	}

	if modDb.ds == nil {
		retModDb.setErr("Count", errors.New("no ds"))
		return
	}

	rows, err := modDb.execQry("select count(*) from "+modDb.model.tableName+" where "+modDb.where, modDb.whereVars...)
	if err != nil {
		modDb.setErr("Count", err)
		return
	}

	if rows.Next() == false {
		modDb.setErr("Count", errors.New("no rows returned"))
		return
	}
	var ct int
	err = rows.Scan(&ct)
	if err != nil {
		modDb.setErr("Count", err)
		return
	}
	*count = ct
	return
}

func (modDb *ModDB) Rows(columnList ...string) (*sql.Rows, error) {
	if modDb.err != nil {
		return nil, modDb.err
	}
	if modDb.model == nil {
		return nil, errors.New("no model")
	}
	if modDb.where == "" && modDb.primaryKeyValue == nil {
		return nil, errors.New("no where and primary key value")
	}
	if modDb.ds == nil {
		name := "default"
		ds := bds.GetDs(name)
		if ds == nil {
			modDb.setErr("Rows", errors.New("no ds found :"+name))
			return nil, errors.New("no ds found :" + name)
		}
		modDb.sqlDb = ds.SqlDb
		modDb.ds = ds
	}
	sqlCols := getModelQueryColumnList(modDb.model)
	if len(columnList) != 0 {
		sqlCols = ""
		for i, v := range columnList {
			if i != 0 {
				sqlCols = sqlCols + ","
			}
			sqlCols = sqlCols + v
		}
	}
	sqlStr := "select " + sqlCols + " from " + modDb.model.tableName
	if modDb.where != "" {
		sqlStr = sqlStr + " where " + modDb.where
		return modDb.execQry(sqlStr, modDb.whereVars...)
	} else {
		sqlStr = sqlStr + " where " + modDb.model.fieldMap[modDb.model.primaryKey].column + "=?"
		return modDb.execQry(sqlStr, modDb.primaryKeyValue)
	}
}
func (modDb *ModDB) Row(columnList ...string) (*sql.Row, error) {
	if modDb.err != nil {
		return nil, modDb.err
	}
	if modDb.model == nil {
		return nil, errors.New("no model")
	}
	if modDb.where == "" && modDb.primaryKeyValue == nil {
		return nil, errors.New("no where and primary key value")
	}
	if modDb.ds == nil {
		name := "default"
		ds := bds.GetDs(name)
		if ds == nil {
			modDb.setErr("Rows", errors.New("no ds found :"+name))
			return nil, errors.New("no ds found :" + name)
		}
		modDb.sqlDb = ds.SqlDb
		modDb.ds = ds
	}
	sqlCols := getModelQueryColumnList(modDb.model)
	if len(columnList) != 0 {
		sqlCols = ""
		for i, v := range columnList {
			if i != 0 {
				sqlCols = sqlCols + ","
			}
			sqlCols = sqlCols + v
		}
	}
	sqlStr := "select " + sqlCols + " from " + modDb.model.tableName
	if modDb.where != "" {
		sqlStr = sqlStr + " where " + modDb.where
		newSql, info := mdb.FirstRow(sqlStr, modDb.ds)
		if info != "" {
			return nil, errors.New(info)
		}
		return modDb.sqlDb.QueryRow(newSql, modDb.whereVars...), nil
	} else {
		sqlStr = sqlStr + " where " + modDb.model.fieldMap[modDb.model.primaryKey].column + "=?"
		return modDb.sqlDb.QueryRow(sqlStr, modDb.primaryKeyValue), nil
	}
}

func (modDb *ModDB) execQry(sqlStr string, args ...interface{}) (*sql.Rows, error) {
	if modDb.sqlDb != nil {
		return modDb.sqlDb.Query(sqlStr, args...)
	} else {
		return modDb.sqlTx.Query(sqlStr, args...)
	}
}

func getModelQueryColumnList(model *OrmModel) string {
	sqlCols := ""
	for _, ormField := range model.columnMap {
		if ormField.excluded {
			continue
		}
		if sqlCols == "" {
			sqlCols = ormField.column
		} else {
			sqlCols = sqlCols + "," + ormField.column
		}
	}
	return sqlCols
}
func convertData(dest reflect.Value, src any) error {
	dv := dest.Elem()
	switch dv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if src == nil {
			dv.SetInt(0)
		}
		s := asString(src)
		i64, err := strconv.ParseInt(s, 10, dv.Type().Bits())
		if err != nil {
			err = strconvErr(err)
			return fmt.Errorf("converting driver.Value type %T (%q) to a %s: %v", src, s, dv.Kind(), err)
		}
		dv.SetInt(i64)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if src == nil {
			dv.SetUint(0)
		}
		s := asString(src)
		u64, err := strconv.ParseUint(s, 10, dv.Type().Bits())
		if err != nil {
			err = strconvErr(err)
			return fmt.Errorf("converting driver.Value type %T (%q) to a %s: %v", src, s, dv.Kind(), err)
		}
		dv.SetUint(u64)
		return nil
	case reflect.Float32, reflect.Float64:
		if src == nil {
			dv.SetFloat(0)
			return nil
		}
		s := asString(src)
		f64, err := strconv.ParseFloat(s, dv.Type().Bits())
		if err != nil {
			err = strconvErr(err)
			return fmt.Errorf("converting driver.Value type %T (%q) to a %s: %v", src, s, dv.Kind(), err)
		}
		dv.SetFloat(f64)
		return nil
	case reflect.String:
		if src == nil {
			dv.SetString("")
			return nil
		}
		switch v := src.(type) {
		case string:
			dv.SetString(v)
			return nil
		case []byte:
			dv.SetString(string(v))
			return nil
		case time.Time:
			dv.SetString(v.Format(time.RFC3339Nano))
			return nil
		}
	}

	if dv.Type().String() == "time.Time" {
		if src == nil {
			return nil
		}
		switch v := src.(type) {
		case time.Time:
			dv.Set(reflect.ValueOf(v))
			return nil
		}
	}

	return fmt.Errorf("unsupported Scan, storing driver.Value type %T into type %T", src, dest)
}
func asString(src any) string {
	switch v := src.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	}
	rv := reflect.ValueOf(src)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(rv.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(rv.Uint(), 10)
	case reflect.Float64:
		return strconv.FormatFloat(rv.Float(), 'g', -1, 64)
	case reflect.Float32:
		return strconv.FormatFloat(rv.Float(), 'g', -1, 32)
	case reflect.Bool:
		return strconv.FormatBool(rv.Bool())
	}
	return fmt.Sprintf("%v", src)
}
func strconvErr(err error) error {
	if ne, ok := err.(*strconv.NumError); ok {
		return ne.Err
	}
	return err
}
