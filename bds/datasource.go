package bds

import (
	"database/sql"
	"sync"
)

type Datasource struct {
	Name          string
	SqlDb         *sql.DB
	User, Pwd, Ip string
	Port          int ``
	Db            string
	DsType        string
	DsVersion     string
	Parameters    string

	Info string
}

var defaultDatasource *Datasource
var datasourceMap = make(map[string]*Datasource)
var datasourceMapLock sync.RWMutex

func GetSqlDb(name ...string) *sql.DB {
	dsName := "default"
	if len(name) != 0 {
		dsName = name[0]
	}
	if dsName == "default" {
		if defaultDatasource != nil {
			return defaultDatasource.SqlDb
		} else {
			return nil
		}
	}
	datasourceMapLock.RLock()
	defer datasourceMapLock.RUnlock()
	if db, ok := datasourceMap[dsName]; ok {
		return db.SqlDb
	}
	return nil
}
func GetDs(name ...string) *Datasource {
	dsName := "default"
	if len(name) != 0 {
		dsName = name[0]
	}
	if dsName == "default" {
		if defaultDatasource != nil {
			ds := &Datasource{Name: defaultDatasource.Name, SqlDb: defaultDatasource.SqlDb, User: defaultDatasource.User,
				Pwd: "", Ip: defaultDatasource.Ip, Port: defaultDatasource.Port,
				Db: defaultDatasource.Db, DsType: defaultDatasource.DsType, DsVersion: defaultDatasource.DsVersion, Info: defaultDatasource.Info}
			return ds
		} else {
			return nil
		}
	}
	datasourceMapLock.RLock()
	defer datasourceMapLock.RUnlock()
	if ds, ok := datasourceMap[dsName]; ok {
		retDs := &Datasource{Name: ds.Name, SqlDb: ds.SqlDb, User: ds.User,
			Pwd: "", Ip: ds.Ip, Port: ds.Port,
			Db: ds.Db, DsType: ds.DsType, DsVersion: ds.DsVersion, Info: ds.Info}
		return retDs
	}
	return nil
}
func SetDs(name string, sqlDb *sql.DB, user, pwd, ip string, port int, db, parameters,
	dstype, dsversion string, info string) {
	ds := &Datasource{Name: name, SqlDb: sqlDb, User: user, Pwd: pwd, Ip: ip, Port: port, Db: db,
		DsType: dstype, DsVersion: dsversion}
	ds.Info = info
	datasourceMapLock.Lock()
	datasourceMap[name] = ds
	datasourceMapLock.Unlock()
	if name == "default" {
		defaultDatasource = ds
	}
}
