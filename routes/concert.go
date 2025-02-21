package routes

import (
	"ConcertGetApp/dataTypes"
	"ConcertGetApp/dbInterface"
	"ConcertGetApp/helpers/cookieHelpers"
	"ConcertGetApp/htmlTemplate"
	"fmt"
	"github.com/cockroachdb/errors"
	"net/http"
	"strconv"
	"time"
)

func Concert(w http.ResponseWriter, r *http.Request) {
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
		concertUrl := r.FormValue("concertUrl")
		if concertUrl == "" {
			http.Error(w, "no concert url provided", http.StatusBadRequest)
			return
		}

		concertDateRaw := r.FormValue("concertDate")
		if concertDateRaw == "" {
			http.Error(w, "no concert date provided", http.StatusBadRequest)
			return
		}

		concertDate, parseErr := time.Parse(time.DateOnly, concertDateRaw)
		if parseErr != nil {
			http.Error(w, parseErr.Error(), http.StatusBadRequest)
			return
		}

		concertSetlistFmUrlRaw := r.FormValue("concertSetlistFmUrl")
		var concertSetlistFmUrl string
		if concertSetlistFmUrlRaw != "" {
			concertSetlistFmUrl = concertSetlistFmUrlRaw
		}

		artist, getArtistErr, errorCode := getArtist(r, dbI)
		if getArtistErr != nil {
			fmt.Printf("Error: %+v", getArtistErr)
			http.Error(w, getArtistErr.Error(), errorCode)
			return
		}

		if artist.Id == nil {
			fmt.Printf("Error: %+v\n", errors.New("artist id is nil"))
			http.Error(w, "artist id is nil", http.StatusInternalServerError)
			return
		}

		venue, getVenueErr, errorCode := getVenue(r, dbI)
		if getVenueErr != nil {
			fmt.Printf("Error: %+v", getVenueErr)
			http.Error(w, getVenueErr.Error(), errorCode)
			return
		}

		if venue.Id == nil {
			fmt.Printf("Error: %+v\n", errors.New("venue id is nil"))
			http.Error(w, "venue id is nil", http.StatusInternalServerError)
			return
		}

		addErr := dbI.NewConcert(*artist.Id, *venue.Id, concertUrl, int(concertDate.Unix()), concertSetlistFmUrl, *user.Id)
		if addErr != nil {
			fmt.Printf("Error: %+v\n", addErr)
			http.Error(w, addErr.Error(), http.StatusInternalServerError)
			return
		}
	} else if r.Method == http.MethodPut {//nolint

	} else if r.Method == http.MethodPatch {//nolint
		// probably not but idk. Probably this or PUT but not both
	} else if r.Method == http.MethodDelete {//nolint

	}

	//populate the rest of the page, what would be the GET method since that will happen after all the other methods as well

	tmpl, err := htmlTemplate.GetTemplate("concert.gohtml")
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

func getArtist(r *http.Request, dbI *dbInterface.DbInterface) (dataTypes.Artist, error, int) {
	artistValue := r.FormValue("selectedArtist")
	if artistValue == "" {
		return dataTypes.Artist{}, errors.New("no selected artist"), http.StatusBadRequest
	}

	artistId, parseErr := strconv.ParseInt(artistValue, 10, 64)
	if parseErr != nil {
		return dataTypes.Artist{}, errors.Wrap(parseErr, "could not parse selectedArtist form value. Given value " + artistValue), http.StatusInternalServerError
	}

	gotArtist, getErr := dbI.GetArtist(int(artistId))
	if getErr != nil {
		return dataTypes.Artist{}, errors.Wrap(getErr, fmt.Sprintf("could not find artist with id: %d", artistId)), http.StatusNotFound
	}

	if gotArtist == nil {
		return dataTypes.Artist{}, errors.Newf("after artist creation, could not find artist with id: %d", artistId), http.StatusInternalServerError
	}

	return *gotArtist, nil, http.StatusOK
}

func getVenue(r *http.Request, dbI *dbInterface.DbInterface) (dataTypes.Venue, error, int) {
	venueValue := r.FormValue("selectedVenue")
	if venueValue == "" {
		return dataTypes.Venue{}, errors.New("no selected venue"), http.StatusBadRequest
	}

	venueId, parseErr := strconv.ParseInt(venueValue, 10, 64)
	if parseErr != nil {
		return dataTypes.Venue{}, errors.Wrap(parseErr, fmt.Sprintf("could not parse selectedVenue form value. Given value: %s", venueValue)), http.StatusInternalServerError
	}

	gotVenue, getErr := dbI.GetVenue(int(venueId))
	if getErr != nil {
		return dataTypes.Venue{}, errors.Wrap(getErr, fmt.Sprintf("could not find artist with id: %d", venueId)), http.StatusNotFound
	}

	if gotVenue == nil {
		return dataTypes.Venue{}, errors.Newf("after venue creation, could not find venue with id: %d", venueId), http.StatusInternalServerError
	}

	return *gotVenue, nil, 200
}