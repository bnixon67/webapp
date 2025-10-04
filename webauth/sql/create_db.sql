create database webauth;
create user 'webauth' identified by 'password';
grant all privileges on webauth.* to webauth;
flush privileges;

use webauth;
source users.sql;
source tokens.sql;
source events.sql;
