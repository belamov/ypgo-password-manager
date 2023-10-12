create table if not exists secrets(
    secret_data bytea not null,
    user_id int not null,
    secret_type smallint not null,
    secret_name varchar not null,
    unique(user_id, secret_type, secret_name)
)