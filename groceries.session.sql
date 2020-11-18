create table lists (
    id int UNIQUE AUTO_INCREMENT,
    identifier VARCHAR(8),
    title VARCHAR(255)
)

create table items (
    id int UNIQUE AUTO_INCREMENT,
    list_id int,
    text VARCHAR(255)
)

select * from lists
select * from items

SELECT title FROM lists WHERE identifier = 'mJrEnUpJ'

delete from lists where true

insert into items (list_id, text) values (10, "gooble")

drop table lists