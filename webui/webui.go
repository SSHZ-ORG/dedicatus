package webui

import (
	"html/template"
	"net/http"

	"github.com/SSHZ-ORG/dedicatus/config"
	"github.com/SSHZ-ORG/dedicatus/dctx"
	"github.com/SSHZ-ORG/dedicatus/models"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/xid"
	"google.golang.org/appengine/log"
)

type page struct {
	Query          string
	NextPageCursor string
	Inventories    []inventory
}

type inventory struct {
	VideoURL    string
	Description string
}

var tmpl = template.Must(template.New("template").Parse(`<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8" />
	<meta name="viewport" content="width=device-width, initial-scale=1.0" />
	<title>Dedicatus</title>
	<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/water.css@2/out/water.min.css">
</head>
<body>
	<h1>Dedicatus</h1>
	<form method="get">
		<label for="query">Query</label>
		<input name="query" id="query" value="{{.Query}}"/>
		<input type="submit" value="Submit" />
	</form>

	<table>
		<thead>
			<tr>
				<th>Video</th>
				<th>Description</th>
			</tr>
		</thead>
		<tbody>
			{{range $i := .Inventories}}
			<tr>
				<td><video controls><source src="{{$i.VideoURL}}" type="video/mp4"></video></td>
				<td><pre>{{$i.Description}}</pre></td>
			</tr>
			{{end}}
		</tbody>
	</table>

	{{if .NextPageCursor}}
		<form method="get">
			<input name="query" value="{{.Query}}" type="hidden"/>
			<input name="cursor" value="{{.NextPageCursor}}" type="hidden"/>
			<input type="submit" value="Next Page" />
		</form>
	{{end}}
</body>
</html>
`))

func Handler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx := dctx.NewContext(r)

	query := r.FormValue("query")
	cursor := r.FormValue("cursor")

	data := &page{Query: query}

	if query != "" {
		is, nextCursor, err := models.QueryInventories(ctx, query, cursor, xid.New().String(), 20)
		if err != nil {
			log.Errorf(ctx, "models.QueryInventories: %+v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		for _, i := range is {
			s, err := i.ToString(ctx)
			if err != nil {
				log.Errorf(ctx, "Inventory.ToString: %+v", err)
				s = "(error rendering text)"
			}
			data.Inventories = append(data.Inventories, inventory{
				VideoURL:    config.WebUIVideoURLPrefix + i.FileID + ".mp4",
				Description: s,
			})
		}
		data.NextPageCursor = nextCursor
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Errorf(ctx, "tmpl.Execute: %+v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
