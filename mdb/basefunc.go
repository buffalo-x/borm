package mdb

import (
	"errors"
	"github.com/buffalo-x/borm/bds"
	"github.com/buffalo-x/borm/mdb/mssql"
	"github.com/buffalo-x/borm/mdb/mysql"
	"github.com/buffalo-x/borm/mdb/postgresql"
	"reflect"
	"strings"
)

// dsSqlOptionMap：   category-》id-》dbType+dbVersion-》sql

// if default exists and no match found, choose default
/* 举例如下
"test":{
  "id1": {
    "mysql": "select * from test_table",
    "mssql": "select * from test_table",
    "default": "select id,now() from test_table"
  },
  "id2":{
    "mysql": "select id,now() from test_table",
    "mssql": "select id,now() from test_table",
    "default": "select id,now() from test_table"
  }
}
*/
var dsSqlOptionMap = make(map[string]map[string]map[string]string)
var adMethodMap = make(map[string]reflect.Value)

func Init() {
	if mysql.MdbFuncs == nil {
		mysql.MdbFuncs = &mysql.FuncStruct{}
		mVal := reflect.ValueOf(mysql.MdbFuncs)
		mType := mVal.Type()
		a := mVal.NumMethod()
		_ = a
		for i := 0; i < mVal.NumMethod(); i++ {
			mName := "mysql^" + mType.Method(i).Name
			adMethodMap[mName] = mVal.Method(i)
		}
	}
	if mssql.MdbFuncs == nil {
		mssql.MdbFuncs = &mssql.FuncStruct{}
		mVal := reflect.ValueOf(mssql.MdbFuncs)
		mType := mVal.Type()
		for i := 0; i < mVal.NumMethod(); i++ {
			mName := "mssql^" + mType.Method(i).Name
			adMethodMap[mName] = mVal.Method(i)
		}
	}
	if postgresql.MdbFuncs == nil {
		postgresql.MdbFuncs = &postgresql.FuncStruct{}
		mVal := reflect.ValueOf(postgresql.MdbFuncs)
		mType := mVal.Type()
		for i := 0; i < mVal.NumMethod(); i++ {
			mName := "oracle^" + mType.Method(i).Name
			adMethodMap[mName] = mVal.Method(i)
		}
	}
}
func AddDsSqlOptionMap(category string, optionMap map[string]map[string]string) error {
	if _, ok := dsSqlOptionMap[category]; ok {
		return errors.New("category duplicated")
	}
	dsSqlOptionMap[category] = optionMap
	return nil
}
func GetDsSqlOptionE(dsType, dsVersion string, category, key string) (string, error) {
	catMap, ok := dsSqlOptionMap[category]
	if !ok {
		return "", errors.New("no category " + category)
	}
	keyMap, ok := catMap[key]
	if !ok {
		return "", errors.New("no key " + key + " in category " + category)
	}
	str, ok := keyMap[dsType+" "+dsVersion]
	if ok {
		return str, nil
	}
	str, ok = keyMap[dsType]
	if ok {
		return str, nil
	}
	str, ok = keyMap["default"]
	if ok {
		return str, nil
	}
	return "", errors.New("no match for " + dsType + " ^ " + dsVersion + " ^ " + category + " ^ " + key)
}
func GetDsSqlOption(dsType, dsVersion string, category, key string) string {
	catMap, ok := dsSqlOptionMap[category]
	if !ok {
		return ""
	}
	keyMap, ok := catMap[key]
	if !ok {
		return ""
	}
	str, ok := keyMap[dsType+" "+dsVersion]
	if ok {
		return str
	}
	str, ok = keyMap[dsType]
	if ok {
		return str
	}
	str, ok = keyMap["default"]
	if ok {
		return str
	}
	return ""
}

func SqlOption(ds *bds.Datasource, sqlStr string) (string, error) {
	sqlStr = strings.TrimLeft(sqlStr, " ")
	if strings.Index(sqlStr, "opt^^") != 0 {
		return sqlStr, nil
	}
	sqlStr = strings.Replace(sqlStr, "opt^^", "", -1)
	if idx := strings.Index(sqlStr, "^^"); idx > 0 {
		sqlStr = sqlStr[:idx]
	}
	strs := strings.Split(sqlStr, ",")
	if len(strs) != 2 {
		return "", errors.New("sql option syntax error: " + sqlStr)
	}
	return GetDsSqlOptionE(ds.DsType, ds.DsVersion, strs[0], strs[1])
}
