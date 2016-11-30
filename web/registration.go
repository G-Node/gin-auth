// Copyright (c) 2016, German Neuroinformatics Node (G-Node)
//
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted under the terms of the BSD License. See
// LICENSE file in the root of the Project.

package web

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/G-Node/gin-auth/conf"
	"github.com/G-Node/gin-auth/data"
	"github.com/G-Node/gin-auth/util"
	"github.com/dchest/captcha"
)

type validateAccount struct {
	*data.Account
	*util.ValidationError
	RequestId      string
	CaptchaId      string
	CaptchaResolve string
}

// RegistrationInit creates a grant request for an account registration
// and redirects to the actual registration entry form.
func RegistrationInit(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	if query.Get("response_type") != "client" {
		PrintErrorHTML(w, r, "Invalid response type", http.StatusBadRequest)
		return
	}
	createGrantRequest(w, r, "/oauth/registration_page")
}

// RegistrationPage displays entry fields required for the creation of a new gin account
func RegistrationPage(w http.ResponseWriter, r *http.Request) {
	requestID := r.URL.Query().Get("request_id")
	grantRequest, ok := data.GetGrantRequest(requestID)
	if !ok {
		PrintErrorHTML(w, r, "Grant request does not exist", http.StatusBadRequest)
		return
	}
	if !grantRequest.ScopeRequested.Contains("account-create") {
		PrintErrorHTML(w, r, "Invalid grant request", http.StatusBadRequest)
		return
	}

	valAccount := &validateAccount{}
	valAccount.Account = &data.Account{}
	valAccount.ValidationError = &util.ValidationError{}
	valAccount.RequestId = requestID
	valAccount.CaptchaId = captcha.New()

	tmpl := conf.MakeTemplate("registration.html")
	w.Header().Add("Cache-Control", "no-store")
	w.Header().Add("Content-Type", "text/html")
	err := tmpl.ExecuteTemplate(w, "layout", valAccount)
	if err != nil {
		panic(err)
	}
}

type passwordData struct {
	Password        string
	PasswordControl string
}

type registration struct {
	verifyCaptcha func(string, string) bool
}

// RegistrationHandler provides an http handler for account registration.
func RegistrationHandler(f func(string, string) bool) http.Handler {
	rh := &registration{verifyCaptcha: f}
	return rh
}

