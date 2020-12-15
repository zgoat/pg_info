package pg_info

import (
	"context"

	"zgo.at/errors"
	"zgo.at/zdb"
)

// Indexes lists information about indexes from pg_stat_user_indexes.
type Indexes []struct {
	Table    string `db:"relname"`
	Size     int    `db:"size"`
	Index    string `db:"indexrelname"`
	Scan     int64  `db:"idx_scan"`
	TupRead  int64  `db:"idx_tup_read"`
	TupFetch int64  `db:"idx_tup_fetch"`
}

func (Indexes) Name() string { return "indexes" }

func (i *Indexes) Data(ctx context.Context) error {
	err := zdb.MustGet(ctx).SelectContext(ctx, i, `/* pg_info */
		select
			relname,
			pg_relation_size('"' || schemaname || '"."' || indexrelname || '"') / 1024/1024 as size,
			indexrelname,
			idx_scan,
			idx_tup_read,
			idx_tup_fetch
		from pg_stat_user_indexes
		order by idx_scan desc
	`)
	return errors.Wrap(err, "pg_info.Indexes.Data")
}
