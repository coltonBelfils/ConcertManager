package dbInterface

import (
	"ConcertGetApp/appPath"
	"ConcertGetApp/dataTypes"
	"database/sql"
	"github.com/cockroachdb/errors"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

type DbInterface struct {
	db *sql.DB
}

func New() (*DbInterface, error) {
	db, dbErr := sql.Open("sqlite3", appPath.Path(".concertDB"))
	if dbErr != nil {
		return nil, errors.Wrap(dbErr, "error opening database")
	}

	_, execErr := db.Exec(`create table if not exists artist
(
    id      integer     not null
        constraint artist_pk
            primary key autoincrement,
    name    TEXT unique not null,
    addedBy integer
        references user (id)
)
    strict;

create table if not exists venue
(
    id      integer     not null
        constraint venue_pk
            primary key autoincrement,
    name    TEXT unique not null,
    city    TEXT        not null,
    state_country TEXT        not null,
    addedBy integer
        references user (id)
)
    strict;

create table if not exists concert
(
    id       integer                           not null
        constraint concert_pk
            primary key autoincrement,
    artistId integer                           not null
        references artist (id),
    venueId  integer                           not null
        references venue (id),
    url      TEXT unique,
    urlResolves integer default true not null,
    date     integer default current_timestamp not null,
    setlistFmUrl TEXT unique,
    processedStage int default 0 not null,
    filePath TEXT,
    addedBy  integer
        references user (id)
)
    strict;

create table if not exists user
(
    id           integer     not null
        constraint user_pk
            primary key autoincrement,
    username     TEXT unique not null,
    passwordHash BLOB        not null
)
    strict;`)
	if execErr != nil {
		return nil, errors.Wrap(execErr, "error creating tables")
	}

	return &DbInterface{
		db: db,
	}, nil
}

func (d *DbInterface) Close() error {
	return d.db.Close()
}

/*
Access patters:
- get all concerts (select * from concert)
- get all artists (select * from artist)
- get all venues (select * from venue)
- get all users (select * from user)
- get user by username (select * from user where username = ?)

- populate artists (insert into artist (name) values ('name'))

- new user (insert or ignore into user (username, passwordHash) values (?, ?))
- new artist (insert or replace into artist (name, addedBy) values (?, ?))
- new venue (insert or replace into venue (name, city, state_country, addedBy) values (?, ?, ?, ?))
- new concert (insert or replace into concert (artistId, venueId, url, date, setlistFmUrl, addedBy) values (?, ?, ?, ?, ?, ?))
*/

func (d *DbInterface) GetConcerts() ([]dataTypes.Concert, error) {
	rows, queryErr := d.db.Query("SELECT * from concert;")
	if queryErr != nil {
		return nil, errors.Wrap(queryErr, "error querying for concerts")
	}

	var concerts []dataTypes.Concert
	for rows.Next() {
		var (
			c       dataTypes.Concert
			dateRaw int
		)
		scanErr := rows.Scan(&c.Id, &c.ArtistId, &c.VenueId, &c.Url, &dateRaw, &c.SetlistFMUrl, &c.ProcessedStage, &c.FilePath, &c.AddedBy)
		if scanErr != nil {
			return nil, errors.Wrap(scanErr, "error scanning concert")
		}

		date := time.Unix(int64(dateRaw), 0)
		c.Date = &date

		concerts = append(concerts, c)
	}

	return concerts, nil
}

func (d *DbInterface) GetArtists() ([]dataTypes.Artist, error) {
	rows, queryErr := d.db.Query("SELECT * from artist;")
	if queryErr != nil {
		return nil, errors.Wrap(queryErr, "error querying for artists")
	}

	var artists []dataTypes.Artist
	for rows.Next() {
		var a dataTypes.Artist
		scanErr := rows.Scan(&a.Id, &a.Name, &a.AddedBy)
		if scanErr != nil {
			return nil, errors.Wrap(scanErr, "error scanning artist")
		}

		artists = append(artists, a)
	}

	return artists, nil
}

func (d *DbInterface) GetVenues() ([]dataTypes.Venue, error) {
	rows, queryErr := d.db.Query("SELECT * from venue;")
	if queryErr != nil {
		return nil, errors.Wrap(queryErr, "error querying for venues")
	}

	var venues []dataTypes.Venue
	for rows.Next() {
		var v dataTypes.Venue
		scanErr := rows.Scan(&v.Id, &v.Name, &v.City, &v.StateOrCountry, &v.AddedBy)
		if scanErr != nil {
			return nil, errors.Wrap(scanErr, "error scanning venue")
		}

		venues = append(venues, v)
	}

	return venues, nil
}

func (d *DbInterface) GetUsers() ([]dataTypes.User, error) {
	rows, queryErr := d.db.Query("SELECT * from user;")
	if queryErr != nil {
		return nil, errors.Wrap(queryErr, "error querying for users")
	}

	var users []dataTypes.User
	for rows.Next() {
		var u dataTypes.User
		scanErr := rows.Scan(&u.Id, &u.Username, &u.PasswordHash)
		if scanErr != nil {
			return nil, errors.Wrap(scanErr, "error scanning user")
		}

		users = append(users, u)
	}

	return users, nil
}

func (d *DbInterface) GetUserByUsername(username string) (*dataTypes.User, error) {
	rows, queryErr := d.db.Query("SELECT * from user where username = ?;", username)
	if queryErr != nil {
		return nil, errors.Wrap(queryErr, "error querying for user")
	}

	var users []*dataTypes.User

	for rows.Next() {
		u := dataTypes.User{}
		scanErr := rows.Scan(&u.Id, &u.Username, &u.PasswordHash)
		if scanErr != nil {
			return nil, errors.Wrap(scanErr, "error scanning user")
		}

		users = append(users, &u)
	}

	if len(users) == 0 {
		return nil, nil //nolint:nilnil
	} else if len(users) > 1 {
		return nil, errors.New("multiple users found")
	}

	return users[0], nil
}

func (d *DbInterface) GetArtist(id int) (*dataTypes.Artist, error) {
	rows, queryErr := d.db.Query(`SELECT * from artist where id = ?`, id)
	if queryErr != nil {
		return nil, errors.Wrap(queryErr, "error querying for artist")
	}

	return singleArtistRowsHelper(rows)
}

func (d *DbInterface) GetArtistByName(name string) (*dataTypes.Artist, error) {
	rows, queryErr := d.db.Query("SELECT * from artist where name = ?;", name)
	if queryErr != nil {
		return nil, errors.Wrap(queryErr, "error querying for artist")
	}

	return singleArtistRowsHelper(rows)
}

func singleArtistRowsHelper(rows *sql.Rows) (*dataTypes.Artist, error) {
	var artists []*dataTypes.Artist
	for rows.Next() {
		a := dataTypes.Artist{}
		scanErr := rows.Scan(&a.Id, &a.Name, &a.AddedBy)
		if scanErr != nil {
			return nil, errors.Wrap(scanErr, "error scanning artist")
		}

		artists = append(artists, &a)
	}

	if len(artists) == 0 {
		return nil, nil //nolint:nilnil
	} else if len(artists) > 1 {
		return nil, errors.New("multiple artists found")
	}

	return artists[0], nil
}

func (d *DbInterface) GetVenue(id int) (*dataTypes.Venue, error) {
	rows, queryErr := d.db.Query(`SELECT * from venue where id = ?`, id)
	if queryErr != nil {
		return nil, errors.Wrap(queryErr, "error querying for venue")
	}

	return singleVenueRowsHelper(rows)
}

func (d *DbInterface) GetVenueByName(name string) (*dataTypes.Venue, error) {
	rows, queryErr := d.db.Query("SELECT * from venue where name = ?;", name)
	if queryErr != nil {
		return nil, errors.Wrap(queryErr, "error querying for venue")
	}

	return singleVenueRowsHelper(rows)
}

func singleVenueRowsHelper(rows *sql.Rows) (*dataTypes.Venue, error) {
	var venues []*dataTypes.Venue
	for rows.Next() {
		v := dataTypes.Venue{}
		scanErr := rows.Scan(&v.Id, &v.Name, &v.City, &v.StateOrCountry, &v.AddedBy)
		if scanErr != nil {
			return nil, errors.Wrap(scanErr, "error scanning venue")
		}

		venues = append(venues, &v)
	}

	if len(venues) == 0 {
		return nil, nil //nolint:nilnil
	} else if len(venues) > 1 {
		return nil, errors.New("multiple venues found")
	}

	return venues[0], nil
}

func (d *DbInterface) NewUser(username string, passwordHash []byte) error {
	_, execErr := d.db.Exec("insert or ignore into user (username, passwordHash) values (?, ?);", username, passwordHash)
	if execErr != nil {
		return errors.Wrap(execErr, "error inserting user")
	}

	return nil
}

func (d *DbInterface) NewArtist(name string, addedBy int) error {
	_, execErr := d.db.Exec("insert or replace into artist (name, addedBy) values (?, ?);", name, addedBy)
	if execErr != nil {
		return errors.Wrap(execErr, "error inserting artist")
	}

	return nil
}

func (d *DbInterface) NewVenue(name string, city string, stateOrCountry string, addedBy int) error {
	_, execErr := d.db.Exec("insert or replace into venue (name, city, state_country, addedBy) values (?, ?, ?, ?);", name, city, stateOrCountry, addedBy)
	if execErr != nil {
		return errors.Wrap(execErr, "error inserting venue")
	}

	return nil
}

func (d *DbInterface) NewConcert(artistId int, venueId int, url string, date int, setlistFmUrl string, addedBy int) error {
	_, execErr := d.db.Exec("insert or replace into concert (artistId, venueId, url, date, setlistFmUrl, addedBy) values (?, ?, ?, ?, ?, ?);", artistId, venueId, url, date, setlistFmUrl, addedBy)
	if execErr != nil {
		return errors.Wrap(execErr, "error inserting concert")
	}

	return nil
}
