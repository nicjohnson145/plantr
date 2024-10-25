package agent

import (
	"database/sql"

	"github.com/nicjohnson145/hlp"
)

type InventoryRow struct {
	Hash    string
	Path    *string
	Package *string
}

func (i *InventoryRow) ToDBRow() DBInventoryRow {
	row := DBInventoryRow{
		Hash: i.Hash,
	}

	if i.Path != nil {
		row.Path = sql.Null[string]{V: *i.Path, Valid: true}
	}

	if i.Package != nil {
		row.Package = sql.Null[string]{V: *i.Package, Valid: true}
	}

	return row
}

type DBInventoryRow struct {
	Hash    string           `db:"hash"`
	Path    sql.Null[string] `db:"path"`
	Package sql.Null[string] `db:"package"`
}

func (d *DBInventoryRow) ToInventoryRow() InventoryRow {
	row := InventoryRow{
		Hash: d.Hash,
	}

	if d.Path.Valid {
		row.Path = hlp.Ptr(d.Path.V)
	}
	if d.Package.Valid {
		row.Package = hlp.Ptr(d.Package.V)
	}

	return row
}
