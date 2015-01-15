package main

import (
	"text/template"
)

var header string = `// Code generated by ModelQ
// {{.TableName}}.go contains model for the database table [{{.DbName}}.{{.TableName}}]

package {{.PkgName}}

import (
	"encoding/json"
	"encoding/gob"
	"fmt"
	"strings"
	"github.com/mijia/modelq/gmq"
	"database/sql"
	{{if .ImportTime}}"time"{{end}}
)
`

var modelStruct string = `type {{.Name}} struct {
	{{range .Fields}}{{.Name}} {{.Type}} {{.JsonMeta}}{{if .Comment}} // {{.Comment}}{{end}}
	{{end}}
}
`

var objApi string = `
// Start of the {{.Name}} APIs.

func (obj {{.Name}}) String() string {
	if data, err := json.Marshal(obj); err != nil {
		{{if .PrimaryField}}return fmt.Sprintf("<{{.Name}} {{.PrimaryField.Name}}=%d>", obj.{{.PrimaryField.Name}}){{else}}return fmt.Sprintf("<{{.Name}}>"){{end}}
	} else {
		return string(data)
	}
}

func (obj {{.Name}}) Insert(dbtx gmq.DbTx) ({{.Name}}, error) {
	{{if .HasAutoIncrementPrimaryKey}}if result, err := {{.Name}}Objs.Insert(obj).Run(dbtx); err != nil {
		return obj, err
	}else {
		if id, err := result.LastInsertId(); err != nil {
			return obj, err
		} else {
			obj.Id = {{if eq .PrimaryField.Type "int64"}}id{{else}}{{.PrimaryField.Type}}(id){{end}}
			return obj, err
		}
	}{{else}}_, err := {{.Name}}Objs.Insert(obj).Run(dbtx)
	return obj, err{{end}}
}

func (obj {{.Name}}) Update(dbtx gmq.DbTx) (int64, error) {
	{{if .PrimaryField}}fields := []string{ {{.UpdatableFields}} }
	filter := {{.Name}}Objs.Filter{{.PrimaryField.Name}}("=", obj.{{.PrimaryField.Name}})
	if result, err := {{.Name}}Objs.Update(obj, fields...).Where(filter).Run(dbtx); err != nil {
		return 0, err
	} else {
		return result.RowsAffected()
	}{{else}}return 0, gmq.ErrNoPrimaryKeyDefined{{end}}
}

func (obj {{.Name}}) Delete(dbtx gmq.DbTx) (int64, error) {
	{{if .PrimaryField}}filter := {{.Name}}Objs.Filter{{.PrimaryField.Name}}("=", obj.{{.PrimaryField.Name}})
	if result, err := {{.Name}}Objs.Delete().Where(filter).Run(dbtx); err != nil {
		return 0, err
	} else {
		return result.RowsAffected()
	}{{else}}return 0, gmq.ErrNoPrimaryKeyDefined{{end}}
}
`

var queryApi string = `
// Start of the inner Query Api

type _{{.Name}}Query struct {
	gmq.Query
}

func (q _{{.Name}}Query) Where(f gmq.Filter) _{{.Name}}Query {
	q.Query = q.Query.Where(f)
	return q
}

func (q _{{.Name}}Query) OrderBy(by ...string) _{{.Name}}Query {
	tBy := make([]string, 0, len(by))
	for _, b := range by {
		sortDir := ""
		if b[0] == '-' || b[0] == '+' {
			sortDir = string(b[0])
			b = b[1:]
		}
		if col, ok := {{.Name}}Objs.fcMap[b]; ok {
			tBy = append(tBy, sortDir+col)
		}
	}
	q.Query = q.Query.OrderBy(tBy...)
	return q
}

func (q _{{.Name}}Query) GroupBy(by ...string) _{{.Name}}Query {
	tBy := make([]string, 0, len(by))
	for _, b := range by {
		if col, ok := {{.Name}}Objs.fcMap[b]; ok {
			tBy = append(tBy, col)
		}
	}
	q.Query = q.Query.GroupBy(tBy...)
	return q
}

func (q _{{.Name}}Query) Limit(offsets ...int64) _{{.Name}}Query {
	q.Query = q.Query.Limit(offsets...)
	return q
}

func (q _{{.Name}}Query) Page(number, size int) _{{.Name}}Query {
	q.Query = q.Query.Page(number, size)
	return q
}

func (q _{{.Name}}Query) Run(dbtx gmq.DbTx) (sql.Result, error) {
	return q.Query.Exec(dbtx)
}

type {{.Name}}RowVisitor func(obj {{.Name}}) bool

func (q _{{.Name}}Query) Iterate(dbtx gmq.DbTx, functor {{.Name}}RowVisitor) error {
	return q.Query.SelectList(dbtx, func(columns []gmq.Column, rb []sql.RawBytes) bool {
		obj := {{.Name}}Objs.to{{.Name}}(columns, rb)
		return functor(obj)
	})
}

func (q _{{.Name}}Query) One(dbtx gmq.DbTx) ({{.Name}}, error) {
	var obj {{.Name}}
	err := q.Query.SelectOne(dbtx, func(columns []gmq.Column, rb []sql.RawBytes) bool {
		obj = {{.Name}}Objs.to{{.Name}}(columns, rb)
		return true
	})
	return obj, err
}

func (q _{{.Name}}Query) List(dbtx gmq.DbTx) ([]{{.Name}}, error) {
	result := make([]{{.Name}}, 0, 10)
	err := q.Query.SelectList(dbtx, func(columns []gmq.Column, rb []sql.RawBytes) bool {
		obj := {{.Name}}Objs.to{{.Name}}(columns, rb)
		result = append(result, obj)
		return true
	})
	return result, err
}
`

