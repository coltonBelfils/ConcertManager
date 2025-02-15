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

func Form(w http.ResponseWriter, r *http.Request) {
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
		err := r.ParseForm()
		if err != nil {
			fmt.Printf("Error: %+v\n", err)
			http.Error(w, errors.Wrap(err, "error parsing form").Error(), http.StatusInternalServerError)
			return
		}

		var artist dataTypes.Artist

		artistValue := r.FormValue("selectedArtist")
		if artistValue == "" {
			http.Error(w, "no artist selected", http.StatusBadRequest)
			return
		}

		if artistValue == "newArtist" {
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
				http.Error(w, "failed to find inserted artist", http.StatusInternalServerError)
				return
			}

			artist = *gotArtist
		} else {
			artistId, parseErr := strconv.ParseInt(artistValue, 10, 64)
			if parseErr != nil {
				http.Error(w, parseErr.Error(), http.StatusBadRequest)
				return
			}

			gotArtist, getErr := dbI.GetArtist(int(artistId))
			if getErr != nil {
				fmt.Printf("Error: %+v\n", getErr)
				http.Error(w, getErr.Error(), http.StatusInternalServerError)
				return
			}

			if gotArtist == nil {
				http.Error(w, "selected artist not found", http.StatusNotFound)
				return
			}

			artist = *gotArtist
		}

		venue := dataTypes.Venue{}

		venueValue := r.FormValue("selectedVenue")
		if venueValue == "" {
			http.Error(w, "no venue selected", http.StatusBadRequest)
			return
		}

		if venueValue == "newVenue" {
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
				http.Error(w, "failed to find inserted venue", http.StatusInternalServerError)
				return
			}

			venue = *gotVenue
		} else {
			gotVenue, getErr := dbI.GetVenueByName(venueValue)
			if getErr != nil {
				fmt.Printf("Error: %+v\n", getErr)
				http.Error(w, getErr.Error(), http.StatusInternalServerError)
				return
			}

			if gotVenue == nil {
				http.Error(w, "selected venue not found", http.StatusNotFound)
				return
			}

			venue = *gotVenue
		}

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

		if artist.Id == nil {
			fmt.Printf("Error: %+v\n", errors.New("artist id is nil"))
			http.Error(w, "artist id is nil", http.StatusInternalServerError)
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
	}

	tmpl, err := htmlTemplate.GetTemplate("form.gohtml")
	if err != nil {
		fmt.Printf("Error: %+v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	artists, getErr := dbI.GetArtists()
	if getErr != nil {
		fmt.Printf("Error: %+v\n", getErr)
		http.Error(w, getErr.Error(), http.StatusInternalServerError)
		return
	}

	venues, getErr := dbI.GetVenues()
	if getErr != nil {
		fmt.Printf("Error: %+v\n", getErr)
		http.Error(w, getErr.Error(), http.StatusInternalServerError)
		return
	}

	concerts, getErr := dbI.GetConcerts()
	if getErr != nil {
		fmt.Printf("Error: %+v\n", getErr)
		http.Error(w, getErr.Error(), http.StatusInternalServerError)
		return
	}

	user.PasswordHash = []byte("redacted")
	user.Id = nil

	formPackage := struct {
		Artists  []dataTypes.Artist
		Venues   []dataTypes.Venue
		Concerts []dataTypes.Concert
		User     dataTypes.User
	}{
		Artists:  artists,
		Venues:   venues,
		Concerts: concerts,
		User:     *user,
	}

	err = tmpl.Execute(w, formPackage)
	if err != nil {
		fmt.Printf("Error: %+v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}