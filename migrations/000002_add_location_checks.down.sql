drop table if exists location_checks;

alter table incidents
    drop column if exists deactivated_at;
