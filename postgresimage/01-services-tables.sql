CREATE TABLE user_accounts (
  id SERIAL NOT NULL UNIQUE,
  name VARCHAR(128) PRIMARY KEY,
  password VARCHAR(64) NOT NULL,
  accessToken VARCHAR(64),
  currentSession INTEGER NOT NULL ,
  created TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE user_sessions (
  id INTEGER NOT NULL,
  userId INTEGER NOT NULL,
  lastcheck TIMESTAMP NOT NULL DEFAULT now(),
  created TIMESTAMP NOT NULL DEFAULT now(),
  PRIMARY KEY (id, userId),
  FOREIGN KEY (userId) REFERENCES user_accounts (id)
);