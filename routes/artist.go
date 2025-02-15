package routes

import (
	"ConcertGetApp/dbInterface"
	"ConcertGetApp/helpers/cookieHelpers"
	"ConcertGetApp/htmlTemplate"
	"fmt"
	"github.com/cockroachdb/errors"
	"net/http"
)

func Artist(w http.ResponseWriter, r *http.Request) {
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

	if user.Id == nil {
		fmtErr := errors.New("user id is nil")
		fmt.Println(fmtErr)
		http.Error(w, fmtErr.Error(), http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodPost {
		parseErr := r.ParseForm()
		if parseErr != nil {
			fmt.Printf("Error: %+v\n", parseErr)
		}

		newArtist := r.FormValue("newArtist")
		if newArtist == "" {
			http.Error(w, "no artist name provided", http.StatusBadRequest)
			return
		}

		addErr := dbI.NewArtist(newArtist, *user.Id)
		if addErr != nil {
			fmt.Printf("Error: %+v\n", addErr)
			http.Error(w, addErr.Error(), http.StatusInternalServerError)
			return
		}

		gotArtist, getErr := dbI.GetArtistByName(newArtist)
		if getErr != nil {
			fmt.Printf("Error: %+v\n", getErr)
			http.Error(w, getErr.Error(), http.StatusInternalServerError)
			return
		}

		if gotArtist == nil {
			fmt.Printf("Error: %+v", errors.New("failed to find inserted artist"))
			http.Error(w, "failed to find inserted artist", http.StatusInternalServerError)
			return
		}
	} else if r.Method == http.MethodPut {//nolint

	} else if r.Method == http.MethodPatch {//nolint
		// probably not but idk. Probably this or PUT but not both
	} else if r.Method == http.MethodDelete {//nolint

	}

	tmpl, err := htmlTemplate.GetTemplate("artist.gohtml")
	if err != nil {
		fmt.Printf("Error: %+v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, nil)//data will need to be added to actually have data later once the corresponding .gohtml file is further built out.
	if err != nil {
		fmt.Printf("Error: %+v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}