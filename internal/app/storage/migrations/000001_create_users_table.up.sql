create table if not exists users(
    id serial unique,
    username varchar unique not null,
    hashed_password varchar not null
)