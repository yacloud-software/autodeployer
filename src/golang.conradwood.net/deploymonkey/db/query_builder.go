package db

import (
	"fmt"
	"sort"
	"strings"
)

type QP map[string]interface{}

type Query struct {
	and_clauses []string
	paras       map[string]interface{}
	max         uint32
	order       string
}

/*
create a new query builder
*/
func NewQuery() *Query {
	return &Query{paras: make(map[string]interface{})}
}

// named parameter like so: "foo = :bar:" and map{"bar":"none"}
func (q *Query) Add(and_clause string, paras map[string]interface{}) {
	q.and_clauses = append(q.and_clauses, and_clause)
	for k, _ := range q.paras {
		_, b := paras[k]
		if b {
			panic(fmt.Sprintf("parameter \"%s\" set multiple times", k))
		}
	}
	for k, v := range paras {
		q.paras[k] = v
	}
}

// set an order by clause
func (q *Query) OrderBy(fieldname string) {
	q.order = fieldname
}

// set a limit on how many rows are returned
func (q *Query) Limit(max uint32) {
	q.max = max
}

// returns a postgres compatible string
func (q *Query) ToPostgres() (string, []interface{}) {
	var keys []string
	for k, _ := range q.paras {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	//build the final query
	deli := ""
	final_clause := ""
	for _, clause := range q.and_clauses {
		final_clause = final_clause + deli + "(" + clause + ")"
		deli = " AND "
	}
	// now replace the keys
	var paras []interface{}
	for pos, key := range keys {
		paras = append(paras, q.paras[key])
		final_clause = strings.ReplaceAll(final_clause, ":"+key+":", fmt.Sprintf("$%d", (pos+1)))
	}
	if q.order != "" {
		final_clause = final_clause + fmt.Sprintf(" ORDER BY "+q.order)
	}
	if q.max != 0 {
		final_clause = final_clause + fmt.Sprintf(" LIMIT %d", q.max)
	}
	return final_clause, paras
}

// add an equal comparison to the query
func (q *Query) AddEqual(field string, value interface{}) {
	vname := fmt.Sprintf("field_%s", field)
	q.Add(field+" = :"+vname+":", QP{vname: value})
}
