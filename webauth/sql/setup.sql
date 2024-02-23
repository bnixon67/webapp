CREATE DATABASE weblogin;
CREATE USER weblogin IDENTIFIED BY 'password';
GRANT ALL PRIVILEGES ON weblogin.* TO weblogin;

CREATE DATABASE weblogin_test;
CREATE USER weblogin_test IDENTIFIED BY 'password';
GRANT ALL PRIVILEGES ON weblogin_test.* TO weblogin_test;
