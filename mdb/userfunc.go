package mdb

import (
	"borm/bds"
	"reflect"
)

func FirstRow(sqlStr string, ds *bds.Datasource) (string, string) {
	m, ok := adMethodMap[ds.DsType+"^FirstRow"]
	if !ok {
		return sqlStr, ""
	}
	params := make([]reflect.Value, 3)
	params[0] = reflect.ValueOf(sqlStr)
	params[1] = reflect.ValueOf(ds.DsType)
	params[2] = reflect.ValueOf(ds.DsVersion)
	result := m.Call(params)
	if len(result) != 2 {
		return "", "FirstRow return parameter error"
	}
	if result[0].Kind() != reflect.String {
		return "", "FirstRow return parameter 1 is not a string"
	}
	if result[1].Kind() != reflect.String {
		return "", "FirstRow return parameter 2 is not a string"
	}
	return result[0].Interface().(string), result[1].Interface().(string)
}
