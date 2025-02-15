package middleware

import (
	"ConcertGetApp/dbInterface"
	"ConcertGetApp/helpers/cookieHelpers"
	"net/http"
	"slices"
)

func LoginCheck(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		username, passwordHash, check, readErr := cookieHelpers.ReadCookies(r)
		if readErr != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		computedCheck, checkErr := cookieHelpers.CookieCheck(username, passwordHash)
		if checkErr != nil {
			http.Error(w, checkErr.Error(), http.StatusInternalServerError)
			return
		}

		if !slices.Equal(check, computedCheck) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		dbI, newErr := dbInterface.New()
		if newErr != nil {
			http.Error(w, "Could not validate user", http.StatusInternalServerError)
			return
		}
		defer dbI.Close()

		user, userErr := dbI.GetUserByUsername(username)
		if userErr != nil {
			http.Error(w, userErr.Error(), http.StatusInternalServerError)
			return
		}

		if user == nil {
			err := cookieHelpers.WriteCookies(w, "", []byte{}, []byte{})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		if !slices.Equal(user.PasswordHash, passwordHash) {
			err := cookieHelpers.WriteCookies(w, "", []byte{}, []byte{})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		next(w, r)
	}
}