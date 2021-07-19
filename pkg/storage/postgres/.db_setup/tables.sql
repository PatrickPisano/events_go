DROP TABLE IF EXISTS events;
CREATE TABLE events
(
    id SERIAL,
    title VARCHAR(64),
    description VARCHAR (512),
    is_virtual BOOLEAN,
    address VARCHAR(128),
    link VARCHAR(128),
    seat_number INT,
    start_time timestamptz,
    end_time timestamptz,
    welcome_message VARCHAR(256),
    is_published BOOLEAN,

    PRIMARY KEY (id)
);