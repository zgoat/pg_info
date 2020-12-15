package pg_info

import "context"

type Explain struct{ Prefix string }

func (Explain) Name() string               { return "explain" }
func (Explain) Data(context.Context) error { return nil }
