package main

import (
	"ConcertGetApp/appPath"
	"ConcertGetApp/dbInterface"
	"ConcertGetApp/helpers"
	"ConcertGetApp/middleware"
	"ConcertGetApp/routes"
	"fmt"
	"github.com/cockroachdb/errors"
	"net/http"
	"os"
	"slices"
)

var dbI *dbInterface.DbInterface

func main() {
	_, statErr := os.Stat(appPath.Path("storage"))
	if os.IsNotExist(statErr) {
		fmt.Printf("Making storage at %s", appPath.Path("storage"))
		makeStorageErr := os.Mkdir(appPath.Path("storage"), 0775)
		if makeStorageErr != nil {
			panic(errors.Wrap(makeStorageErr, fmt.Sprintf("Attempted to make directory at %s. Failed.", appPath.Path("storage"))))
		}
	}

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

	adminHash, hashErr := helpers.HashPassword(adminPassword, "admin")
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
	http.HandleFunc("/login", routes.Login)
	http.HandleFunc("/logout", routes.Logout)
	http.HandleFunc("/new-user", routes.NewUser)
	http.HandleFunc("/artist", middleware.LoginCheck(routes.Artist))
	http.HandleFunc("/form", middleware.LoginCheck(routes.Form))
	http.HandleFunc("/validurl", middleware.LoginCheck(func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.Query().Get("url")
		resp, getErr := http.Get(url)
		if getErr != nil {
			w.WriteHeader(400)
			return
		}

		w.WriteHeader(resp.StatusCode)
	}))
	http.HandleFunc("/index", middleware.LoginCheck(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/artist", http.StatusSeeOther)
	}))
	http.HandleFunc("/", middleware.LoginCheck(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/artist", http.StatusSeeOther)
	}))

	fmt.Printf("*Server Running*\n")

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
