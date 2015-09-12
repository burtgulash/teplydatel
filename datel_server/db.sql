create table players (
    player_id serial primary key,
    name text unique,
    rank integer,
    registered_at timestamp with time zone,
    avg_wpm real
);

create table race_texts (
    race_text_id serial primary key,
    race_text text
);

create type enum_race_status as enum (
    'created',
    'live',
    'finished',
    'canceled',
    'error'
);

create table races (
    race_id bigserial primary key,
    race_code varchar(7),
    status enum_race_status,
    created_time timestamp with time zone,
    start_time timestamp with time zone,
    num_players integer,
    race_text_id integer references race_texts(race_text_id)
);

create table r_races_players (
    race_id bigint references races(race_id),
    player_id integer references players(player_id),
    finishing_wpm real,
    finishing_order smallint,
    primary key (race_id, player_id)
);

