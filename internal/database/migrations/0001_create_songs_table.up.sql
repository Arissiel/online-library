CREATE TABLE IF NOT EXISTS artists (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    bio TEXT
);

CREATE TABLE IF NOT EXISTS songs (
    id SERIAL PRIMARY KEY,
    artist_id INT NOT NULL,
    name VARCHAR(255) NOT NULL,
    lyrics TEXT,
    genre VARCHAR(100),
    video TEXT,
    release_date DATE,
    FOREIGN KEY (artist_id) REFERENCES artists(id) ON DELETE CASCADE
);
