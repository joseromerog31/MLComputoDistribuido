DROP TABLE IF EXISTS partidos;

CREATE TABLE partidos (
    id BIGINT PRIMARY KEY,

    referee VARCHAR(150),
    timezone VARCHAR(100),
    date TIMESTAMPTZ NOT NULL,

    venue_id BIGINT,
    venue_name VARCHAR(200),
    venue_city VARCHAR(150),

    season INTEGER NOT NULL,
    round VARCHAR(100),

    home_team VARCHAR(150) NOT NULL,
    away_team VARCHAR(150) NOT NULL,

    home_win BOOLEAN,
    away_win BOOLEAN,

    home_goals INTEGER,
    away_goals INTEGER,

    home_goals_half_time INTEGER,
    away_goals_half_time INTEGER,

    home_goals_fulltime INTEGER,
    away_goals_fulltime INTEGER,

    home_goals_extra_time INTEGER,
    away_goals_extratime INTEGER,

    home_goals_penalty INTEGER,
    away_goals_penalty INTEGER
);