// The http handler of the registration class parses user entries for a new account. It will redirect back to the
// entry form, if input is invalid. If the input is correct, it will create a new account,
// send an e-mail with an activation link and redirect to the the registered page.
func (rh *registration) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tmpl := conf.MakeTemplate("registration.html")
	w.Header().Add("Cache-Control", "no-store")
	w.Header().Add("Content-Type", "text/html")

	account := &data.Account{}
	pw := &passwordData{}

	err := util.ReadFormIntoStruct(r, account, true)
	if err != nil {
		PrintErrorHTML(w, r, err, http.StatusInternalServerError)
		return
	}

	err = util.ReadFormIntoStruct(r, pw, true)
	if err != nil {
		PrintErrorHTML(w, r, err, http.StatusInternalServerError)
		return
	}

	valAccount := &validateAccount{}
	valAccount.ValidationError = &util.ValidationError{}
	valAccount.Account = account

	if r.Form.Encode() == "" {
		valAccount.Message = "Please add all required fields (*)"
		err := tmpl.ExecuteTemplate(w, "layout", valAccount)
		if err != nil {
			panic(err)
		}
		return
	}

	valAccount.RequestId = r.Form.Get("request_id")
	_, ok := data.GetGrantRequest(valAccount.RequestId)
	if !ok {
		// TODO check if handling this fail is sufficient or if there should be
		// a redirect to registration_init and start the registration process again.
		PrintErrorHTML(w, r, "Grant request does not exist", http.StatusBadRequest)
		return
	}
	valAccount.ValidationError = valAccount.Account.Validate()

	if pw.Password != pw.PasswordControl {
		valAccount.FieldErrors["password"] = "Provided password did not match password control"
		if valAccount.Message == "" {
			valAccount.Message = valAccount.FieldErrors["password"]
		}
	}
	if pw.Password == "" || pw.PasswordControl == "" {
		valAccount.FieldErrors["password"] = "Please enter password and password control"
		if valAccount.Message == "" {
			valAccount.Message = valAccount.FieldErrors["password"]
		}
	}
	if len(pw.Password) > 512 || len(pw.PasswordControl) > 512 {
		valAccount.FieldErrors["password"] =
			fmt.Sprintf("Entry too long, please shorten to %d characters", 512)
		if valAccount.Message == "" {
			valAccount.Message = valAccount.FieldErrors["password"]
		}
	}

	captchaRes := r.PostForm.Get("captcha_resolve")
	captchaId := r.PostForm.Get("captcha_id")
	isCorrectCaptcha := rh.verifyCaptcha(captchaId, captchaRes)
	if !isCorrectCaptcha {
		valAccount.FieldErrors["password"] = "Please enter password and password control"
		valAccount.FieldErrors["captcha"] = "Please resolve verification"
		if valAccount.Message == "" {
			valAccount.Message = valAccount.FieldErrors["captcha"]
		}
	}

	if valAccount.Message != "" {
		valAccount.CaptchaId = captcha.New()
		err := tmpl.ExecuteTemplate(w, "layout", valAccount)
		if err != nil {
			panic(err)
		}
		return
	}

	valAccount.Account.SetPassword(pw.Password)
	valAccount.Account.ActivationCode = sql.NullString{String: util.RandomToken(), Valid: true}

	err = account.Create()
	if err != nil {
		valAccount.Message = "An error occurred during registration."
		err := tmpl.ExecuteTemplate(w, "layout", valAccount)
		if err != nil {
			panic(err)
		}
		return
	}

	tmplFields := &struct {
		From    string
		To      string
		Subject string
		BaseUrl string
		Code    string
	}{}
	tmplFields.From = conf.GetSmtpCredentials().From
	tmplFields.To = account.Email
	tmplFields.Subject = "GIN account activation"
	tmplFields.BaseUrl = conf.GetServerConfig().BaseURL
	tmplFields.Code = account.ActivationCode.String

	content := util.MakeEmailTemplate("emailactivate.txt", tmplFields)
	email := &data.Email{}
	err = email.Create(util.NewStringSet(account.Email), content.Bytes())
	if err != nil {
		msg := "An error occurred trying to send registration e-mail. Please contact an administrator."
		PrintErrorHTML(w, r, msg, http.StatusInternalServerError)
		return
	}

	w.Header().Add("Cache-Control", "no-store")
	http.Redirect(w, r, "/oauth/registered_page", http.StatusFound)
}

// RegisteredPage displays information about how a newly created gin account can be activated.
func RegisteredPage(w http.ResponseWriter, r *http.Request) {
	head := "Account registered"
	message := "Your account activation is pending. "
	message += "An e-mail with an activation code has been sent to your e-mail address."

	info := struct {
		Header  string
		Message string
	}{head, message}

	tmpl := conf.MakeTemplate("success.html")
	w.Header().Add("Cache-Control", "no-store")
	w.Header().Add("Content-Type", "text/html")
	err := tmpl.ExecuteTemplate(w, "layout", info)
	if err != nil {
		panic(err)
	}
}

// Activation removes an existing activation code from an account, thus rendering the account active.
func Activation(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		PrintErrorHTML(w, r, "Activation request was malformed", http.StatusBadRequest)
		return
	}

	getCode := r.Form.Get("activation_code")
	if getCode == "" {
		PrintErrorHTML(w, r, "Account activation code was absent", http.StatusBadRequest)
		return
	}

	account, exists := data.GetAccountByActivationCode(getCode)
	if !exists {
		PrintErrorHTML(w, r, "Requested account does not exist", http.StatusNotFound)
		return
	}

	err = account.RemoveActivationCode()
	if err != nil {
		panic(err)
	}

	head := "Account activation"
	message := fmt.Sprintf("Congratulation %s %s! The account for %s has been activated and can now be used.",
		account.FirstName, account.LastName, account.Login)
	info := struct {
		Header  string
		Message string
	}{head, message}

	tmpl := conf.MakeTemplate("success.html")
	w.Header().Add("Cache-Control", "no-store")
	w.Header().Add("Content-Type", "text/html")

	err = tmpl.ExecuteTemplate(w, "layout", info)
	if err != nil {
		panic(err)
	}
}
