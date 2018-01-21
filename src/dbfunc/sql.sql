create database btc;
create user rich identified by rich;
grant all privileges on btc.* to 'rich'@'%';
flush privileges;

create table btc10min(
    id int auto_increment,
    timestamp int(11),
    qty float(11, 8),
    avgPrice int(11),
    firstPrice int(11),
    lastPrice int(11),
    maxPrice int(11),
    minPrice int(11),
    bolband int(11),
    bolbandsd int(3),
    Primary key(id)
);

select avg(avgPrice) as avg from btc10min where id >= (select max(id) from btc10min) - 20;
select avgPrice from btc10min where id >= (select max(id) from btc10min) - 20;