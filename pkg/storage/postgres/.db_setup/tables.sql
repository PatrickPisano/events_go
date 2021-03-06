DROP TABLE IF EXISTS users;
CREATE TABLE users
(
    id SERIAL,
    names VARCHAR(64) NOT NULL,
    email VARCHAR (128) NOT NULL,
    password VARCHAR(60) NOT NULL,
    is_admin BOOLEAN DEFAULT FALSE,

    PRIMARY KEY (id),
    UNIQUE (email)
);

DROP TABLE IF EXISTS events;
CREATE TABLE events
(
    id SERIAL,
    title VARCHAR(64),
    description VARCHAR (512),
    link VARCHAR(128),
    start_time timestamptz,
    end_time timestamptz,
    welcome_message VARCHAR (256),
    cover_image_path VARCHAR (128),
    is_published BOOLEAN,
    host_id INT NOT NULL,

    PRIMARY KEY (id),
    FOREIGN KEY (host_id)
        REFERENCES users (id)
        ON DELETE CASCADE
);

DROP TABLE IF EXISTS event_invitations;
CREATE TABLE event_invitations
(
    email VARCHAR(128),
    event_id INT,
    has_responded BOOLEAN DEFAULT FALSE,
    response BOOLEAN DEFAULT FALSE,
    token CHAR(32),
    responded_at timestamptz,

    UNIQUE (email, event_id),
    FOREIGN KEY (event_id)
        REFERENCES events (id)
        ON DELETE CASCADE
);