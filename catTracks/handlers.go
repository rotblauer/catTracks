package catTracks

import (
	"html/template"
	"net/http"


)

var funcMap = template.FuncMap{
	"eq": func(a, b interface{}) bool {
		return a == b
	},
}

var templates = template.Must(template.ParseGlob("templates/*.html"))



//Welcome
func indexHandler(w http.ResponseWriter, r *http.Request) {
	templates.Funcs(funcMap)
	templates.ExecuteTemplate(w, "base", nil)
}

