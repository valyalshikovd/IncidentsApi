create table if not exists incidents
(
    id bigserial primary key,
    title varchar(200) not null,
    description text,
    latitude double precision not null,
    longitude double precision not null,
    danger_radius_m integer not null default 100 check (danger_radius_m > 0),
    is_active boolean not null default true,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    constraint incidents_latitude_range check (latitude >= -90.0 and latitude <= 90.0),
    constraint incidents_longitude_range check (longitude >= -180.0 and longitude <= 180.0)
);

create index if not exists idx_incidents_is_active
    on incidents (is_active);

create index if not exists idx_incidents_active_lat_lon
    on incidents (is_active, latitude, longitude);

comment on table incidents is 'инциденты (опасные зоны). зона считается опасной в радиусе danger_radius_m от точки (latitude/longitude)';

comment on column incidents.id is 'первичный ключ инцидента';
comment on column incidents.title is 'краткое название/заголовок инцидента';
comment on column incidents.description is 'описание инцидента';
comment on column incidents.latitude is 'широта точки инцидента';
comment on column incidents.longitude is 'долгота точки инцидента';
comment on column incidents.danger_radius_m is 'радиус опасной зоны вокруг точки, метры';
comment on column incidents.is_active is 'признак активности';
comment on column incidents.created_at is 'время создания записи';
comment on column incidents.updated_at is 'время последнего обновления записи';
