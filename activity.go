package pg_info

import (
	"context"
	"strings"

	"zgo.at/errors"
	"zgo.at/zdb"
)

// Activity lists the current query activity from pg_stat_activity.
type Activity []struct {
	PID      int64  `db:"pid"`
	Duration string `db:"duration"`
	Query    string `db:"query"`
}

func (Activity) Name() string { return "activity" }

func (a *Activity) Data(ctx context.Context) error {
	err := zdb.MustGet(ctx).SelectContext(ctx, a, `/* pg_info */
		select
			pid,
			now() - pg_stat_activity.query_start as duration,
			query
		from pg_stat_activity
		where state != 'idle' and query not like '%f/* pg_info */%';
	`)
	if err != nil {
		return errors.Errorf("pg_info.Activity.Data: %w", err)
	}

	aa := *a
	for i := range aa {
		aa[i].Query = normalizeQueryIndent(aa[i].Query)
	}

	*a = aa
	return nil
}

// Normalize the indent a bit, because there are often of extra tabs inside
// Go literal strings.
func normalizeQueryIndent(q string) string {
	lines := strings.Split(q, "\n")
	if len(lines) < 2 {
		return strings.TrimSpace(q)
	}

	var n int
	for _, l := range lines {
		if strings.TrimSpace(l) == "" {
			continue
		}
		if strings.HasPrefix(strings.TrimSpace(l), "/*") {
			continue
		}
		n = strings.Count(lines[1], "\t") - 1
		break
	}

	for j := range lines {
		lines[j] = strings.Replace(lines[j], "\t", "", n)
	}
	return strings.Join(lines, "\n")
}