var managedApi string = `
// Start of the model facade Apis.

type _{{.Name}}Objs struct {
	fcMap map[string]string
}

func (o _{{.Name}}Objs) Names() (schema, tbl, alias string) { 
	return "{{.DbName}}", "{{.TableName}}", "{{.Name}}" 
}

func (o _{{.Name}}Objs) Select(fields ...string) _{{.Name}}Query {
	q := _{{.Name}}Query{}
	if len(fields) == 0 {
		fields = []string{ {{.AllFields}} }
	}
	q.Query = gmq.Select(o, o.columns(fields...))
	return q
}

func (o _{{.Name}}Objs) Insert(obj {{.Name}}) _{{.Name}}Query {
	q := _{{.Name}}Query{}
	q.Query = gmq.Insert(o, o.columnsWithData(obj, {{.InsertableFields}}))
	return q
}

func (o _{{.Name}}Objs) Update(obj {{.Name}}, fields ...string) _{{.Name}}Query {
	q := _{{.Name}}Query{}
	q.Query = gmq.Update(o, o.columnsWithData(obj, fields...))
	return q
}

func (o _{{.Name}}Objs) Delete() _{{.Name}}Query {
	q := _{{.Name}}Query{}
	q.Query = gmq.Delete(o)
	return q
}

{{$ModelName := .Name }}
///// Managed Objects Filters definition
{{range .Fields}}
func (o _{{$ModelName}}Objs) Filter{{.Name}}(op string, p {{.Type}}, ps ...{{.Type}}) gmq.Filter {
	params := make([]interface{}, 1+len(ps))
	params[0] = p
	for i := range ps {
		params[i+1] = ps[i]
	}
	return o.newFilter("{{.ColumnName}}", op, params...)
}

{{end}}

///// Managed Objects Columns definition
{{range .Fields}}
func (o _{{$ModelName}}Objs) Column{{.Name}}(p ...{{.Type}}) gmq.Column {
	var value interface{}
	if len(p) > 0 {
		value = p[0]
	}
	return gmq.Column{"{{.ColumnName}}", value}
}
{{end}}

////// Internal helper funcs

func (o _{{.Name}}Objs) newFilter(name, op string, params ...interface{}) gmq.Filter {
	if strings.ToUpper(op) == "IN" {
		return gmq.InFilter(name, params)
	}
	return gmq.UnitFilter(name, op, params[0])
}

func (o _{{.Name}}Objs) to{{.Name}}(columns []gmq.Column, rb []sql.RawBytes) {{.Name}} {
	obj := {{.Name}}{}
	if len(columns) == len(rb) {
		for i := range columns {
			switch columns[i].Name {
			{{range .Fields}}case "{{.ColumnName}}":
				obj.{{.Name}} = gmq.{{.ConverterFuncName}}(rb[i])
			{{end}} }
		}
	}
	return obj
}

func (o _{{.Name}}Objs) columns(fields ...string) []gmq.Column {
	data := make([]gmq.Column, 0, len(fields))
	for _, f := range fields {
		switch f {
		{{range .Fields}}case "{{.Name}}":
			data = append(data, o.Column{{.Name}}())
		{{end}} }
	}
	return data
}

func (o _{{.Name}}Objs) columnsWithData(obj {{.Name}}, fields ...string) []gmq.Column {
	data := make([]gmq.Column, 0, len(fields))
	for _, f := range fields {
		switch f {
		{{range .Fields}}case "{{.Name}}":
			data = append(data, o.Column{{.Name}}(obj.{{.Name}}))
		{{end}} }
	}
	return data
}

var {{.Name}}Objs _{{.Name}}Objs

func init() {
	{{.Name}}Objs.fcMap = map[string]string{
		{{range .Fields}}"{{.Name}}": "{{.ColumnName}}",
		{{end}} }
	gob.Register({{.Name}}{})
}
`

var (
	tmHeader        *template.Template
	tmStruct        *template.Template
	tmObjApi        *template.Template
	tmQueryApi      *template.Template
	tmManagedObjApi *template.Template
)

func init() {
	tmHeader = template.Must(template.New("header").Parse(header))
	tmStruct = template.Must(template.New("modelStruct").Parse(modelStruct))
	tmObjApi = template.Must(template.New("objApi").Parse(objApi))
	tmQueryApi = template.Must(template.New("queryApi").Parse(queryApi))
	tmManagedObjApi = template.Must(template.New("managedApi").Parse(managedApi))
}
