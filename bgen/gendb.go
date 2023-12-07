package bgen

import (
	"database/sql"
	"errors"
	"github.com/buffalo-x/borm/bds"
	"github.com/buffalo-x/borm/mdb"
)

type GenDB struct {
	sqlTx *sql.Tx
	sqlDb *sql.DB

	ds *bds.Datasource

	sqlStr  string
	sqlVars []interface{}

	err     error
	errFunc string

	RowsAffected int64
	LastInsertId int64
}

func (genDb *GenDB) clear() {
	genDb.sqlTx = nil
	genDb.sqlDb = nil

	genDb.ds = nil

	genDb.sqlStr = ""
	genDb.sqlVars = nil

	genDb.errFunc = ""
	genDb.err = nil
	genDb.RowsAffected = 0
	genDb.LastInsertId = 0
}
func (genDb *GenDB) setErr(errFunc string, err error) {
	genDb.errFunc = errFunc
	genDb.err = err
}
func (genDb *GenDB) GetErr() (error, string) {
	return genDb.err, genDb.errFunc
}
func (genDb *GenDB) Error() error {
	return genDb.err
}
func (genDb *GenDB) ErrorFunc() string {
	return genDb.errFunc
}
func (genDb *GenDB) Sql(sqlStr string, args ...interface{}) (retGenDb *GenDB) {
	retGenDb = genDb
	if genDb.err != nil {
		if genDb.ds != nil {
			genDb.setErr("", nil)
		} else {
			return
		}
	}
	sqlStr, err := mdb.SqlOption(genDb.ds, sqlStr)
	if err != nil {
		genDb.setErr("Sql", err)
		return
	}
	genDb.sqlStr = sqlStr
	genDb.sqlVars = args
	return
}
func Sql(sqlStr string, args ...interface{}) (genDb *GenDB) {
	genDb = &GenDB{}
	name := "default"
	ds := bds.GetDs(name)
	if ds == nil {
		genDb.setErr("Sql", errors.New("no ds found :"+name))
		return
	}
	genDb.sqlDb = ds.SqlDb
	genDb.ds = ds
	genDb.Sql(sqlStr, args...)
	return
}
func Db(dsName ...string) (genDb *GenDB) {
	genDb = &GenDB{}
	name := "default"
	if len(dsName) != 0 {
		name = dsName[0]
	}
	ds := bds.GetDs(name)
	if ds == nil {
		genDb.setErr("Db", errors.New("no ds found :"+name))
		return
	}
	genDb.sqlDb = ds.SqlDb
	genDb.ds = ds
	return
}
func Tx(tx *sql.Tx, dsName ...string) (genDb *GenDB) {
	genDb = &GenDB{}
	name := "default"
	if len(dsName) != 0 {
		name = dsName[0]
	}
	ds := bds.GetDs(name)
	if ds == nil {
		genDb.setErr("Tx", errors.New("no ds found :"+name))
	}
	genDb.sqlTx = tx
	genDb.ds = ds
	return
}
