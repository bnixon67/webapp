DROP TABLE IF EXISTS users;
source users.sql;

INSERT INTO users(username, fullName, email, hashedPassword)
VALUES ('test', 'Test User', 'test@email', '$2a$10$2bLycFqUmc6m6iLkaeUgKOGwzekGd9IoAPMbXRNNuJ8Sv9ItgV29O');

INSERT INTO users(username, fullName, email, hashedPassword, admin)
VALUES ('admin', 'Admin User', 'admin@email', '$2a$10$2bLycFqUmc6m6iLkaeUgKOGwzekGd9IoAPMbXRNNuJ8Sv9ItgV29O', 1);

INSERT INTO users(username, fullName, email, hashedPassword, admin, confirmed)
VALUES ('unconfirmed', 'Unconfirmed User', 'unconfirmed@email', '$2a$10$2bLycFqUmc6m6iLkaeUgKOGwzekGd9IoAPMbXRNNuJ8Sv9ItgV29O', 1, 0);

INSERT INTO users(username, fullName, email, hashedPassword, admin, confirmed)
VALUES ('confirmed', 'Unconfirmed User', 'confirmed@email', '$2a$10$2bLycFqUmc6m6iLkaeUgKOGwzekGd9IoAPMbXRNNuJ8Sv9ItgV29O', 1, 1);

INSERT INTO users(username, fullName, email, hashedPassword, admin, confirmed)
VALUES ('expired', 'Expired Confirm Token', 'expired@email', '$2a$10$2bLycFqUmc6m6iLkaeUgKOGwzekGd9IoAPMbXRNNuJ8Sv9ItgV29O', 1, 1);

DROP TABLE IF EXISTS tokens;
source tokens.sql;

DROP TABLE IF EXISTS events;
source events.sql;

INSERT INTO events(username, created, name, succeeded)
VALUES
("test1", "2023-01-15 01:00:00", "login", true),

("test2", "2023-01-15 01:00:00", "login", true),
("test2", "2023-01-15 02:00:00", "login", true),

("test3", "2023-01-15 03:00:00", "login", true),
("test3", "2023-01-15 02:00:00", "login", true),
("test3", "2023-01-15 01:00:00", "login", true),

("test4", "2023-01-15 01:00:00", "login", true),
("test4", "2023-01-15 04:00:00", "login", true),
("test4", "2023-01-15 02:00:00", "login", true),
("test4", "2023-01-15 03:00:00", "login", true);
