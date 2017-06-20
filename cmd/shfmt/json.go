// Copyright (c) 2017, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package main

import (
	"encoding/json"
	"go/ast"
	"io"
	"reflect"

	"github.com/mvdan/sh/syntax"
)

func writeJSON(w io.Writer, f *syntax.File, pretty bool) error {
	v, _ := recurse(reflect.ValueOf(f))
	enc := json.NewEncoder(w)
	if pretty {
		enc.SetIndent("", "\t")
	}
	return enc.Encode(v)
}

func recurse(val reflect.Value) (interface{}, string) {
	switch val.Kind() {
	case reflect.Ptr:
		elem := val.Elem()
		if !elem.IsValid() {
			return nil, ""
		}
		return recurse(elem)
	case reflect.Interface:
		v, tname := recurse(val.Elem())
		m := v.(map[string]interface{})
		m["Type"] = tname
		return m, ""
	case reflect.Struct:
		m := make(map[string]interface{}, val.NumField()+1)
		typ := val.Type()
		for i := 0; i < val.NumField(); i += 1 {
			tfield := typ.Field(i)
			if tfield.Type.Name() == "Pos" {
				continue
			}
			if !ast.IsExported(tfield.Name) {
				continue
			}
			m[tfield.Name], _ = recurse(val.Field(i))
		}
		return m, typ.Name()
	case reflect.Slice:
		l := make([]interface{}, val.Len())
		for i := 0; i < val.Len(); i += 1 {
			l[i], _ = recurse(val.Index(i))
		}
		return l, ""
	default:
		return val.Interface(), ""
	}
}
