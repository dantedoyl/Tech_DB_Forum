CREATE EXTENSION IF NOT EXISTS citext;

CREATE UNLOGGED TABLE users
(
    nickname    citext     PRIMARY KEY,
    fullname    text       NOT NULL,
    about       text,
    email       citext     UNIQUE
);

CREATE UNLOGGED TABLE forum
(
    slug        citext      PRIMARY KEY,
    author      citext,
    title       text        NOT NULL,
    posts       BIGINT      DEFAULT 0,
    threads     INT         DEFAULT 0
);

CREATE UNLOGGED TABLE thread
(
    id          SERIAL      PRIMARY KEY,
    title       text        not null,
    author      citext      REFERENCES users (nickname),
    created     timestamp with time zone        default now(),
    forum       citext      REFERENCES forum (slug),
    message     text        NOT NULL,
    slug        citext      UNIQUE,
    votes       INT         default 0
);

CREATE UNLOGGED TABLE post
(
    id          BIGSERIAL   PRIMARY KEY,
    author      citext,
    created     timestamp with time zone    default now(),
    forum       citext,
    isEdited    BOOLEAN     DEFAULT FALSE,
    message     text        NOT NULL,
    parent      BIGINT      DEFAULT 0,
    thread      INT,
    route       BIGINT[]    DEFAULT ARRAY []::INTEGER[],

    FOREIGN KEY (thread) REFERENCES thread (id),
    FOREIGN KEY (author) REFERENCES users  (nickname)
);

CREATE UNLOGGED TABLE votes
(
    id          BIGSERIAL   PRIMARY KEY,
    author      citext      REFERENCES users (nickname),
    voice       INT         NOT NULL,
    thread_id   INT,

    UNIQUE (author, thread_id),
    FOREIGN KEY (thread_id) REFERENCES thread (id)
);

CREATE UNLOGGED TABLE forum_users
(
    nickname    citext      NOT NULL,
    fullname    TEXT        NOT NULL,
    about       TEXT,
    email       CITEXT,
    slug        citext      NOT NULL,

    UNIQUE (nickname, slug),
    FOREIGN KEY (nickname) REFERENCES users (nickname),
    FOREIGN KEY (slug) REFERENCES forum (slug)
);

CREATE OR REPLACE FUNCTION count_threads() RETURNS TRIGGER AS
$count_threads$
BEGIN
    UPDATE forum SET threads=threads+1 WHERE slug=NEW.forum;
    RETURN NEW;
END
$count_threads$ LANGUAGE plpgsql;

CREATE TRIGGER count_threads
    BEFORE INSERT
    ON thread
    FOR EACH ROW
    EXECUTE PROCEDURE count_threads();

CREATE OR REPLACE FUNCTION insert_new_forum_user() RETURNS TRIGGER AS
$insert_new_forum_user$
DECLARE
    temp_fullname CITEXT;
    temp_about    TEXT;
    temp_email CITEXT;
BEGIN
SELECT fullname, about, email FROM users WHERE nickname = NEW.author INTO temp_fullname, temp_about, temp_email;
INSERT INTO forum_users (nickname, fullname, about, email, slug)
    VALUES (New.Author, temp_fullname, temp_about, temp_email, NEW.forum) on conflict do nothing;
RETURN NEW;
END
$insert_new_forum_user$ LANGUAGE plpgsql;

CREATE TRIGGER insert_new_forum_user
    AFTER INSERT
    ON post
    FOR EACH ROW
    EXECUTE PROCEDURE insert_new_forum_user();

CREATE TRIGGER insert_new_forum_user
    AFTER INSERT
    ON thread
    FOR EACH ROW
    EXECUTE PROCEDURE insert_new_forum_user();

CREATE OR REPLACE FUNCTION add_vote() RETURNS TRIGGER AS
$add_vote$
BEGIN
    UPDATE thread SET votes= votes+NEW.voice WHERE id=NEW.thread_id;
    RETURN NEW;
END
$add_vote$ LANGUAGE plpgsql;

CREATE TRIGGER add_vote
    AFTER INSERT
    ON votes
    FOR EACH ROW
    EXECUTE PROCEDURE add_vote();

CREATE OR REPLACE FUNCTION update_vote() RETURNS TRIGGER AS
$update_vote$
BEGIN
    IF OLD.Voice <> NEW.Voice THEN
        UPDATE thread SET votes = votes + NEW.voice*2 WHERE id=NEW.thread_id;
    END IF;
    RETURN NEW;
END
$update_vote$ LANGUAGE plpgsql;

CREATE TRIGGER update_vote
    AFTER UPDATE
    ON votes
    FOR EACH ROW
    EXECUTE PROCEDURE update_vote();

CREATE OR REPLACE FUNCTION post_update() RETURNS TRIGGER AS
$post_update$
DECLARE
    prevRoute         BIGINT[];
    parent_thread INT;
BEGIN
    IF (NEW.parent = 0) THEN
        NEW.route := array_append(new.route, new.id);
    ELSE
        SELECT route FROM post WHERE id = new.parent INTO prevRoute;
        --- проверка на parent
        SELECT thread FROM post WHERE id = prevRoute[1] INTO parent_thread;
        IF NOT FOUND OR parent_thread <> NEW.thread THEN
            RAISE EXCEPTION 'parent is from different thread';
        END IF;
        ---
        NEW.route := NEW.route || prevRoute || new.id;
    end if;
    UPDATE forum SET posts=posts + 1 WHERE forum.slug = new.forum;
    RETURN NEW;
END
$post_update$ LANGUAGE plpgsql;

CREATE TRIGGER post_update
    BEFORE INSERT
    ON post
    FOR EACH ROW
    EXECUTE PROCEDURE post_update();