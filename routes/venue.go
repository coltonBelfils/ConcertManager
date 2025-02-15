package routes

import (
	"ConcertGetApp/dbInterface"
	"ConcertGetApp/helpers/cookieHelpers"
	"fmt"
	"github.com/cockroachdb/errors"
	"net/http"
)

func NewVenue(w http.ResponseWriter, r *http.Request) {
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
		venueName := r.FormValue("venueName")
		if venueName == "" {
			http.Error(w, "no venue name provided", http.StatusBadRequest)
			return
		}

		venueCity := r.FormValue("venueCity")
		if venueCity == "" {
			http.Error(w, "no city provided", http.StatusBadRequest)
			return
		}

		venueCountry := r.FormValue("venueCountry")
		if venueCountry == "" {
			http.Error(w, "no country provided", http.StatusBadRequest)
			return
		}

		addErr := dbI.NewVenue(venueName, venueCity, venueCountry, *user.Id)
		if addErr != nil {
			fmt.Printf("Error: %+v\n", addErr)
			http.Error(w, addErr.Error(), http.StatusInternalServerError)
			return
		}

		gotVenue, getErr := dbI.GetVenueByName(venueName)
		if getErr != nil {
			fmt.Printf("Error: %+v\n", getErr)
			http.Error(w, getErr.Error(), http.StatusInternalServerError)
			return
		}

		if gotVenue == nil {
			fmt.Printf("Error: %+v", errors.New("failed to find inserted venue"))
			http.Error(w, "failed to find inserted venue", http.StatusInternalServerError)
			return
		}
	} else if r.Method == http.MethodPut {//nolint

	} else if r.Method == http.MethodPatch {//nolint
		// probably not but idk. Probably this or PUT but not both
	} else if r.Method == http.MethodDelete {//nolint

	}

	//populate the rest of the page, what would be the GET method since that will happen after all the other methods as well
}