package main

import (
	"encoding/json"
	"fmt"

	"github.com/andreyvit/edb"
	"github.com/andreyvit/mvp"
	"github.com/andreyvit/mvp/httperrors"
)

type ObjectMeta1 struct {
	ModCount uint64 `json:"_mod"`
}

type Object1 struct {
	Object any
	Meta   ObjectMeta1
}

type ListResult1 struct {
	Objects []Object1 `json:"objects"`
}

func (app *App) listSuperadminTables(rc *mvp.RC, in *struct {
}) (*mvp.ViewData, error) {
	return &mvp.ViewData{
		View:         "db/tables",
		Title:        "DB Tables",
		SemanticPath: "superadmin/db",
		Data: &struct {
			Tables []*edb.Table
		}{
			Tables: app.DBSchema.Tables(),
		},
		Layout: "default",
	}, nil
}

func (app *App) listSuperadminTableRows(rc *mvp.RC, in *struct {
	Table string `json:"-" form:"table,path"`
	Index string `json:"index"`
}) (*mvp.ViewData, error) {
	tbl := app.DBSchema.TableNamed(in.Table)
	if tbl == nil {
		return nil, httperrors.BadRequest.Msg(fmt.Sprintf("unknown table %q", in.Table))
	}
	if in.Index != "" {
		return nil, httperrors.BadRequest.Msg("index scans not implemented yet")
	}

	var rows []*adminDBTableRow
	res := new(ListResult1)
	var n int
	for c := rc.DBTx().TableScan(tbl, edb.FullScan()); c.Next(); {
		row, rowMeta := c.Row()
		keyStr := tbl.RawKeyString(c.RawKey())
		n++
		rows = append(rows, &adminDBTableRow{
			Index:     n,
			Key:       keyStr,
			SchemaVer: rowMeta.SchemaVer,
			ModCount:  rowMeta.ModCount,
			Data:      string(must(json.MarshalIndent(row, "", "  "))),
		})
		res.Objects = append(res.Objects, buildObject1(row, rowMeta))
	}
	return &mvp.ViewData{
		View:         "db/rows",
		Title:        tbl.Name(),
		SemanticPath: "superadmin/db/table",
		Data: &struct {
			Table *edb.Table
			Rows  []*adminDBTableRow
		}{
			Table: tbl,
			Rows:  rows,
		},
		Layout: "default",
	}, nil
}

type adminDBTableRow struct {
	Index     int
	Key       string
	ModCount  uint64
	SchemaVer uint64
	Data      string
}

const (
	NewRowKeyString = "--new"
)

func (app *App) handleSuperadminTableRowForm(rc *mvp.RC, in *struct {
	IsSaving bool   `json:"-" form:",issave"`
	Table    string `json:"-" form:"table,path"`
	Key      string `json:"-" form:"key,path"`
	Index    string `json:"index"`
	Data     string `json:"data"`
	ModCount int    `json:"modcount"`
	Delete   bool   `json:"delete"`
}) (any, error) {
	tbl := app.DBSchema.TableNamed(in.Table)
	if tbl == nil {
		return nil, httperrors.BadRequest.Msg(fmt.Sprintf("unknown table %q", in.Table))
	}
	if in.Index != "" {
		return nil, httperrors.BadRequest.Msg("index scans not implemented yet")
	}

	var row any
	var rowMeta edb.ValueMeta
	var isNew bool
	if in.Key == NewRowKeyString {
		isNew = true
		row, rowMeta = tbl.NewRow(), edb.ValueMeta{}
		app.SetNewKeyOnRow(row)
	} else {
		key, err := tbl.ParseKey(in.Key)
		if err != nil {
			return nil, httperrors.BadRequest.Msg("invalid key: " + err.Error())
		}

		row, rowMeta = rc.DBTx().Get(tbl, key)
		if row == nil {
			return nil, httperrors.NotFound.Msg(fmt.Sprintf("%s: key not found: %q", tbl.Name(), in.Key))
		}
	}

	var errors []string
	if in.IsSaving {
		if in.Delete {
			if isNew {
			} else {
				rc.DBTx().DeleteByKey(tbl, tbl.RowKey(row))
			}
			return app.Redirect("db.table.list", "table", in.Table), nil
		} else {
			row = tbl.NewRow()
			err := json.Unmarshal([]byte(in.Data), &row)
			if err != nil {
				errors = append(errors, fmt.Sprintf("invalid JSON: %v", err))
				goto render
			}

			if rowMeta.ModCount != uint64(in.ModCount) {
				errors = append(errors, fmt.Sprintf("conflict: edited mod %d != correct mod %d", uint64(in.ModCount), rowMeta.ModCount))
				goto render
			}

			if isNew && tbl.RowHasZeroKey(row) {
				app.SetNewKeyOnRow(row)
			}
			rc.DBTx().Put(tbl, row)
			return app.Redirect("db.table.show", "table", in.Table, "key", tbl.KeyString(tbl.RowKey(row))), nil
		}
	}

render:
	var keyStr string
	if isNew {
		keyStr = "New"
	} else {
		keyStr = tbl.KeyString(tbl.RowKey(row))
	}
	data := string(must(json.MarshalIndent(row, "", "  ")))

	return &mvp.ViewData{
		View:         "db/row",
		Title:        fmt.Sprintf("%s / %s", tbl.Name(), keyStr),
		SemanticPath: "superadmin/db/table/row",
		Data: &struct {
			Errors   []string
			Table    *edb.Table
			Key      string
			ModCount uint64
			Data     string
		}{
			Errors:   errors,
			Table:    tbl,
			Key:      keyStr,
			ModCount: rowMeta.ModCount,
			Data:     data,
		},
		Layout: "default",
	}, nil
}

func (app *App) showDBDump(rc *RC, in *struct{}) (mvp.DebugOutput, error) {
	var s string
	app.DB().Read(func(tx *edb.Tx) {
		s = tx.Dump(edb.DumpAll)
	})
	return mvp.DebugOutput(s), nil
}

func buildObject1(row any, meta edb.ValueMeta) Object1 {
	return Object1{
		Object: row,
		Meta:   buildObjectMeta1(meta),
	}
}

func buildObjectMeta1(meta edb.ValueMeta) ObjectMeta1 {
	return ObjectMeta1{ModCount: meta.ModCount}
}
