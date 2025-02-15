package routes

import (
	"ConcertGetApp/dataTypes"
	"ConcertGetApp/dbInterface"
	"ConcertGetApp/helpers"
	"ConcertGetApp/helpers/cookieHelpers"
	"ConcertGetApp/htmlTemplate"
	"fmt"
	"github.com/cockroachdb/errors"
	"net/http"
)

func NewUser(w http.ResponseWriter, r *http.Request) {
	username, _, _, readErr := cookieHelpers.ReadCookies(r)
	if readErr != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	dbI, newDbIErr := dbInterface.New()
	if newDbIErr != nil {
		http.Error(w, "Could not access the logged in user.", http.StatusInternalServerError)
		return
	}
	defer dbI.Close()

	user, getUserErr := dbI.GetUserByUsername(username)
	if getUserErr != nil {
		fmt.Printf("Error: %+v\n", getUserErr)
		http.Error(w, getUserErr.Error(), http.StatusInternalServerError)
		return
	}

	if user == nil {
		fmtErr := errors.New("user is nil")
		fmt.Println(fmtErr)
		http.Error(w, fmtErr.Error(), http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodPost {
		parseErr := r.ParseForm()
		if parseErr != nil {
			fmt.Printf("Error: %s", parseErr)
			http.Error(w, errors.Wrap(parseErr, "error parsing form").Error(), http.StatusInternalServerError)
		}

		newUsername := r.FormValue("username")
		newPassword := r.FormValue("password")

		hashed, hashErr := helpers.HashPassword(newPassword, newUsername)
		if hashErr != nil {
			fmtErr := errors.Wrap(hashErr, "error hashing password")
			fmt.Println(fmtErr)
			http.Error(w, fmtErr.Error(), http.StatusInternalServerError)
		}

		newUserErr := dbI.NewUser(newUsername, hashed)
		if newUserErr != nil {
			fmtErr := errors.Wrap(newUserErr, "error creating new user")
			fmt.Println(fmtErr)
			http.Error(w, fmtErr.Error(), http.StatusInternalServerError)
		}
	}

	user.PasswordHash = []byte("redacted")
	user.Id = nil

	formPackage := struct {
		Artists  []dataTypes.Artist
		Venues   []dataTypes.Venue
		Concerts []dataTypes.Concert
		User     dataTypes.User
	}{
		User: *user,
	}

	tmpl, err := htmlTemplate.GetTemplate("new-user.gohtml")
	if err != nil {
		fmt.Printf("Error: %+v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, formPackage)
	if err != nil {
		fmt.Printf("Error: %+v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}