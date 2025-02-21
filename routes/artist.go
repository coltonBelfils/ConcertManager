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
		fmt.Printf("%+v\n", fmtErr)
		http.Error(w, fmtErr.Error(), http.StatusInternalServerError)
		return
	}

	if user.Id == nil {
		fmtErr := errors.New("user id is nil")
		fmt.Printf("%+v\n", fmtErr)
		http.Error(w, fmtErr.Error(), http.StatusInternalServerError)
		return
	}

	artists, getArtistsErr := dbI.GetArtists()
	if getArtistsErr != nil {
		fmtErr := errors.Wrap(getArtistsErr, "could not get artists when trying to populate artist page")
		fmt.Printf("%+v\n", fmtErr)
		http.Error(w, fmtErr.Error(), http.StatusInternalServerError)
		return
	}

	artistPackage := struct {
		CannotDeleteErr bool
		AffectedArtist  dataTypes.Artist
		User            dataTypes.User
		Artists         []dataTypes.Artist
	}{
		Artists: artists,
	}

	parseErr := r.ParseForm()
	if parseErr != nil {
		fmt.Printf("Error: %+v\n", parseErr)
	}

	crud := r.Form.Get("crud")

	if crud == "create" {
		newArtist := r.Form.Get("newArtist")
		if newArtist == "" {
			http.Error(w, "no artist name provided", http.StatusBadRequest)
			return
		}

		fmt.Printf("New artist: %s", newArtist)

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

		artists, getArtistsErr := dbI.GetArtists()
		if getArtistsErr != nil {
			fmtErr := errors.Wrap(getArtistsErr, "could not get artists when trying to populate artist page")
			fmt.Printf("%+v\n", fmtErr)
			http.Error(w, fmtErr.Error(), http.StatusInternalServerError)
			return
		}

		artistPackage.Artists = artists
	} else if crud == "update" { //nolint

	} else if crud == "delete" {
		artistIdRaw := r.Form.Get("toDeleteArtist")
		if artistIdRaw == "" {
			fmt.Println("toDeleteArtist was empty")
			http.Error(w, "toDeleteArtist was empty", http.StatusBadRequest)
			return
		}

		artistId64, parseErr := strconv.ParseInt(artistIdRaw, 10, 64)
		if parseErr != nil {
			fmtErr := errors.Newf("toDeleteArtist was given with invalid id: %s\n", artistIdRaw)
			fmt.Printf("%+v\n", fmtErr)
			http.Error(w, fmtErr.Error(), http.StatusBadRequest)
			return
		}
		artistId := int(artistId64)

		concerts, getConcertsErr := dbI.GetConcertsByArtist(artistId)
		if getConcertsErr != nil {
			fmtErr := errors.Wrapf(getConcertsErr, "could not verify no concerts use this artist with id: %d\n", artistId)
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
			delErr := dbI.DeleteArtist(artistId)
			if delErr != nil {
				fmtErr := errors.Wrapf(delErr, "error when attempting to delete artist with id: %d\n", artistId)
				fmt.Printf("%+v\n", fmtErr)
				http.Error(w, fmtErr.Error(), http.StatusBadRequest)
				return
			}

			artists, getArtistsErr = dbI.GetArtists()
			if getArtistsErr != nil {
				fmtErr := errors.Wrap(getArtistsErr, "could not get artists when trying to populate artist page")
				fmt.Printf("%+v\n", fmtErr)
				http.Error(w, fmtErr.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			for _, artist := range artists {
				if artist.Id != nil && *artist.Id == artistId {
					// This probably is a temporary solution. Something more fancy later.
					*artist.Name += " -- Cannot delete. Delete affiliated concerts first."
				}
			}
		}

		artistPackage.Artists = artists
	}

	tmpl, err := htmlTemplate.GetTemplate("artist.gohtml")
	if err != nil {
		fmt.Printf("Error: %+v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user.PasswordHash = []byte("redacted")
	user.Id = nil

	artistPackage.User = *user

	err = tmpl.Execute(w, artistPackage) //data will need to be added to actually have data later once the corresponding .gohtml file is further built out.
	if err != nil {
		fmt.Printf("Error: %+v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
