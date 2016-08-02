// Copyright (c) 2016, German Neuroinformatics Node (G-Node)
//
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted under the terms of the BSD License. See
// LICENSE file in the root of the Project.

package web

import (
	"fmt"
	"net/http"

	"github.com/G-Node/gin-auth/conf"
	"github.com/G-Node/gin-auth/data"
	"github.com/G-Node/gin-auth/util"
)

type credentialData struct {
	Credential string
	ErrMessage string
}

// ResetInitPage provides an input form for resetting an account password
func ResetInitPage(w http.ResponseWriter, r *http.Request) {

	credData := &credentialData{}

	tmpl := conf.MakeTemplate("resetinit.html")
	w.Header().Add("Cache-Control", "no-store")
	w.Header().Add("Content-Type", "text/html")
	err := tmpl.ExecuteTemplate(w, "layout", credData)
	if err != nil {
		panic(err)
	}
}

// ResetInit checks whether a provided login or e-mail address
// belongs to a non-disabled account. If this is the case, the corresponding
// account is updated with a password reset code and an email containing
// the code is sent to the e-mail address of the account.
func ResetInit(w http.ResponseWriter, r *http.Request) {

	w.Header().Add("Cache-Control", "no-store")
	w.Header().Add("Content-Type", "text/html")

	credData := &credentialData{}

	err := util.ReadFormIntoStruct(r, credData, true)
	if err != nil {
		panic(err)
	}

	if credData.Credential == "" {
		credData.ErrMessage = "Please enter your login or e-mail address"
		tmpl := conf.MakeTemplate("resetinit.html")
		w.Header().Add("Warning", credData.ErrMessage)
		err = tmpl.ExecuteTemplate(w, "layout", credData)
		if err != nil {
			panic(err)
		}
		return
	}

	account, ok := data.SetPasswordReset(credData.Credential)
	if !ok {
		credData.ErrMessage = "Invalid login or e-mail address"
		tmpl := conf.MakeTemplate("resetinit.html")
		w.Header().Add("Warning", credData.ErrMessage)
		err = tmpl.ExecuteTemplate(w, "layout", credData)
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
	tmplFields.Subject = "Your GIN Account Password Reset Request"
	tmplFields.BaseUrl = conf.GetServerConfig().BaseURL
	tmplFields.Code = account.ResetPWCode.String

	content := util.MakeEmailTemplate("emailreset.txt", tmplFields)
	disp := util.NewEmailDispatcher()

	err = disp.Send([]string{account.Email}, content.Bytes())
	if err != nil {
		msg := "An error occurred trying to send password reset e-mail. Please try again later."
		PrintErrorHTML(w, r, msg, http.StatusInternalServerError)
		return
	}

	head := "Success!"
	message := "An e-mail with a password reset token has been sent to your e-mail address. "
	message += "Please follow the contained link to reset your password. "
	message += "Please note that your account will stay deactivated until your password reset has been completed."
	info := struct {
		Header  string
		Message string
	}{head, message}

	tmpl := conf.MakeTemplate("success.html")
	err = tmpl.ExecuteTemplate(w, "layout", info)
}

// ResetPage checks whether a password reset code submitted by request URI query exists and is still valid.
// Display enter password form if valid, an error message otherwise.
func ResetPage(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		PrintErrorHTML(w, r, "Request was malformed", http.StatusBadRequest)
		return
	}

	code := r.Form.Get("reset_code")
	if code == "" {
		PrintErrorHTML(w, r, "Request was malformed", http.StatusBadRequest)
		return
	}

	_, exists := data.GetAccountByResetPWCode(code)
	if !exists {
		PrintErrorHTML(w, r, "Your request is invalid or outdated. Please request a new reset code.",
			http.StatusNotFound)
		return
	}

	hidden := &struct {
		ResetCode string
		*util.ValidationError
	}{code, &util.ValidationError{}}

	tmpl := conf.MakeTemplate("reset.html")
	w.Header().Add("Cache-Control", "no-store")
	w.Header().Add("Content-Type", "text/html")
	err = tmpl.ExecuteTemplate(w, "layout", hidden)
}

// Reset checks whether a submitted password reset code exists and is still valid. It further checks,
// whether posted password and confirm password are identical and updates the account associated with
// the password reset code with the new password. This update further removes any existing
// password reset and account activation codes rendering the account active.
func Reset(w http.ResponseWriter, r *http.Request) {
	formData := &struct {
		ResetCode       string
		Password        string
		PasswordControl string
		*util.ValidationError
	}{}

	err := util.ReadFormIntoStruct(r, formData, true)
	if err != nil {
		panic(err)
	}

	account, exists := data.GetAccountByResetPWCode(formData.ResetCode)
	if !exists {
		PrintErrorHTML(w, r, "Your request is invalid or outdated. Please request a new reset code.",
			http.StatusNotFound)
		return
	}

	formData.ValidationError = &util.ValidationError{FieldErrors: make(map[string]string)}
	if formData.Password != formData.PasswordControl {
		formData.FieldErrors["password"] = "Provided password did not match password control"
		formData.Message = formData.FieldErrors["password"]
	}
	if formData.Password == "" || formData.PasswordControl == "" {
		formData.FieldErrors["password"] = "Please enter password and password control"
		formData.Message = formData.FieldErrors["password"]
	}
	if len(formData.Password) > 512 || len(formData.PasswordControl) > 512 {
		formData.FieldErrors["password"] =
			fmt.Sprintf("Entry too long, please shorten to %d characters", 512)
		formData.Message = formData.FieldErrors["password"]
	}

	if formData.FieldErrors["password"] != "" {
		formData.Password = ""
		formData.PasswordControl = ""
		tmpl := conf.MakeTemplate("reset.html")
		w.Header().Add("Cache-Control", "no-store")
		w.Header().Add("Content-Type", "text/html")
		w.Header().Add("Warning", formData.Message)
		err := tmpl.ExecuteTemplate(w, "layout", formData)
		if err != nil {
			panic(err)
		}
		return
	}

	account.SetPassword(formData.Password)
	account.ActivationCode.Valid = false
	account.ResetPWCode.Valid = false
	err = account.Update()
	if err != nil {
		panic(err)
	}

	head := "Your password has been reset!"
	message := "You can now login using your new password."
	info := struct {
		Header  string
		Message string
	}{head, message}

	tmpl := conf.MakeTemplate("success.html")
	w.Header().Add("Cache-Control", "no-store")
	w.Header().Add("Content-Type", "text/html")
	err = tmpl.ExecuteTemplate(w, "layout", info)
}
