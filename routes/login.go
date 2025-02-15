package routes

import (
	"ConcertGetApp/dbInterface"
	"ConcertGetApp/helpers"
	"ConcertGetApp/helpers/cookieHelpers"
	"ConcertGetApp/htmlTemplate"
	"fmt"
	"github.com/cockroachdb/errors"
	"net/http"
	"slices"
)

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl, err := htmlTemplate.GetTemplate("login.gohtml")
		if err != nil {
			fmt.Printf("Error: %+v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = tmpl.Execute(w, nil)
		if err != nil {
			fmt.Printf("Error: %+v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			fmt.Printf("Error: %+v\n", err)
			http.Error(w, errors.Wrap(err, "error parsing form").Error(), http.StatusInternalServerError)
			return
		}

		username := r.FormValue("username")
		password := r.FormValue("password")

		dbI, newDbIErr := dbInterface.New()
		if newDbIErr != nil {
			http.Error(w, "Could not access the logged in user.", http.StatusInternalServerError)
			return
		}
		defer dbI.Close()

		user, getErr := dbI.GetUserByUsername(username)
		if getErr != nil {
			fmt.Printf("Error: %+v\n", getErr)
			http.Error(w, getErr.Error(), http.StatusInternalServerError)
			return
		}

		if user == nil {
			fmt.Printf("no user found with username: %s\n", username)
			http.Error(w, fmt.Sprintf("no user found with username: %s", username), http.StatusNotFound)
			return
		}

		hash, hashErr := helpers.HashPassword(password, username)
		if hashErr != nil {
			fmt.Printf("Error: %+v\n", hashErr)
			http.Error(w, hashErr.Error(), http.StatusInternalServerError)
			return
		}

		if !slices.Equal(user.PasswordHash, hash) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		check, checkErr := cookieHelpers.CookieCheck(username, hash)
		if checkErr != nil {
			fmt.Printf("Error: %+v\n", checkErr)
			http.Error(w, checkErr.Error(), http.StatusInternalServerError)
			return
		}

		writeErr := cookieHelpers.WriteCookies(w, username, hash, check)
		if writeErr != nil {
			fmt.Printf("Error: %+v\n", writeErr)
			http.Error(w, writeErr.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}