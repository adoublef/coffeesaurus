CREATE TABLE profiles (
    `id` TEXT,
    `login` TEXT NOT NULL UNIQUE,
    `photo_url` TEXT,
    `name` TEXT,
    --
    PRIMARY KEY (id)
);

CREATE TABLE credentials (
    `oauth` TEXT,
    `profile` TEXT,
    --
    FOREIGN KEY (profile) REFERENCES profiles (id),
    PRIMARY KEY (oauth)
);