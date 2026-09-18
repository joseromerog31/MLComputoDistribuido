CREATE TEMP TABLE partidos_staging (
    id TEXT,
    referee TEXT,
    timezone TEXT,
    date TEXT,
    venue_id TEXT,
    venue_name TEXT,
    venue_city TEXT,
    season TEXT,
    round TEXT,
    home_team TEXT,
    away_team TEXT,
    home_win TEXT,
    away_win TEXT,
    home_goals TEXT,
    away_goals TEXT,
    home_goals_half_time TEXT,
    away_goals_half_time TEXT,
    home_goals_fulltime TEXT,
    away_goals_fulltime TEXT,
    home_goals_extra_time TEXT,
    away_goals_extratime TEXT,
    home_goals_penalty TEXT,
    away_goals_penalty TEXT
);

COPY partidos_staging
FROM '/data/liga_mx.csv'
WITH (
    FORMAT CSV,
    HEADER TRUE,
    DELIMITER ','
);

INSERT INTO partidos (
    id,
    referee,
    timezone,
    date,
    venue_id,
    venue_name,
    venue_city,
    season,
    round,
    home_team,
    away_team,
    home_win,
    away_win,
    home_goals,
    away_goals,
    home_goals_half_time,
    away_goals_half_time,
    home_goals_fulltime,
    away_goals_fulltime,
    home_goals_extra_time,
    away_goals_extratime,
    home_goals_penalty,
    away_goals_penalty
)
SELECT
    id::BIGINT,

    NULLIF(referee, ''),
    NULLIF(timezone, ''),
    date::TIMESTAMPTZ,

    NULLIF(venue_id, '')::NUMERIC::BIGINT,

    NULLIF(venue_name, ''),
    NULLIF(venue_city, ''),

    season::INTEGER,

    NULLIF(round, ''),

    home_team,
    away_team,

    NULLIF(home_win, '')::BOOLEAN,
    NULLIF(away_win, '')::BOOLEAN,

    NULLIF(home_goals, '')::NUMERIC::INTEGER,
    NULLIF(away_goals, '')::NUMERIC::INTEGER,

    NULLIF(home_goals_half_time, '')::NUMERIC::INTEGER,
    NULLIF(away_goals_half_time, '')::NUMERIC::INTEGER,

    NULLIF(home_goals_fulltime, '')::NUMERIC::INTEGER,
    NULLIF(away_goals_fulltime, '')::NUMERIC::INTEGER,

    NULLIF(home_goals_extra_time, '')::NUMERIC::INTEGER,
    NULLIF(away_goals_extratime, '')::NUMERIC::INTEGER,

    NULLIF(home_goals_penalty, '')::NUMERIC::INTEGER,
    NULLIF(away_goals_penalty, '')::NUMERIC::INTEGER

FROM partidos_staging;