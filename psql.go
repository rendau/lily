package lily

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

func (dss *PsqlDynamicSqlSelectSt) Query() string {
	var q string
	var wq string

	if len(dss.With) > 0 {
		for _, w := range dss.With {
			if wq == `` {
				wq = `with `
			} else {
				wq += `, `
			}
			wq += w.Alias + ` as (` + w.Sql.Query() + `) `
		}
	}
	q = `select ` + dss.Select + ` `
	if dss.NeedRowNum {
		q += `, row_number() over(order by ` + dss.OrderBy + `) row_num `
	}
	q += `from ` + dss.From + ` `
	if dss.Where != `` {
		q += `where 1=1 ` + dss.Where + ` `
	}
	if dss.OrderBy != `` {
		q += `order by ` + dss.OrderBy + ` `
	}
	if dss.Offset != `` {
		q += `offset ` + dss.Offset + ` `
	}
	if dss.Limit != `` {
		q += `limit ` + dss.Limit + ` `
	}
	if dss.Json == "list" {
		return wq + PsqlJsonListQuery(q)
	} else if dss.Json == "row" {
		return wq + PsqlJsonRowQuery(q)
	}
	return wq + q
}

func (dss *PsqlDynamicSqlSelectSt) TotalCount() string {
	var q string
	if len(dss.With) > 0 {
		for _, w := range dss.With {
			if q != `` {
				q += `, `
			}
			q += `with ` + w.Alias + ` as (` + w.Sql.Query() + `) `
		}
	}
	q += `select count(*) from ` + dss.From + ` `
	if dss.Where != `` {
		q += `where 1=1 ` + dss.Where
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
