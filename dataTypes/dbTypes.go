package dataTypes

import (
	"time"
)

/*
create table if not exists artist
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
    country TEXT        not null,
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
    date     integer default current_timestamp not null,
    addedBy  integer
        references user (id)
)
    strict;

create table if not exists user
(
    id           integer not null
        constraint user_pk
            primary key autoincrement,
    username     TEXT    not null,
    passwordHash TEXT    not null
)
    strict;
*/

type Artist struct {
	Id      *int    `json:"id"`
	Name    *string `json:"name"`
	AddedBy *int    `json:"addedBy"`
}

type Venue struct {
	Id             *int    `json:"id"`
	Name           *string `json:"name"`
	City           *string `json:"city"`
	StateOrCountry *string `json:"state_country"`
	AddedBy        *int    `json:"addedBy"`
}

type Concert struct {
	Id             *int       `json:"id"`
	ArtistId       *int       `json:"artistId"`
	VenueId        *int       `json:"venueId"`
	Url            *string    `json:"url"`
	Date           *time.Time `json:"date"`
	SetlistFMUrl   *string    `json:"setlistFmUrl"`
	ProcessedStage *int       `json:"processedStage"`
	FilePath       *string    `json:"filePath"`
	AddedBy        *int       `json:"addedBy"`
}

type User struct {
	Id           *int    `json:"id"`
	Username     *string `json:"username"`
	PasswordHash []byte  `json:"passwordHash"`
}
