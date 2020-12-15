package pg_info

import (
	"context"
	"time"

	"zgo.at/errors"
	"zgo.at/zdb"
)

// Tables overview from pg_stat_user_tables.
type Tables []struct {
	Table   string `db:"relname"`
	SeqScan int64  `db:"seq_scan"`
	IdxScan int64  `db:"idx_scan"`
	SeqRead int64  `db:"seq_tup_read"`
	IdxRead int64  `db:"idx_tup_fetch"`

	LastVacuum      time.Time `db:"last_vacuum"`
	LastAutoVacuum  time.Time `db:"last_autovacuum"`
	LastAnalyze     time.Time `db:"last_analyze"`
	LastAutoAnalyze time.Time `db:"last_autoanalyze"`

	VacuumCount  int `db:"vacuum_count"`
	AnalyzeCount int `db:"analyze_count"`

	LiveTup         int64 `db:"n_live_tup"`
	DeadTup         int64 `db:"n_dead_tup"`
	ModSinceAnalyze int64 `db:"n_mod_since_analyze"`

	TableSize   int `db:"table_size"`
	IndexesSize int `db:"indexes_size"`
}

func (Tables) Name() string { return "tables" }

func (t *Tables) Data(ctx context.Context) error {
	err := zdb.MustGet(ctx).SelectContext(ctx, t, `/* pg_info */
		select
			relname,

			coalesce(seq_scan, 0) as seq_scan,
			coalesce(seq_tup_read, 0) as seq_tup_read,
			coalesce(idx_scan, 0) as idx_scan,
			coalesce(idx_tup_fetch, 0) as idx_tup_fetch,

			date(coalesce(last_vacuum,      now() - interval '50 year')) as last_vacuum,
			date(coalesce(last_autovacuum,  now() - interval '50 year')) as last_autovacuum,
			date(coalesce(last_analyze,     now() - interval '50 year')) as last_analyze,
			date(coalesce(last_autoanalyze, now() - interval '50 year')) as last_autoanalyze,

			vacuum_count  + autovacuum_count  as vacuum_count,
			analyze_count + autoanalyze_count as analyze_count,

			n_live_tup,
			n_dead_tup,
			n_mod_since_analyze,

			pg_table_size(  '"' || schemaname || '"."' || relname || '"'  ) / 1024/1024 as table_size,
			pg_indexes_size('"' || schemaname || '"."'  || relname || '"') / 1024/1024 as indexes_size

		from pg_stat_user_tables
		order by table_size desc
		-- order by n_dead_tup
		-- 	/(n_live_tup
		-- 	* current_setting('autovacuum_vacuum_scale_factor')::float8
		-- 	+ current_setting('autovacuum_vacuum_threshold')::float8)
		-- 	desc
	`)
	return errors.Wrap(err, "pg_info.Tables.Data")
}
