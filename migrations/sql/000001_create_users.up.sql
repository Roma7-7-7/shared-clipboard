create table if not exists users
(
    user_id       serial primary key,
    name          varchar(256) not null unique,
    password      varchar(256) not null,
    password_salt varchar(256) not null,
    created_at    timestamp    not null default now(),
    updated_at    timestamp    not null default now()
);
