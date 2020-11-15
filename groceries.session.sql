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