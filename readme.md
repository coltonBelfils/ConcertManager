# Concert Manager

## File Structure

<pre>
┌──────────┐               ┌─────────────────┐
│ App Root │            ─ ─│    sqlite db    │
└──────────┘           │   └─────────────────┘
      │  ┌─────────┐            ┌─────────────────────────────┐
      ├─▶│ records │◀─ ┘        │ This is the canonical place │
      │  └─────────┘         ┌ ─│ where concert audio, video, │
      │  ┌────────────────┐     │     and data is stored.     │
      ├─▶│    storage     │◀ ┘  └─────────────────────────────┘
      │  └────────────────┘                     ┌──────────────────────┐
      │           │ ┌──────────────────────┐    │ The concert id comes │
      │           └▶│ concert [concert id] │◀─ ─│ from the records db  │
      │             └──────────────────────┘    └──────────────────────┘
      │                         │
      │                         │      ┌─────────────────┐
      │                         ├─────▶│    video.mp4    │
      │                         │      ├─────────────────┤
      │                         ├─────▶│    audio.mp3    │
      │                         │      ├─────────────────┤
      │                         ├─────▶│ description.txt │◀ ─ ┐    ┌──────────────────────────┐
      │                         │      ├─────────────────┤         │  These three files come  │
      │                         ├─────▶│    info.json    │◀ ─ ┼ ─ ─│   from yt-dlp/Youtube    │
      │                         │      ├─────────────────┤         └──────────────────────────┘
      │                         ├─────▶│  thumbnail.png  │◀ ─ ┘    ┌─────────────────────────────────────┐
      │                         │      ├─────────────────┤         │ This file mirrors the data found in │
      │                         └─────▶│  details.json   │◀ ─ ─ ─ ─│ its concert row in the records db.  │
      │                                └─────────────────┘         └─────────────────────────────────────┘
      │  ┌──────┐
      └─▶│ Plex │
         └──────┘
             │  ┌────────────┐
             └─▶│   Media    │                                          ┌────────────────────────────────┐
                └────────────┘                                          │ These are both simlinks to the │
                       │ ┌───────────────┐                              │ corresponding files in Concert │─ ─ ─ ─ ─ ─ ─ ─ ─ ─
                       └▶│     Music     │                              │    Manager/storage/ folder.    │           │       │
                         └───────────────┘                              └────────────────────────────────┘
                                 │  ┌───────────────┐                                                                │       │
                                 └─▶│ [Artist Name] │
                                    └───────────────┘                                                                │       │
                                            │  ┌─────────────────────────────────────┐
                                            └─▶│ yyyy-mm-dd venue city,country/state │                               │       │
                                               └─────────────────────────────────────┘
                                                                  │  ┌────────────────────────────────────────────┐  │       │
                                                                  ├─▶│ 01-yyyy-mm-dd venue city,country/state.mp3 │◀─
                                                                  │  └────────────────────────────────────────────┘          │
                                                                  │  ┌────────────────────────────────────────────────────┐
                                                                  └─▶│ 01-yyyy-mm-dd venue city,country/state-concert.mp4 │◀ ┘
                                                                     └────────────────────────────────────────────────────┘
</pre>

## records db structure

<pre>
create table sqlite_master
(
    type     TEXT,
    name     TEXT,
    tbl_name TEXT,
    rootpage INT,
    sql      TEXT
);

create table sqlite_sequence
(
    name,
    seq
);

create table user
(
    id           integer not null
        constraint user_pk
            primary key autoincrement,
    username     TEXT    not null
        unique,
    passwordHash BLOB    not null
)
    strict;

create table artist
(
    id      integer not null
        constraint artist_pk
            primary key autoincrement,
    name    TEXT    not null
        unique,
    addedBy integer
        references user
)
    strict;

create table venue
(
    id      integer not null
        constraint venue_pk
            primary key autoincrement,
    name    TEXT    not null
        unique,
    city    TEXT    not null,
    country TEXT    not null,
    addedBy integer
        references user
)
    strict;

create table concert
(
    id             integer                           not null
        constraint concert_pk
            primary key autoincrement,
    artistId       integer                           not null
        references artist,
    venueId        integer                           not null
        references venue,
    url            TEXT
        unique,
    date           integer default current_timestamp not null,
    setlistFmUrl   TEXT
        unique,
    processedStage int     default 0                 not null,
    filePath       TEXT,
    addedBy        integer
        references user
)
    strict;
</pre>