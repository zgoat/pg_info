package pg_info

import (
	"context"
	"fmt"
	"strings"
	"text/tabwriter"

	"zgo.at/errors"
	"zgo.at/zdb"
)

// Column statistics.
type Column struct {
	AttName     string  `db:"attname"`
	NullFrac    float64 `db:"null_frac"`
	AvgWidth    int     `db:"avg_width"`
	NDistinct   float64 `db:"n_distinct"`
	Correlation float64 `db:"correlation"`
}

func (c Column) String() string {
	return fmt.Sprintf("%#v\n", c)
}

type Columns []Column

func (c Columns) String() string {
	var (
		b = new(strings.Builder)
		t = tabwriter.NewWriter(b, 8, 8, 2, ' ', 0)
	)
	fmt.Fprint(t, "Col\tNullFrac\tAvgWidth\tNDistinct\tCorrlation\n")
	for _, x := range c {
		fmt.Fprintf(t, "%s\t%f\t%d\t%f\t%f\n", x.AttName, x.NullFrac, x.AvgWidth, x.NDistinct, x.Correlation)
	}
	t.Flush()
	return b.String()
}

// List all column statistics for a table.
func (c *Columns) List(ctx context.Context, table string) error {
	// TODO: public
	err := zdb.MustGet(ctx).SelectContext(ctx, c, `/* pg_info */
		select
			attname, coalesce(null_frac, 0) as null_frac, avg_width, n_distinct, correlation
		from pg_stats
		where schemaname='public' and tablename=$1`,
		table)
	return errors.Wrap(err, "pg_info.Columns.List")
}
