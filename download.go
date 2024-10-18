package main

import (
	"ConcertGetApp/appPath"
	"ConcertGetApp/dataTypes"
	"ConcertGetApp/dbInterface"
	"fmt"
	"github.com/pkg/errors"
	"os/exec"
	"strings"
	"time"
)

func Download(dbI *dbInterface.DbInterface, concert *dataTypes.Concert) error {
	if concert.ArtistId == nil {
		return errors.New("concert artist id is nil")
	}
	artist, getArtistErr := dbI.GetArtist(*concert.ArtistId)
	if getArtistErr != nil {
		return errors.Wrap(getArtistErr, "error getting artist name")
	}
	if artist.Name == nil {
		return errors.New("artist name is nil")
	}
	artistName := *artist.Name
	artistName = strings.ReplaceAll(artistName, "/", "-")

	if concert.VenueId == nil {
		return errors.New("concert venue id is nil")
	}
	venue, getVenueErr := dbI.GetVenue(*concert.VenueId)
	if getVenueErr != nil {
		return errors.Wrap(getVenueErr, "error getting venue name")
	}
	if venue.Name == nil {
		return errors.New("venue name is nil")
	}
	venueName := *venue.Name
	if venue.City != nil {
		venueName += fmt.Sprintf(" - %s", *venue.City)
	}
	if venue.StateOrCountry != nil {
		venueName += fmt.Sprintf(", %s", *venue.StateOrCountry)
	}
	venueName = strings.ReplaceAll(venueName, "/", "-")

	runPath := appPath.Path(fmt.Sprintf("%s/%s %s/01 - %s %s", artistName, concert.Date.Format(time.DateOnly), venueName, concert.Date.Format(time.DateOnly), venueName))
	runPath = strings.TrimSpace(runPath)

	if concert.Url == nil {
		return errors.New("concert url is nil")
	}
	ytdlpCmd := exec.Command("yt-dlp", "-o", runPath, "--write-thumbnail", "--write-description", "--write-info-json", "--write-comments", "--", *concert.Url)

	runErr := ytdlpCmd.Run()
	if runErr != nil {
		return errors.Wrap(runErr, "error running yt-dlp")
	}

	return nil
}
