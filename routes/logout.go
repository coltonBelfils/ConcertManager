package routes

import (
	"ConcertGetApp/helpers/cookieHelpers"
	"fmt"
	"net/http"
)

func Logout(w http.ResponseWriter, r *http.Request) {
	err := cookieHelpers.WriteCookies(w, "", []byte{}, []byte{})
	if err != nil {
		fmt.Printf("Error: %+v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}