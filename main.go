package main

import (
	"ConcertGetApp/dataTypes"
	"ConcertGetApp/dbInterface"
	"ConcertGetApp/htmlTemplate"
	"crypto"
	"encoding/base64"
	"fmt"
	"github.com/pkg/errors"
	"net/http"
	"os"
	"slices"
	"strconv"
	"time"
)

var dbI *dbInterface.DbInterface

func main() {
	var newErr error
	dbI, newErr = dbInterface.New()
	if newErr != nil {
		panic(errors.Wrap(newErr, "Received an error when creating a new dbInterface instance"))
	}
	defer dbI.Close()

	admin, getErr := dbI.GetUserByUsername("admin")
	if getErr != nil {
		panic(errors.Wrap(getErr, "Received an error while getting admin user"))
	}

	adminPassword, ok := os.LookupEnv("ADMIN_PASSWORD")
	if !ok {
		panic("The ADMIN_PASSWORD env var is not set.")
	}

	adminHash, hashErr := HashPassword(adminPassword, "admin")
	if hashErr != nil {
		panic(errors.Wrap(hashErr, "Could not hash admin password"))
	}

	if admin == nil {
		newUserErr := dbI.NewUser("admin", adminHash)
		if newUserErr != nil {
			panic(errors.Wrap(newUserErr, "Could not make admin user"))
		}
	} else if !slices.Equal(admin.PasswordHash, adminHash) {
		panic("the admin user already exists and for one reason or another the current password does not match the ADMIN_PASSWORD env var")
	}

	// Move the funcs for these to the /routes package
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/new-user", newUser)
	http.HandleFunc("/form", LoginCheck(form))
	http.HandleFunc("/validurl", LoginCheck(func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.Query().Get("url")
		resp, getErr := http.Get(url)
		if getErr != nil {
			w.WriteHeader(400)
			return
		}

		w.WriteHeader(resp.StatusCode)
	}))
	http.HandleFunc("/", LoginCheck(index))

	fmt.Printf("Server Running\n\n")

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

func ReadCookies(r *http.Request) (string, []byte, []byte, error) {
	username64, err := r.Cookie("username")
	if err != nil {
		return "", nil, nil, errors.Wrap(err, "error reading username cookie")
	}

	username, decodeErr := base64.StdEncoding.DecodeString(username64.Value)
	if decodeErr != nil {
		return "", nil, nil, errors.Wrap(decodeErr, "error decoding username")
	}

	passwordHash64, err := r.Cookie("passwordHash")
	if err != nil {
		return "", nil, nil, errors.Wrap(err, "error reading passwordHash cookie")
	}

	passwordHash, decodeErr := base64.StdEncoding.DecodeString(passwordHash64.Value)
	if decodeErr != nil {
		return "", nil, nil, errors.Wrap(decodeErr, "error decoding passwordHash")
	}

	check, err := r.Cookie("check")
	if err != nil {
		return "", nil, nil, errors.Wrap(err, "error reading check cookie")
	}

	checkDecode, checkDecodeErr := base64.StdEncoding.DecodeString(check.Value)
	if checkDecodeErr != nil {
		return "", nil, nil, errors.Wrap(checkDecodeErr, "error decoding check")
	}

	return string(username), passwordHash, checkDecode, nil
}

func WriteCookies(w http.ResponseWriter, username string, passwordHash []byte, check []byte) error {
	username64 := base64.StdEncoding.EncodeToString([]byte(username))
	passwordHash64 := base64.StdEncoding.EncodeToString(passwordHash)
	check64 := base64.StdEncoding.EncodeToString(check)

	http.SetCookie(w, &http.Cookie{
		Name:  "username",
		Value: username64,
	})
	http.SetCookie(w, &http.Cookie{
		Name:  "passwordHash",
		Value: passwordHash64,
	})
	http.SetCookie(w, &http.Cookie{
		Name:  "check",
		Value: check64,
	})

	return nil
}

func CookieCheck(username string, passwordHash []byte) ([]byte, error) {
	hasher := crypto.SHA256.New()

	cookieCheckSalt, ok := os.LookupEnv("COOKIE_CHECK_SALT")
	if !ok {
		return nil, errors.New("could not validate login")
	}

	for i := range 11 + len(cookieCheckSalt) + len(username) {
		hasher.Write([]byte(strconv.Itoa(i)))
		hasher.Write([]byte(username))
		hasher.Write(passwordHash)
		hasher.Write([]byte(cookieCheckSalt))
	}

	return hasher.Sum(nil), nil
}

func LoginCheck(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		username, passwordHash, check, readErr := ReadCookies(r)
		if readErr != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		computedCheck, checkErr := CookieCheck(username, passwordHash)
		if checkErr != nil {
			http.Error(w, checkErr.Error(), http.StatusInternalServerError)
			return
		}

		if !slices.Equal(check, computedCheck) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		user, userErr := dbI.GetUserByUsername(username)
		if userErr != nil {
			http.Error(w, userErr.Error(), http.StatusInternalServerError)
			return
		}

		if user == nil {
			err := WriteCookies(w, "", []byte{}, []byte{})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		if !slices.Equal(user.PasswordHash, passwordHash) {
			err := WriteCookies(w, "", []byte{}, []byte{})
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

func HashPassword(password string, username string) ([]byte, error) {
	hasher := crypto.SHA256.New()

	passwordSalt, ok := os.LookupEnv("PASSWORD_SALT")
	if !ok {
		return []byte{}, errors.New("could not hash password")
	}

	for i := range 13 + len(passwordSalt) + len(username) {
		hasher.Write([]byte(strconv.Itoa(i)))
		hasher.Write([]byte(password))
		hasher.Write([]byte(username))
		hasher.Write([]byte(passwordSalt))
	}

	return hasher.Sum(nil), nil
}

func login(w http.ResponseWriter, r *http.Request) {
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

		hash, hashErr := HashPassword(password, username)
		if hashErr != nil {
			fmt.Printf("Error: %+v\n", hashErr)
			http.Error(w, hashErr.Error(), http.StatusInternalServerError)
			return
		}

		if !slices.Equal(user.PasswordHash, hash) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		check, checkErr := CookieCheck(username, hash)
		if checkErr != nil {
			fmt.Printf("Error: %+v\n", checkErr)
			http.Error(w, checkErr.Error(), http.StatusInternalServerError)
			return
		}

		writeErr := WriteCookies(w, username, hash, check)
		if writeErr != nil {
			fmt.Printf("Error: %+v\n", writeErr)
			http.Error(w, writeErr.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func logout(w http.ResponseWriter, r *http.Request) {
	err := WriteCookies(w, "", []byte{}, []byte{})
	if err != nil {
		fmt.Printf("Error: %+v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func newUser(w http.ResponseWriter, r *http.Request) {
	username, _, _, readErr := ReadCookies(r)
	if readErr != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

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

		hashed, hashErr := HashPassword(newPassword, newUsername)
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

func index(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/form", http.StatusSeeOther)

	//indexFile, err := os.OpenFile("./static/index.gohtml", os.O_RDONLY, 0644)
	//if err != nil {
	//	fmt.Printf("Error: %+v\n", err)
	//	http.Error(w, err.Error(), http.StatusInternalServerError)
	//	return
	//}
	//_, err = io.Copy(w, indexFile)
	//if err != nil {
	//	fmt.Printf("Error: %+v\n", err)
	//	http.Error(w, err.Error(), http.StatusInternalServerError)
	//	return
	//}
}

func form(w http.ResponseWriter, r *http.Request) {
	username, _, _, readErr := ReadCookies(r)
	if readErr != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

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
