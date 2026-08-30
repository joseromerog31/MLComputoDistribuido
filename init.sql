CREATE TABLE IF NOT EXISTS partidos (
    id VARCHAR(100) PRIMARY KEY,
    equipo_local VARCHAR(100) NOT NULL,
    equipo_visitante VARCHAR(100) NOT NULL,
    goles_equipo_local INTEGER NOT NULL,
    goles_equipo_visitante INTEGER NOT NULL,
    jornada INTEGER NOT NULL,
    estadio VARCHAR(150) NOT NULL
);