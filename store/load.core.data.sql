
attach database './data/core/core.db' as 'core';

drop table if exists main.core_obras_plus;
create table main.core_obras_plus as select * from core.core_obras_plus;

drop table if exists main.core_obras_plus_phone;
create table main.core_obras_plus_phone as select * from core.core_obras_plus_phone;

detach database 'core';

-- Build Indexes
create unique index idx_core_obras_plus_id on core_obras_plus (id);
create index idx_core_obras_plus_city on core_obras_plus (city);
create index idx_core_obras_plus_phone_nome on core_obras_plus_phone (nome);
create index idx_core_obras_plus_phone_uf_cidade on core_obras_plus_phone (uf, cidade);

vacuum;