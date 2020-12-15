package pg_info

import (
	"context"
	"fmt"

	"zgo.at/errors"
	"zgo.at/zdb"
)

// Statement from pg_stat_statements.
type Statement struct {
	Total      float64 `db:"total"`
	MeanTime   float64 `db:"mean_time"`
	MinTime    float64 `db:"min_time"`
	MaxTime    float64 `db:"max_time"`
	StdDevTime float64 `db:"stddev_time"`
	Calls      int     `db:"calls"`
	HitPercent float64 `db:"hit_percent"`
	QueryID    int64   `db:"queryid"`
	Query      string  `db:"query"`
}

type Statements struct {
	Statements []Statement
	Filter     string
	Order      string
	Asc        bool
}

func (Statements) Name() string { return "statements" }

func (s *Statements) Data(ctx context.Context) error {
	if s.Order == "" {
		s.Order = "total"
	}
	dir := "desc"
	if s.Asc {
		dir = "asc"
	}

	var (
		args  []interface{}
		where string
	)
	if s.Filter != "" {
		args = append(args, "%"+s.Filter+"%")
		where = ` query like $1 `
	} else {
		where = ` calls >= 20 `
	}

	err := zdb.MustGet(ctx).SelectContext(ctx, &s.Statements, fmt.Sprintf(`/* pg_info */
		select
			(total_time / 1000 / 60) as total,
			mean_time,
			min_time,
			max_time,
			stddev_time,
			calls,
			coalesce(100.0 * shared_blks_hit / nullif(shared_blks_hit + shared_blks_read, 0), 0) as hit_percent,
			queryid,
			query
		from pg_stat_statements where
			userid = (select usesysid from pg_user where usename = CURRENT_USER) and
			query !~* '(^ *(copy|create|alter|explain) | (pg_stat_|pg_catalog)|^(COMMIT|BEGIN READ WRITE)$)' and
			%s
		order by %s %s, queryid asc
		limit 100
	`, where, s.Order, dir), args...)
	if err != nil {
		return errors.Errorf("pg_info.Statements.Data: %w", err)
	}

	ss := *s
	for i := range ss.Statements {
		ss.Statements[i].Query = normalizeQueryIndent(ss.Statements[i].Query)
	}
	*s = ss
	return nil
}
