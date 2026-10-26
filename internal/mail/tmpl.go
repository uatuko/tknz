package mail

import "html/template"

const (
	signUpVerityHtmlFile = ".static/tmpl/email/sign-up/verify.html"
)

var (
	signUpVerityTmpl *template.Template
)

func signUpVerityTemplate() *template.Template {
	if signUpVerityTmpl == nil {
		signUpVerityTmpl = template.Must(template.ParseFiles(signUpVerityHtmlFile))
	}

	return signUpVerityTmpl
}
