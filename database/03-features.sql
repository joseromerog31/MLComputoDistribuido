CREATE VIEW ml_features AS
SELECT
    id,
    date,
    season,
    home_team,
    away_team,

    AVG(home_goals_fulltime) OVER (
        PARTITION BY home_team
        ORDER BY date, id
        ROWS BETWEEN 5 PRECEDING AND 1 PRECEDING
    ) AS home_avg_goals_last5,

    home_goals_fulltime

FROM partidos;