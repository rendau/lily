package lily

import "strings"

type PsqlDynamicSqlSelectSt struct {
	With       []PsqlDynamicSqlWithSt
	Select     string
	NeedRowNum bool
	From       string
	Where      string
	OrderBy    string
	Offset     string
	Limit      string
	Json       string
	Args       []interface{}
}

type PsqlDynamicSqlWithSt struct {
	Alias string
	Sql   *PsqlDynamicSqlSelectSt
}

type PsqlDynamicSqlUpdateSt struct {
	Table string
	Set   string
	Where string
	Args  []interface{}
}

type PsqlDynamicSqlInsertSt struct {
	Table     string
	Fields    string
	Values    string
	Returning string
	Args      []interface{}
}

func (ds *PsqlDynamicSqlSelectSt) Query() string {
	var q string
	var wq string

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
	q += `from ` + ds.From + ` `
	if ds.Where != `` {
		q += `where `
		if strings.HasPrefix(strings.TrimLeft(ds.Where, " "), "and ") {
			q += `1=1 `
		}
		q += ds.Where + ` `
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
	if ds.Json == "list" {
		return wq + PsqlJsonListQuery(q)
	} else if ds.Json == "row" {
		return wq + PsqlJsonRowQuery(q)
	}
	return wq + q
}

func (ds *PsqlDynamicSqlSelectSt) TotalCount() string {
	var q string
	if len(ds.With) > 0 {
		for _, w := range ds.With {
			if q != `` {
				q += `, `
			}
			q += `with ` + w.Alias + ` as (` + w.Sql.Query() + `) `
		}
	}
	q += `select count(*) from ` + ds.From + ` `
	if ds.Where != `` {
		q += `where 1=1 ` + ds.Where
	}
	return q
}

func (ds *PsqlDynamicSqlUpdateSt) Query() string {
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

func (ds *PsqlDynamicSqlInsertSt) Query() string {
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

func PsqlJsonListQuery(q string) string {
	return `(select coalesce(array_to_json(array_agg(row_to_json(sq.*))), '[]')
		from (` + q + `) sq)`
}

func PsqlJsonRowQuery(q string) string {
	return `(select row_to_json(sq.*) from (` + q + `) sq)`
}
