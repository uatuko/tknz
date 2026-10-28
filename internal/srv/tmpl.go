package srv

import "html/template"

const (
	errorHtmlFile = ".static/tmpl/errors/server.html"
)

var (
	errorTmpl *template.Template
)

func errorTemplate() *template.Template {
	if errorTmpl == nil {
		errorTmpl = template.Must(template.ParseFiles(errorHtmlFile))
	}

	return errorTmpl
}
