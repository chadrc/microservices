CREATE TABLE Users (
  name VARCHAR(127) PRIMARY KEY,
  password CHARACTER(63) NOT NULL,
  accessToken CHARACTER(63)
);