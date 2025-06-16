package handlers

import (
	"fmt"
	"html/template"
	"log"
	textplate "text/template"
)


var (
	LoginTmpl *template.Template
	HomeTmpl *template.Template
	ErrorTmpl *template.Template
	//UploadTmpl *template.Template
	//ResultTmpl *template.Template

	HomeTmplTerm *textplate.Template
	ResultTmplTerm *textplate.Template
)

func LoadTemplates() {
	var err error

	LoginTmpl, err = template.ParseFiles("templates/login.tmpl")
	if err != nil {
		log.SetPrefix(fmt.Sprintf("[\033[31mTEMPLATE ERR\033[0m] "))
		log.Fatalf("Failed to parse template: %v", err)
	}
	HomeTmpl, err = template.ParseFiles("templates/home.tmpl")
	if err != nil {
		log.SetPrefix(fmt.Sprintf("[\033[31mTEMPLATE ERR\033[0m] "))
		log.Fatalf("Failed to parse template: %v", err)
	}
	ErrorTmpl, err = template.ParseFiles("templates/error.tmpl")
	if err != nil {
		log.SetPrefix(fmt.Sprintf("[\033[31mTEMPLATE ERR\033[0m] "))
		log.Fatalf("Failed to parse template: %v", err)
	}
	//UploadTmpl, err = template.ParseFiles("templates/upload.tmpl")
	//if err != nil {
	//	log.Fatalf("Failed to parse template: %v", err)
	//}
	//ResultTmpl, err = template.ParseFiles("templates/result.tmpl")
	//if err != nil {
	//	log.Fatalf("Failed to parse template: %v", err)
	//}
	//Terminal Response Templates
	HomeTmplTerm, err = textplate.ParseFiles("templates/home_term.tmpl")
	if err != nil {
		log.SetPrefix(fmt.Sprintf("[\033[31mTEMPLATE ERR\033[0m] "))
		log.Fatalf("Failed to parse template: %v", err)
	}
	ResultTmplTerm, err = textplate.ParseFiles("templates/result_term.tmpl")
	if err != nil {
		log.SetPrefix(fmt.Sprintf("[\033[31mTEMPLATE ERR\033[0m] "))
		log.Fatalf("Failed to parse template: %v", err)
	}
}
