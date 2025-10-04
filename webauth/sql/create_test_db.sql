create database weblogin_test;
create user 'weblogin_test' identified by 'password';
grant all privileges on weblogin_test.* to weblogin_test;
flush privileges;
