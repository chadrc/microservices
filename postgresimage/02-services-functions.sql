CREATE OR REPLACE FUNCTION create_user(username VARCHAR, pass VARCHAR, token VARCHAR)
  RETURNS user_accounts
    as $$
      INSERT INTO user_accounts ("name", "password", "accesstoken", "currentsession") VALUES (username, pass, token, 1);
      INSERT INTO user_sessions ("userid", "id") VALUES ((SELECT id FROM user_accounts WHERE "name"=username), 1);
      SELECT * FROM user_accounts WHERE "name"=username;
    $$
    LANGUAGE SQL;