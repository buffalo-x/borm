package bmod

import (
	"errors"
	"go/ast"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

type OrmModel struct {
	name       string
	primaryKey string //目前先支持单一主键或者无主键，多主键日后考虑
	fieldMap   map[string]*OrmModelField
	columnMap  map[string]*OrmModelField
	tableName  string
}
type OrmModelField struct {
	name              string
	goType            reflect.Type
	goTypeName        string
	tag               string
	column            string
	dbType            string
	primaryKey        bool
	excluded          bool
	readonly          bool
	autoIncrement     bool
	autoCreateTime    bool
	autoUpdateTime    bool
	decimalConstraint struct {
		m, d  int
		valid bool
	}
}

var ormModelMap = make(map[string]*OrmModel)
var ormModelLock sync.RWMutex
var acceptedDataTypeForOrmModel = map[string]bool{"string": true,
	"int": true, "int64": true, "int32": true, "int16": true, "int8": true,
	"float32": true, "float64": true, "bool": true,
	"time.Time": true}

func modelStructPtr(structPtr interface{}) (*OrmModel, error) {
	ptrType := reflect.TypeOf(structPtr)

	if ptrType.Kind() != reflect.Pointer {
		return nil, errors.New("parameter is not a pointer")
	}
	objType := ptrType.Elem() //struct
	if objType.Kind() != reflect.Struct {
		return nil, errors.New("parameter is not a struct pointer")
	}
	return modelStructType(objType)
}

func modelStructType(objType reflect.Type) (*OrmModel, error) {

	if objType.Kind() != reflect.Struct {
		return nil, errors.New("parameter is not a struct")
	}

	structTypeName := strings.Replace(objType.String(), "*", "", -1)
	ormModel := &OrmModel{name: structTypeName, fieldMap: make(map[string]*OrmModelField),
		columnMap: make(map[string]*OrmModelField)}

	for i := 0; i < objType.NumField(); i++ {
		field := objType.Field(i)
		if !ast.IsExported(field.Name) {
			continue
		}
		goType := field.Type
		goTypeName := field.Type.String()
		if _, ok := acceptedDataTypeForOrmModel[field.Type.String()]; !ok {
			return nil, errors.New(field.Name + " has a wrong datatype " + goTypeName)
		}

		fieldTag := field.Tag.Get("borm")
		ormField := &OrmModelField{name: field.Name, goType: goType, goTypeName: goTypeName, tag: fieldTag}
		ormModel.fieldMap[field.Name] = ormField

		retErr := analyzeTag(fieldTag, ormModel, ormField)
		if retErr != "" {
			return nil, errors.New(retErr)
		}

		if ormField.excluded {
			if ormField.primaryKey {
				return nil, errors.New("primaryKey can't be excluded")
			}
		} else { //set column name in database table
			if ormField.column == "" {
				ormField.column = camelStrToUnderline(ormField.name)
			}
			if _, ok := ormModel.columnMap[ormField.column]; ok {
				return nil, errors.New("more than one column " + ormField.column)
			}
			ormModel.columnMap[ormField.column] = ormField
		}

		if ormField.primaryKey {
			if !strings.HasPrefix(ormField.goTypeName, "int") && !strings.HasPrefix(ormField.goTypeName, "string") {
				return nil, errors.New("primaryKey column needs datatype int or string")
			}
		}
		if ormField.autoIncrement {
			if ormField.goTypeName != "int64" {
				return nil, errors.New("column with tag 'autoIncrement' needs datatype int64")
			}
			if !ormField.primaryKey {
				return nil, errors.New("column with tag 'autoIncrement' must be primaryKey")
			}
		}
	}

	nameStrs := strings.Split(structTypeName, ".")
	ormModel.tableName = camelStrToUnderline(nameStrs[len(nameStrs)-1])

	return ormModel, nil
}

func Model(structPtr interface{}) (modDb *ModDB) {
	modDb = &ModDB{}

	ptrType := reflect.TypeOf(structPtr)
	structTypeName := strings.Replace(ptrType.String(), "*", "", -1)
	ormModelLock.RLock()
	model, ok := ormModelMap[structTypeName]
	ormModelLock.RUnlock()
	if ok {
		modDb.model = model
		modDb.model.tableName = model.tableName
		return
	}

	model, err := modelStructPtr(structPtr)
	if err != nil {
		modDb.setErr("Model", err)
		return
	}
	//add to map
	ormModelLock.Lock()
	defer ormModelLock.Unlock()
	ormModelMap[model.name] = model

	modDb.model = model
	modDb.model.tableName = model.tableName

	if model.primaryKey != "" {
		modDb.primaryKeyValue = reflect.ValueOf(structPtr).Elem().FieldByName(model.primaryKey).Interface()
	}

	return
}

func analyzeTag(fieldTag string, ormModel *OrmModel, ormField *OrmModelField) string {
	fieldTag = strings.Trim(fieldTag, " ")
	if fieldTag == "" {
		return ""
	}
	tagSecion := strings.Split(fieldTag, ";")
	for i := 0; i < len(tagSecion); i++ {
		section := strings.Trim(tagSecion[i], " ")
		if strings.ReplaceAll(section, " ", "") == "" {
			continue
		}
		if section == "primaryKey" {
			if ormModel.primaryKey != "" {
				return "only one primaryKey accepted"
			}
			ormModel.primaryKey = ormField.name
			ormField.primaryKey = true
		} else if section == "excluded" {
			ormField.excluded = true
		} else if section == "readonly" {
			ormField.readonly = true
		} else if section == "autoIncrement" {
			ormField.autoIncrement = true
		} else if section == "autoCreateTime" {
			ormField.autoCreateTime = true
		} else if section == "autoUpdateTime" {
			ormField.autoUpdateTime = true
		} else {
			idx := strings.Index(section, ":")
			if idx <= 0 || idx >= len(section)-1 {
				return "model" + ormModel.name + "field " + ormField.name + " tag error" + " " + section
			}
			attrName := strings.Trim(section[:idx], " ")
			attrVal := strings.Trim(section[idx+1:], " ")
			if attrName == "column" {
				ormField.column = strings.ReplaceAll(attrVal, "", "")
			} else if attrName == "type" {
				ormField.dbType = attrVal
				retErr := analyzeType(attrVal, ormField)
				if retErr != "" {
					return ormField.name + " " + retErr
				}
			} else {
				return "model" + ormModel.name + "field " + ormField.name + " tag error" + " " + section
			}
		}
	}
	return ""
}
func analyzeType(attrVal string, field *OrmModelField) string {
	attrVal = strings.ReplaceAll(attrVal, " ", "")
	if strings.HasPrefix(attrVal, "decimal") {
		//数据库的decimal类型只能对应model中的string float 类型
		if field.goTypeName != "string" && !strings.HasPrefix(field.goTypeName, "float") &&
			!strings.HasPrefix(field.goTypeName, "int") {
			return "datatype mismatch:" + field.goTypeName + " " + attrVal
		}
		valStr := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(attrVal, "decimal", ""), "(", ""), ")", ""), " ", "")
		valSections := strings.Split(valStr, ",")
		if len(valSections) != 2 {
			return "type syntax error " + attrVal
		}
		m, e := strconv.Atoi(valSections[0])
		if e != nil {
			return "type syntax error " + attrVal
		}
		if m <= 0 {
			return "type syntax error " + attrVal
		}
		d, e := strconv.Atoi(valSections[1])
		if e != nil {
			return "type syntax error " + attrVal
		}
		if d <= 0 {
			return "type syntax error " + attrVal
		}
		field.decimalConstraint.m = m
		field.decimalConstraint.d = d
		field.decimalConstraint.valid = true
	} else {
		return "unknow type " + attrVal
	}
	return ""
}
func camelStrToUnderline(camelStr string) (ret string) {
	ret = ""
	for i := 0; i < len(camelStr); i++ {
		currChar := camelStr[i : i+1]
		lowerChar := strings.ToLower(currChar)
		if lowerChar != currChar {
			if ret == "" {
				ret = ret + lowerChar
			} else {
				ret = ret + "_" + lowerChar
			}
		} else {
			ret = ret + lowerChar
		}
	}
	return
}
