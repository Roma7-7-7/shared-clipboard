create table if not exists sessions
(
    session_id serial primary key,
    user_id    int       not null,
    name       varchar(255) not null,
    created_at timestamp not null default now(),
    updated_at timestamp not null default now()
);

create index sessions_user_id_idx on sessions (user_id);