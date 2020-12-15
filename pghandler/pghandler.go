//go:generate go run gen.go

package pghandler

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/go-chi/chi"
	"github.com/jmoiron/sqlx"
	"zgo.at/errors"
	"zgo.at/pg_info"
	"zgo.at/zdb"
	"zgo.at/zhttp"
	"zgo.at/zhttp/ztpl/tplfunc"
)

var prefix string

func New(p string, db *sql.DB) *chi.Mux {
	// TODO: because I don't want to refactor right now and just want to get it
	// working.
	prefix = p

	// TODO: don't really need chi here.
	r := chi.NewRouter()

	dbx := sqlx.NewDb(db, "postgres")
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			*r = *r.WithContext(zdb.With(ctx, dbx))
			next.ServeHTTP(w, r)
		})
	})

	r.Get("/", zhttp.Wrap(dashboard))
	r.Get("/table/{table}", zhttp.Wrap(table))
	r.Post("/explain", zhttp.Wrap(explain))

	r.Get("/widget/{name}", zhttp.Wrap(widget))

	if prefix != "" {
		//"public/all.css": []byte(`/* FILE: ./aside.css */

		var mod = make(map[string][]byte)
		for k, v := range public {
			mod["public"+prefix+k[6:]] = v
		}
		for k, v := range mod {
			public[k] = v
		}

		for k := range public {
			fmt.Println("  =>", k)
		}
	}

	st := zhttp.NewStatic("./public", "", nil, public)
	r.Get("/{file}.js", st.ServeHTTP)
	r.Get("/{file}.css", st.ServeHTTP)

	return r
}

type Widget interface {
	Data(context.Context) error
	Name() string
}

var funcMap = func() template.FuncMap {
	f := make(template.FuncMap)
	for k, v := range tplfunc.FuncMap {
		f[k] = v
	}

	f["nformat"] = tplfunc.Number

	f["nformat64"] = func(n int64) string {
		s := strconv.FormatInt(n, 10)
		if len(s) < 4 {
			return s
		}

		b := []byte(s)
		for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
			b[i], b[j] = b[j], b[i]
		}

		var out []rune
		for i := range b {
			if i > 0 && i%3 == 0 && ',' > 1 {
				out = append(out, ',')
			}
			out = append(out, rune(b[i]))
		}

		for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
			out[i], out[j] = out[j], out[i]
		}
		return string(out)
	}
	return f
}()

var templates = func() *template.Template {
	tpls := template.New("").Funcs(funcMap)
	for path, tpl := range packed {
		tpls = template.Must(tpls.New(path[4:]).Parse(string(tpl)))
	}
	return tpls

	//return template.Must(template.New("").Funcs(funcMap).ParseGlob("./tpl/*.gohtml"))
}()

func dashboard(w http.ResponseWriter, r *http.Request) error {
	filter := r.URL.Query().Get("filter")
	order := r.URL.Query().Get("order")
	asc := r.URL.Query().Get("asc") != ""

	wantWidgets := []string{"explain", "system", "activity", "progress", "tables", "indexes", "statements"}

	var widgets []Widget
	for _, w := range wantWidgets {
		switch w {
		default:
			return errors.Errorf("unknown widget: %q", w)
		case "explain":
			widgets = append(widgets, &pg_info.Explain{Prefix: prefix})
		case "system":
			widgets = append(widgets, &pg_info.System{})
		case "activity":
			widgets = append(widgets, &pg_info.Activity{})
		case "progress":
			widgets = append(widgets, &pg_info.Progress{})
		case "tables":
			widgets = append(widgets, &pg_info.Tables{})
		case "indexes":
			widgets = append(widgets, &pg_info.Indexes{})
		case "statements":
			widgets = append(widgets, &pg_info.Statements{Filter: filter, Order: order, Asc: asc})
		}
	}

	var (
		widgetsHTML = make([]template.HTML, len(widgets))
		wg          sync.WaitGroup
	)
	wg.Add(len(widgets))
	for i, w := range widgets {
		go func(i int, w Widget) {
			defer wg.Done()

			err := w.Data(r.Context())
			if err != nil {
				panic(err)
			}

			// tpl, err := template.New("").Funcs(funcMap).Parse(w.Template())
			// if err != nil {
			// 	panic(err)
			// }

			buf := new(bytes.Buffer)
			err = templates.ExecuteTemplate(buf, w.Name()+".gohtml", w)
			if err != nil {
				panic(err)
			}

			widgetsHTML[i] = template.HTML(buf.String())
		}(i, w)
	}
	wg.Wait()

	//return zhttp.Template(w, "index.gohtml", struct {
	return templates.ExecuteTemplate(w, "dashboard.gohtml", struct {
		Widgets     []Widget
		WidgetsHTML []template.HTML
		Prefix      string
	}{widgets, widgetsHTML, prefix})
}

func widget(w http.ResponseWriter, r *http.Request) error {
	var widget Widget
	switch chi.URLParam(r, "name") {
	default:
		return errors.Errorf("unknown widget: %q", chi.URLParam(r, "name"))
	case "explain":
		widget = &pg_info.Explain{Prefix: prefix}
	case "system":
		widget = &pg_info.System{}
	case "activity":
		widget = &pg_info.Activity{}
	case "progress":
		widget = &pg_info.Progress{}
	case "tables":
		widget = &pg_info.Tables{}
	case "indexes":
		widget = &pg_info.Indexes{}
	case "statements":
		widget = &pg_info.Statements{}
	}

	err := widget.Data(r.Context())
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	err = templates.ExecuteTemplate(buf, widget.Name()+".gohtml", widget)
	if err != nil {
		panic(err)
	}
	widgetHTML := template.HTML(buf.String())

	wantWidgets := []string{"explain", "system", "activity", "progress", "tables", "indexes", "statements"}

	return templates.ExecuteTemplate(w, "widget.gohtml", struct {
		Widget     Widget
		WidgetHTML template.HTML
		Widgets    []string
		Prefix     string
	}{widget, widgetHTML, wantWidgets, prefix})
}

func explain(w http.ResponseWriter, r *http.Request) error {
	var args struct {
		Query string `json:"query"`
	}
	_, err := zhttp.Decode(r, &args)
	if err != nil {
		return err
	}

	var e []string
	err = zdb.MustGet(r.Context()).SelectContext(r.Context(), &e,
		`explain (analyze, costs, verbose, buffers) `+args.Query)
	if err != nil {
		return err
	}
	return zhttp.String(w, strings.Join(e, "\n"))
}

func table(w http.ResponseWriter, r *http.Request) error {
	var stats pg_info.Columns
	err := stats.List(r.Context(), chi.URLParam(r, "table"))
	if err != nil {
		return err
	}

	var b strings.Builder
	b.WriteString(`<table class="sort"><thead><tr>
		<th>Column</th>
		<th class="n">NullFrac</th>
		<th class="n">AvgWidth</th>
		<th class="n">NDistinct</th>
		<th class="n">Correlation</th>
	</tr></thead><tobdy>`)
	for _, s := range stats {
		b.WriteString(fmt.Sprintf(`<tr>
			<td>%s</td>
			<td class="n">%.3f</td>
			<td class="n">%d</td>
			<td class="n">%f</td>
			<td class="n">%f</td>
		</tr>`,
			template.HTMLEscapeString(s.AttName),
			s.NullFrac, s.AvgWidth, s.NDistinct, s.Correlation))
	}
	b.WriteString(`</tbody></table>`)

	return zhttp.String(w, b.String())
}
