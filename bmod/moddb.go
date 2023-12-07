package bmod

import (
	"database/sql"
	"errors"
	"github.com/buffalo-x/borm/bds"
	"github.com/buffalo-x/borm/mdb"
)

type ModDB struct {
	model           *OrmModel
	primaryKeyValue interface{}

	sqlTx *sql.Tx
	sqlDb *sql.DB
	ds    *bds.Datasource

	where     string
	whereVars []interface{}

	err          error
	errFunc      string
	RowsAffected int64
	LastInsertId int64
}

func (modDb *ModDB) clear() {
	modDb.sqlTx = nil
	modDb.sqlDb = nil
	modDb.ds = nil

	modDb.where = ""
	modDb.whereVars = nil

	modDb.errFunc = ""
	modDb.err = nil
	modDb.RowsAffected = 0
	modDb.LastInsertId = 0
}

func (modDb *ModDB) setErr(errFunc string, err error) {
	modDb.errFunc = errFunc
	modDb.err = err
}

func (modDb *ModDB) GetErr() (error, string) {
	return modDb.err, modDb.errFunc
}
func (modDb *ModDB) Error() error {
	return modDb.err
}
func (modDb *ModDB) ErrorFunc() string {
	return modDb.errFunc
}

func (modDb *ModDB) Tx(tx *sql.Tx, dsName ...string) (retModDb *ModDB) {
	retModDb = modDb
	if modDb.err != nil {
		return
	}
	if tx == nil {
		modDb.setErr("Tx", errors.New("tx is null"))
		return
	}
	name := "default"
	if len(dsName) != 0 {
		name = dsName[0]
	}
	ds := bds.GetDs(name)
	if ds == nil {
		modDb.setErr("Tx", errors.New("no ds found :"+name))
		return
	}
	modDb.sqlTx = tx
	modDb.ds = ds
	return
}
func (modDb *ModDB) Db(dsName ...string) (retModDb *ModDB) {
	retModDb = modDb
	if modDb.err != nil {
		return
	}
	name := "default"
	if len(dsName) != 0 {
		name = dsName[0]
	}
	ds := bds.GetDs(name)
	if ds == nil {
		modDb.setErr("Db", errors.New("no ds found :"+name))
		return modDb
	}
	modDb.sqlDb = ds.SqlDb
	modDb.ds = ds
	return modDb
}

func Db(dsName ...string) (retModDb *ModDB) {
	retModDb = &ModDB{}
	name := "default"
	if len(dsName) != 0 {
		name = dsName[0]
	}
	ds := bds.GetDs(name)
	if ds == nil {
		retModDb.setErr("Db", errors.New("no ds found :"+name))
		return
	}
	retModDb.sqlDb = ds.SqlDb
	retModDb.ds = ds
	return
}
func Tx(tx *sql.Tx, dsName ...string) (retModDb *ModDB) {
	retModDb = &ModDB{}
	name := "default"
	if len(dsName) != 0 {
		name = dsName[0]
	}
	ds := bds.GetDs(name)
	if ds == nil {
		retModDb.setErr("Db", errors.New("no ds found :"+name))
		return
	}
	retModDb.sqlTx = tx
	retModDb.ds = ds
	return
}

func Where(sqlStr string, args ...interface{}) (retModDb *ModDB) {
	retModDb = &ModDB{}
	ds := bds.GetDs("default")
	if ds == nil {
		retModDb.setErr("Where", errors.New("no ds found :"+"default"))
		return
	}
	retModDb.sqlDb = ds.SqlDb
	retModDb.ds = ds
	retModDb.Where(sqlStr, args...)
	return
}
func (modDb *ModDB) Where(sqlStr string, args ...interface{}) (retModDb *ModDB) {
	retModDb = modDb
	if modDb.err != nil {
		return
	}

	if modDb.ds == nil {
		ds := bds.GetDs("default")
		if ds == nil {
			retModDb.setErr("Where", errors.New("no ds found :"+"default"))
			return
		}
		modDb.ds = ds
		modDb.sqlDb = ds.SqlDb
	}

	sqlStr, err := mdb.SqlOption(modDb.ds, sqlStr)
	if err != nil {
		modDb.setErr("Sql", err)
		return
	}
	modDb.where = sqlStr
	modDb.whereVars = args
	return
}
