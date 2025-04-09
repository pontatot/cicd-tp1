create or replace table city (
    id int primary key,
    department_code varchar(255) not null,
    insee_code varchar(255),
    zip_code varchar(255),
    name varchar(255) not null,
    lat float not null,
    lon float not null
)
