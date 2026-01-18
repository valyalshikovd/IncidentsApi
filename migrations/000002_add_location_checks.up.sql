alter table incidents
    add column if not exists deactivated_at timestamptz;

create table if not exists location_checks
(
    id bigserial primary key,
    user_id varchar(128) not null,
    latitude double precision not null,
    longitude double precision not null,
    has_danger boolean not null,
    created_at timestamptz not null default now(),
    constraint location_checks_latitude_range check (latitude >= -90.0 and latitude <= 90.0),
    constraint location_checks_longitude_range check (longitude >= -180.0 and longitude <= 180.0)
);

create index if not exists idx_location_checks_created_at
    on location_checks (created_at);

create index if not exists idx_location_checks_user_id_created_at
    on location_checks (user_id, created_at);
