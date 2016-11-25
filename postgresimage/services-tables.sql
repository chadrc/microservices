CREATE TABLE Users (
  name VARCHAR(127) CONSTRAINT first_key PRIMARY KEY,
  password CHARACTER(63) NOT NULL,
  accessToken UUID
);