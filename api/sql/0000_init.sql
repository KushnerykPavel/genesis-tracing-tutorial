CREATE
USER postgres SUPERUSER;

create table orders
(
    id         bigserial
        primary key,
    customer_id varchar not null,
    price      double precision not null,
    created_at timestamp default now(),
    updated_at timestamp default now()
);
