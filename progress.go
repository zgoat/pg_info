package pg_info

import (
	"context"

	"zgo.at/errors"
	"zgo.at/zdb"
)

// Progress lists the progress of VACUUM, CLUSTER, and CREATE INDEX commands.
type Progress []struct {
	Table   string `db:"relname"`
	Command string `db:"command"`
	Phase   string `db:"phase"`
	Status  string `db:"status"`
}

func (Progress) Name() string { return "progress" }

func (p *Progress) Data(ctx context.Context) error {
	// https://www.postgresql.org/docs/current/progress-reporting.html
	err := zdb.MustGet(ctx).SelectContext(ctx, p, `/* pg_info */
		select
			relname,
			phase,
			command,
				'lockers: '    || lockers_done    || '/' || lockers_total    || '; ' ||
				'blocks: '     || blocks_done     || '/' || blocks_total     || '; ' ||
				'tuples: '     || tuples_done     || '/' || tuples_total     || '; ' ||
				'partitions: ' || partitions_done || '/' || partitions_total
			as status
		from pg_stat_progress_create_index
		join pg_stat_all_tables using(relid)

		union select
			relname,
			phase,
			'VACUUM' as command,
				'heap_blks: '          || heap_blks_total    || ', ' || heap_blks_scanned || ', ' || heap_blks_vacuumed || '; ' ||
				'index_vacuum_count: ' || index_vacuum_count || '; ' ||
				'max_dead_tuples: '    || max_dead_tuples    || '; ' ||
				'num_dead_tuples: '    || num_dead_tuples
			as status
		from pg_stat_progress_vacuum
		join pg_stat_all_tables using(relid)

		union select
			relname,
			phase,
			'VACUUM FULL' as command,
				'cluster_index_relid: ' || cluster_index_relid || '; ' ||
				'heap_tuples: '         || heap_tuples_scanned || '/'  || heap_tuples_written || '; ' ||
				'heap_blks: '           || heap_blks_scanned   || '/'  || heap_blks_total     || '; ' ||
				'index_rebuild_count: ' || index_rebuild_count
			as status
		from pg_stat_progress_cluster
		join pg_stat_all_tables using(relid)

	`)

	return errors.Wrap(err, "pg_info.Progress.Data")
}
