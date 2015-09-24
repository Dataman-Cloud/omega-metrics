use oapp;
create table if not exists application (
    id bigint unsigned primary key auto_increment,
    cid bigint unsigned not null,
    name varchar(64) not null unique,
    status tinyint not null,
    json text,
    created timestamp,
    updated timestamp DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
    );
