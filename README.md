# borm
## a simple go orm 

We are trying to create a very simple orm framework using the go language that can support traditional database operations and orm patterns. The focus of the project is to easily adapt to multiple databases.

## get borm
1. go get github.com/buffalo-x/borm
2. download zip file and unzip to directoty github.com/buffalo-x/borm

## start to use borm
1. _borm.Init()_ should be called in your project
2. connect to a database, here,Taking MySQL as an example.\
	_dsn :=  "root:XXX@tcp(local:3307)/db"\
	sqlDb, err := sql.Open("mysql", dsn)_

3. Add default database connection to datasource\
	_bds.SetDs("default", sqlDb, "root", "","127.0.0.1",3306, "dbName", "","mysql", "version info", "more info")_\
    Here, "mysql" is dsType\
    More connections can be added to the datasource, and they can be accessed by function\
	_bds.GetDs(name string)(*Datasource)_\
	DS with the name "default" should generally exist, otherwise it will increase code complexity. Because, if there is no explicit selection of ds, many functions default to using the ds with name of "default".

## use bmod method

1. Define the go structure of a database table 
	```json
	type TestTable struct {
		Id          int64  `borm:"primaryKey;autoIncrement"`
		Code        string `borm:"column:code;"`
		Name        string
		Pwd         string    `borm:"excluded"`
		ColInt      int       `borm:"excluded"`
		ColDecimal  float64   `borm:"type:decimal(20,2)"`
		ColDatetime string    `borm:"column:col_datetime;autoCreateTime"`
		ColDate     time.Time `borm:"excluded"`
		ColTs       time.Time `borm:"autoUpdateTime;"`
	}
	```
	The accepted datatype is among [string int int8 int16 int32 int64 float32 float64 bool time.Time] 
	
	borm tags :
	- primaryKey : this is a primary key column
	- autoIncrement : only primary key has this tag and field datatype must be int64
	- readonly : display data only
	- excluded : not handled
	- column : the real table column name in database
	- autoCreateTime : the time when record created, server's time 
	- autoUpdateTime : the time when record updated, server's time
	- type : datatype constraints 


2. Query table rows\
	_var rs []TestTable\
	modDb := bmod.Where("id>?",1).Query(&rs)_\
	If everything is normal, you can see the data in rs. You can also use\
	_err := bmod.Query(&rs).Error()_ to find if error occured or not.	

3. Query one table row\
	_var row TestTable\
	_modDb := bmod.Where("id=1").Query(&row)_

4. About func Where\
	_func (modDb *ModDB) Where(sqlStr string, args ...interface{}) (retModDb *ModDB)_,this func will create a *ModDB and set its where condition\
for example \
	_modDb := bmod.Where("code=? and name=?","z01","jim").Query(&row)_

5. Specify a datasource\
	_func Db(dsName ...string) (retModDb *ModDB)_\
	We can use it in this way.\
	 _modDb := bmod.Db("ds1").Query(&rs)_\
	This will choose a datasource with name of "ds1",then query.

6. Create a database record\
	_tb := &TestTable{Code: "aa01"}\
	modDb := bmod.Create(tb)_\
	You can also choose a datasource\
	_modDb := bmod.Db("ds1").Create(tb)_

7. Use Transaction\
	_func Tx(tx *sql.Tx, dsName ...string) (retModDb *ModDB)_\
	We can use it in this way.\
	Firstly, you should create a *sql.Tx object. Then,\
	_modDb := bmod.Tx(tx,"ds1").Create(tb)_\
	This will create record using tx

8. Save a database record\
	_tb := &TestTable{Code: "aa01",name:"mike"}\	
	modDb := bmod.Save(tb)_	

9. Delete a database record\
	_modDb := bmod.Delete(tb)_

10. Use Count func\
	_var ct int\
	modDb := bmod.Model(tb).Where("1=1").Count(&ct)_

11. Update one column\
_bmod.Model(tb).Update("name","cat")_\
This will update the name column using table primary key \
_bmod.Model(tb).Where("id=20").Update("name","cat")_\
This will update the name column where column id=20

12. Update more than one columns\
_bmod.Model(tb).UpdateColumns(bsql.CV{"name": "cat", "code": "z002"})_\
bsql.CV is a map[string]interface{}

13. Use bsql.Expr\
_bmod.Model(tb).Update("col_int",bsql.Expr("col_int+?", 1))_\
This means\ update test_table set col_int=col_int+1

14. Use sql.Rows as Output\
_func (modDb *ModDB) Rows(columnList ...string) (*sql.Rows, error)_\
Youcan use\
 _bmod.Model(tb).Where("1=1").Rows()\
