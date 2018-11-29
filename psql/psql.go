package psql

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/rendau/lily"
	"strconv"
	"strings"
)

type DynamicSqlSelectSt struct {
	With       []DynamicSqlWithSt
	Select     string
	NeedRowNum bool
	From       string
	Where      string
	GroupBy    string
	OrderBy    string
	Offset     string
	Limit      string
	Json       string
	Args       []interface{}
}

type DynamicSqlWithSt struct {
	Alias string
	Sql   *DynamicSqlSelectSt
}

type DynamicSqlUpdateSt struct {
	Table string
	Set   string
	Where string
	Args  []interface{}
}

type DynamicSqlInsertSt struct {
	Table     string
	Fields    string
	Values    string
	Returning string
	Args      []interface{}
}

func (ds *DynamicSqlSelectSt) Query() string {
	var q string
	var wq string

	if len(ds.OrderBy) > 0 && ds.OrderBy[0] == ',' {
		ds.OrderBy = ds.OrderBy[1:]
	}

	if len(ds.With) > 0 {
		for _, w := range ds.With {
			if wq == `` {
				wq = `with `
			} else {
				wq += `, `
			}
			wq += w.Alias + ` as (` + w.Sql.Query() + `) `
		}
	}
	q = `select ` + ds.Select + ` `
	if ds.NeedRowNum {
		q += `, row_number() over(order by ` + ds.OrderBy + `) row_num `
	}
	if ds.From != `` {
		q += `from ` + ds.From + ` `
		if ds.Where != `` {
			q += `where `
			if strings.HasPrefix(strings.TrimLeft(ds.Where, " "), "and ") {
				q += `1=1 `
			}
			q += ds.Where + ` `
		}
		if ds.GroupBy != `` {
			q += `group by ` + ds.GroupBy + ` `
		}
		if ds.OrderBy != `` {
			q += `order by ` + ds.OrderBy + ` `
		}
		if ds.Offset != `` {
			q += `offset ` + ds.Offset + ` `
		}
		if ds.Limit != `` {
			q += `limit ` + ds.Limit + ` `
		}
	}
	if ds.Json == "list" {
		return wq + JsonListQuery(q)
	} else if ds.Json == "row" {
		return wq + JsonRowQuery(q)
	}
	return wq + q
}

func (ds *DynamicSqlSelectSt) TotalCount() string {
	var q string
	if len(ds.With) > 0 {
		for _, w := range ds.With {
			if q == `` {
				q = `with `
			} else {
				q += `, `
			}
			q += w.Alias + ` as (` + w.Sql.Query() + `) `
		}
	}
	q += `select count(*) from ` + ds.From + ` `
	if ds.Where != `` {
		q += `where `
		if strings.HasPrefix(strings.TrimLeft(ds.Where, " "), "and ") {
			q += `1=1 `
		}
		q += ds.Where + ` `
	}
	if ds.GroupBy != `` {
		q += `group by ` + ds.GroupBy + ` `
	}
	return q
}

func (ds *DynamicSqlUpdateSt) Query() string {
	var q string
	q = `update ` + ds.Table + ` set `
	if ds.Set[0] == ',' {
		q += ds.Set[1:]
	} else {
		q += ds.Set
	}
	if ds.Where != `` {
		q += ` where ` + ds.Where
	}
	return q
}

func (ds *DynamicSqlInsertSt) Query() string {
	var q string
	q = `insert into ` + ds.Table + ` (`
	if ds.Fields[0] == ',' {
		q += ds.Fields[1:]
	} else {
		q += ds.Fields
	}
	q += `) values (`
	if ds.Values[0] == ',' {
		q += ds.Values[1:]
	} else {
		q += ds.Values
	}
	q += `)`
	if ds.Returning != `` {
		q += ` returning ` + ds.Returning
	}
	return q
}

//func PanicRecoverTxnRollback(txn *sqlx.Tx) {
//	if err := recover(); err != nil {
//		txn.Rollback()
//	}
//}

func DeferHandleTxn(txn *sqlx.Tx) {
	if p := recover(); p != nil {
		txn.Rollback()
		panic(p)
	} else {
		err := txn.Commit()
		if err != sql.ErrTxDone {
			panic(err)
		}
	}
}

func TransactionWithTimezone(db *sqlx.DB, tzHOffset int) (error, *sqlx.Tx) {
	tx, err := db.Beginx()
	lily.ErrPanicSilent(err)

	_, err = tx.Exec(`set local time zone ` + strconv.Itoa(tzHOffset))
	if err != nil {
		tx.Rollback()
		return err, nil
	}

	return nil, tx
}

func JsonListQuery(q string) string {
	return `(select coalesce(array_to_json(array_agg(row_to_json(sq.*))), '[]')
		from (` + q + `) sq)`
}

func JsonRowQuery(q string) string {
	return `(select row_to_json(sq.*) from (` + q + `) sq)`
}
