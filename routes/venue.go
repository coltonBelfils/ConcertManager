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
)

func Venue(w http.ResponseWriter, r *http.Request) {
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

	venues, getVenuesErr := dbI.GetVenues()
	if getVenuesErr != nil {
		fmtErr := errors.Wrap(getVenuesErr, "could not get venues when trying to populate venue page")
		fmt.Printf("%+v\n", fmtErr)
		http.Error(w, fmtErr.Error(), http.StatusInternalServerError)
		return
	}

	venuePackage := struct {
		CannotDeleteErr bool
		AffectedVenue  dataTypes.Venue
		User            dataTypes.User
		Venues         []dataTypes.Venue
	}{
		Venues: venues,
	}

	parseErr := r.ParseForm()
	if parseErr != nil {
		fmt.Printf("Error: %+v\n", parseErr)
	}

	crud := r.Form.Get("crud")

	if crud == "create" {
		venueName := r.Form.Get("venueName")
		if venueName == "" {
			http.Error(w, "no venue name provided", http.StatusBadRequest)
			return
		}

		venueCity := r.Form.Get("venueCity")
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
	} else if crud == "update" {//nolint

	} else if crud == "delete" {//nolint
		venueIdRaw := r.Form.Get("toDeleteVenue")
		if venueIdRaw == "" {
			fmt.Println("toDeleteVenue was empty")
			http.Error(w, "toDeleteVenue was empty", http.StatusBadRequest)
			return
		}

		venueId64, parseErr := strconv.ParseInt(venueIdRaw, 10, 64)
		if parseErr != nil {
			fmtErr := errors.Newf("toDeleteVenue was given with invalid id: %s\n", venueIdRaw)
			fmt.Printf("%+v\n", fmtErr)
			http.Error(w, fmtErr.Error(), http.StatusBadRequest)
			return
		}
		venueId := int(venueId64)

		concerts, getConcertsErr := dbI.GetConcertsByVenue(venueId)
		if getConcertsErr != nil {
			fmtErr := errors.Wrapf(getConcertsErr, "could not verify no concerts use this venue with id: %d\n", venueId)
			fmt.Printf("%+v\n", fmtErr)
			http.Error(w, fmtErr.Error(), http.StatusInternalServerError)
			return
		}

		canDelete := true

		if len(concerts) != 0 {
			canDelete = false

			//fmtErr := errors.Newf("the artist that is being deleted is in use in a concert. Those concerts need to be deleted. ArtistId: %d", artistId)
			//fmt.Printf("%+v\n", fmtErr)
			//http.Error(w, fmtErr.Error(), http.StatusBadRequest)
			//return
		}

		if canDelete {
			delErr := dbI.DeleteVenue(venueId)
			if delErr != nil {
				fmtErr := errors.Wrapf(delErr, "error when attempting to delete venue with id: %d\n", venueId)
				fmt.Printf("%+v\n", fmtErr)
				http.Error(w, fmtErr.Error(), http.StatusBadRequest)
				return
			}

			venues, getVenuesErr = dbI.GetVenues()
			if getVenuesErr != nil {
				fmtErr := errors.Wrap(getVenuesErr, "could not get artists when trying to populate venue page")
				fmt.Printf("%+v\n", fmtErr)
				http.Error(w, fmtErr.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			for _, venue := range venues {
				if venue.Id != nil && *venue.Id == venueId {
					// This probably is a temporary solution. Something more fancy later.
					*venue.Name += " -- Cannot delete. Delete affiliated concerts first."
				}
			}
		}

		venuePackage.Venues = venues
	}

	for _, venue := range venuePackage.Venues {
		fmt.Printf("%s\n", *venue.Name)
	}

	tmpl, err := htmlTemplate.GetTemplate("venue.gohtml")
	if err != nil {
		fmt.Printf("Error: %+v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, venuePackage)
	if err != nil {
		fmt.Printf("Error: %+v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}