package bgen

import (
	"database/sql"
	"errors"
	"github.com/buffalo-x/borm/bsql"
	"github.com/buffalo-x/borm/mdb"
)

func (genDb *GenDB) Query() (retRows *bsql.Rows, retErr error) {

	retRows = nil
	retErr = nil
	if genDb.err != nil {
		retErr = genDb.err
		return
	}

	if genDb.sqlStr == "" {
		genDb.setErr("Query", errors.New("no sql"))
		retErr = genDb.err
		return
	}
	rs, err := genDb.execQry(genDb.sqlStr, genDb.sqlVars...)
	if err == nil {
		retRows, err = bsql.FetchRowsE(rs)
		if err != nil {
			genDb.setErr("Query", err)
			retErr = genDb.err
			return
		}
	} else {
		genDb.setErr("Query", err)
		retErr = genDb.err
	}
	return
}
func (genDb *GenDB) First() (retRow *bsql.Row, retErr error) {
	retRow = nil
	retErr = nil
	if genDb.err != nil {
		retErr = genDb.err
		return
	}
	if genDb.sqlStr == "" {
		genDb.setErr("First", errors.New("no sql"))
		retErr = genDb.err
		return
	}
	newSql, info := mdb.FirstRow(genDb.sqlStr, genDb.ds)
	if info != "" {
		genDb.setErr("First", errors.New(info))
		return
	}

	rs, err := genDb.execQry(newSql, genDb.sqlVars...)
	if err == nil {
		retRow, err = bsql.FetchFirstRowE(rs)
		if err != nil {
			genDb.setErr("First", err)
			retErr = genDb.err
			return
		}
		return
	} else {
		genDb.setErr("First", err)
		retErr = genDb.err
	}
	return
}
func (genDb *GenDB) Value() (retValue string, retErr error) {
	retValue = ""
	retErr = nil
	if genDb.err != nil {
		retErr = genDb.err
		return
	}
	if genDb.sqlStr == "" {
		genDb.setErr("Value", errors.New("no sql"))
		retErr = genDb.err
		return
	}
	newSql, info := mdb.FirstRow(genDb.sqlStr, genDb.ds)
	if info != "" {
		genDb.setErr("Value", errors.New(info))
		retErr = genDb.err
		return
	}
	rs, err := genDb.execQry(newSql, genDb.sqlVars...)
	if err == nil {
		brow, toErr := bsql.FetchFirstRowE(rs)
		if toErr != nil {
			genDb.setErr("Value", toErr)
			retErr = genDb.err
			return
		} else {
			retValue = brow.Data[0]
		}
	} else {
		genDb.setErr("Value", err)
		retErr = genDb.err
	}
	return
}

func (genDb *GenDB) Rows() (*sql.Rows, error) {
	if genDb.err != nil {
		return nil, genDb.err
	}
	if genDb.sqlStr == "" {
		return nil, errors.New("no sql")
	}
	return genDb.execQry(genDb.sqlStr, genDb.sqlVars...)
}
func (genDb *GenDB) Row() (*sql.Row, error) {
	if genDb.err != nil {
		return nil, genDb.err
	}
	if genDb.sqlStr == "" {
		return nil, errors.New("no sql")
	}
	newSql, info := mdb.FirstRow(genDb.sqlStr, genDb.ds)
	if info != "" {
		return nil, errors.New(info)
	}
	var rs *sql.Row
	if genDb.sqlDb != nil {
		rs = genDb.sqlDb.QueryRow(newSql, genDb.sqlVars...)
	} else {
		rs = genDb.sqlTx.QueryRow(newSql, genDb.sqlVars...)
	}
	return rs, nil
}

func (genDb *GenDB) execQry(sqlStr string, args ...interface{}) (*sql.Rows, error) {
	if genDb.sqlDb != nil {
		return genDb.sqlDb.Query(sqlStr, args...)
	} else {
		return genDb.sqlTx.Query(sqlStr, args...)
	}
}
