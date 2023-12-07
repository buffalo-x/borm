package bgen

import (
	"borm/bsql"
	"borm/mdb"
	"database/sql"
	"errors"
)

func (genDb *GenDB) Query(rows *bsql.Rows) (retGenDb *GenDB) {
	retGenDb = genDb
	if genDb.err != nil {
		return
	}
	if genDb.sqlStr == "" {
		genDb.setErr("Query", errors.New("no sql"))
		return
	}
	rs, err := genDb.execQry(genDb.sqlStr, genDb.sqlVars...)
	if err == nil {
		toErr := bsql.ToBsqlRows(rs, rows)
		if toErr != nil {
			genDb.setErr("Query", toErr)
		}
	} else {
		genDb.setErr("Query", err)
	}
	return
}
func (genDb *GenDB) First(row *bsql.Row) (retGenDb *GenDB) {
	retGenDb = genDb
	if genDb.err != nil {
		return
	}
	if genDb.sqlStr == "" {
		genDb.setErr("First", errors.New("no sql"))
		return
	}
	newSql, info := mdb.FirstRow(genDb.sqlStr, genDb.ds)
	if info != "" {
		genDb.setErr("First", errors.New(info))
		return
	}

	rs, err := genDb.execQry(newSql, genDb.sqlVars...)
	if err == nil {
		brow, toErr := bsql.FetchFirstRowE(rs)
		if toErr != nil {
			genDb.setErr("First", toErr)
		} else {
			*row = *brow
		}
	} else {
		genDb.setErr("First", err)
	}
	return
}
func (genDb *GenDB) Value(value *string) (retGenDb *GenDB) {
	retGenDb = genDb
	if genDb.err != nil {
		return
	}
	if genDb.sqlStr == "" {
		genDb.setErr("Value", errors.New("no sql"))
		return
	}
	newSql, info := mdb.FirstRow(genDb.sqlStr, genDb.ds)
	if info != "" {
		genDb.setErr("Value", errors.New(info))
		return
	}
	rs, err := genDb.execQry(newSql, genDb.sqlVars...)
	if err == nil {
		brow, toErr := bsql.FetchFirstRowE(rs)
		if toErr != nil {
			genDb.setErr("Value", toErr)
		} else {
			*value = brow.Data[0]
		}
	} else {
		genDb.setErr("Value", err)
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