or\
 _bmod.Model(tb).Where("1=1").Rows("name,code")

15. Use sql.Row as Output\
_func (modDb *ModDB) Row(columnList ...string) (*sql.Rows, error)_\
Youcan use\
 _bmod.Model(tb).Where("1=1").Row()\
or\
 _bmod.Model(tb).Where("id>?",2).Row("name,code")

	

## use bgen method
We also have a traditional way to deal with database.

1. Query table rows\
	_var rs bsql.Rows\
	genMod := bgen.Sql("select * from test_table where id>?", 1).Query(&rs)_\
	If everything is normal, you can see the data in rs. You can also use\
	_err := bgen.Sql("select * from test_table where id>?", 1).Query(&rs).Error()_\
     to find if error occured or not.

2. Query table row\
	_var rs bsql.Row\
	genMod := bgen.Sql("select * from test_table where id>?", 1).First(&rs)_

3. Specify a datasource\
	_genMod := bgen.Db("ds1").Sql("select * from test_table where id>?", 1).First(&rs)_\
	This will choose a datasource with name of "ds1"

4. Use Transaction
	_func Tx(tx *sql.Tx, dsName ...string) (retModDb *ModDB)_\
	We can use it in this way.\
	Firstly, you should create a *sql.Tx object. Then,\
	_genMod := bmod.Tx(tx,"ds1").Sql("select * from test_table where id>?", 1).First(&rs)_\
	This will query using tx

5. Query a value\
	_var ct int\
	bgen.Sql("select count(*) from test_table").Value(&ct)_\
	If your query return more then one rows, the first row will be used. \
	_bgen.Sql("select id from test_table").Value(&ct)_

6. Use sql.Rows as Output\
	_rows,err:=bgen.Sql("select * from test_table").Rows()_

7. Use sql.Row as Output\
	_row,err:=bgen.Sql("select * from test_table").Row()_

8. Exec sql \
	_genMod:= bgen.Sql("insert into test_table(name,code) values(?,?)","mike","z11").Exec()_\
	This will execute one insert sql. The following code has the same result.\
	_args := []interface{}{"mike", "z11"}\
	db1 := bgen.Sql("insert into test_table(name,code) values(?,?)",args).Exec()_

	We can create args like the following:\
	_args := [][]interface{}{{"susan", "z01"}, {"jose", "z02"}}\
	db1 := bgen.Sql("insert into test_table(name,code) values(?,?)",args).Exec()_\
	This will execute two insert sqls. 

9. Exec batch sqls\
	_func (genDb *GenDB) Batch(sqls []bsql.BatchSql) (retGenDb *GenDB)_\
	This will execute sqls in []bsql.BatchSql.\
	bsql.BatchSql is a struct\
	```json
	type BatchSql struct {
		Sql  string
		Args []interface{}
	}
	```
10. Create record\
	_bgen.Create("test_table", bsql.CV{"code": "33", "name": nil, "col_int": 1})_

11. Update one column\
	_bgen.Sql("id=?", 23).Update("test_table", "name", "mike")_\
	this will update one column

12. Update more columns\
	_bgen.Sql("id=?", 23).\
	UpdateColumns("test_table",\
			bsql.CV{"code": "33", "name": "cccs", "col_int":\ bsql.Expr("col_int+1")})_


## multiple databases
1. prepare sql option data\
You can organize sqls like this. **id-》dbType+dbVersion-》sql**
    ```json       
    sqlMap := map[string]map[string]string{
        "id1": {
		"mysql": "select id,now() from test_table",
		"mssql": "select id,getdate() from test_table",
		"default": "select id,now() from test_table",
		},
		"id2":{
			"mysql": "select id,left(name,5) from test_table",
			"mssql": "select id,substring(name,0,5) from test_table",
		},    
	}
    ```
- "id1" and "id2" are ids
- "mysql" and "mssql" are dbTypes
- "default" is the option when no matchs

2. add the option sql data to global map\    
    _mdb.AddDsSqlOptionMap("test",sqlMap)\
    Here "test" is a category.

3. use the option data\
    _var rs bsql.Rows\
    bgen.Sql("opt^^test,id1^^demo sql").Query(&rs)_\
    This will use the ds with name "default" to locate sql in group "id1" of category "test". \
    If no match found, 
    "default": "select id,now() from test_table" 
    will be returned.

    You can use like this to specify a datasource:\
    _bgen.Db("ds1").Sql("opt^^test,id1^^demo sql").Query(&rs)_   

    This will work in Where function. 

    



















