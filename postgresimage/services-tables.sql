CREATE TABLE users (
  name VARCHAR(128) PRIMARY KEY,
  password VARCHAR(64) NOT NULL,
  sessionToken VARCHAR(64),
  created TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE session_tokens (
  username VARCHAR(128) NOT NULL,
  token VARCHAR(64) NOT NULL,
  lastcheck TIMESTAMP NOT NULL DEFAULT now(),
  created TIMESTAMP NOT NULL DEFAULT now(),
  PRIMARY KEY (username, token),
  FOREIGN KEY (username) REFERENCES users (name)
);