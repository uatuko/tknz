package auth

import "html/template"

const (
	errorHtmlFile = ".static/tmpl/auth/errors/server.html"

	signUpVerifyEmailHtmlFile = ".static/tmpl/auth/sign-up/verify-email.html"
)

var (
	errorTmpl *template.Template

	signUpVerifyEmailTmpl *template.Template
)

func errorTemplate() *template.Template {
	if errorTmpl == nil {
		errorTmpl = template.Must(template.ParseFiles(errorHtmlFile))
	}

	return errorTmpl
}

func signUpVerifyEmailTemplate() *template.Template {
	if signUpVerifyEmailTmpl == nil {
		signUpVerifyEmailTmpl = template.Must(template.ParseFiles(signUpVerifyEmailHtmlFile))
	}

	return signUpVerifyEmailTmpl
}
