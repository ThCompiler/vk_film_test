CREATE EXTENSION IF NOT EXISTS citext;

CREATE TYPE roles as ENUM ('user', 'admin');

CREATE TABLE IF NOT EXISTS users
(
    id       bigserial   not null primary key,
    login    text unique not null,
    password text        not null,
    role     roles       not null default 'user'
);

CREATE TYPE sexes as ENUM ('male', 'female');

CREATE TABLE IF NOT EXISTS actors
(
    id       bigserial not null primary key,
    name     citext    not null,
    sex      sexes     not null,
    birthday date      not null
);

CREATE TABLE IF NOT EXISTS films
(
    id           bigserial not null primary key,
    name         citext    not null check (char_length(name) >= 1 and char_length(name) <= 150),
    description  text      not null check (char_length(description) <= 1000),
    publish_date date      not null,
    rating       int8      not null check (rating >= 0 and rating <= 10)
);

CREATE TABLE IF NOT EXISTS film_actor
(
    id       bigserial not null primary key,
    film_id  bigint    not null references films (id) on delete cascade,
    actor_id bigint    not null references actors (id) on delete cascade
);


INSERT INTO users (login, password, role)
VALUES ('admin', '$2a$10$tmOxaYBG7wIw5eWrghzwueVVtJTaKQFHGxO6Ndri7JERsaiG/V.f.', 'admin'),
       ('user', '$2a$10$h32WBtNWkQU3Nhm9jC.tv.xEVYVuqGisHIqWflLQkIr.j43fbH2XW', 'user')
