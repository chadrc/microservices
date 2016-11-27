CREATE TABLE user_accounts (
  id SERIAL NOT NULL UNIQUE,
  name VARCHAR(255) PRIMARY KEY,
  password VARCHAR(64) NOT NULL,
  access_token VARCHAR(64) NOT NULL UNIQUE DEFAULT '',
  current_session INTEGER NOT NULL,
  created TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE user_sessions (
  id INTEGER NOT NULL,
  user_id INTEGER NOT NULL,
  last_check TIMESTAMP NOT NULL DEFAULT now(),
  created TIMESTAMP NOT NULL DEFAULT now(),
  PRIMARY KEY (id, user_id),
  FOREIGN KEY (user_id) REFERENCES user_accounts (id)
);