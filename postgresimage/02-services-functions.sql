CREATE OR REPLACE FUNCTION username_exists(username VARCHAR)
  RETURNS VARCHAR(255)
    AS $$
    DECLARE
      user_name VARCHAR = '';
    BEGIN
      IF EXISTS (SELECT name FROM user_accounts WHERE name=username) THEN
        SELECT name FROM user_accounts WHERE name=username INTO user_name;
      END IF;
      RETURN user_name;
    END;
    $$
    LANGUAGE "plpgsql" IMMUTABLE;

CREATE OR REPLACE FUNCTION create_user(username VARCHAR, pass VARCHAR, token VARCHAR)
  RETURNS TABLE (
    userId INTEGER,
    sessionId INTEGER
  )
    AS $$
      INSERT INTO user_accounts (name, password, access_token, current_session) VALUES (username, pass, token, 1);
      INSERT INTO user_sessions (user_id, id) VALUES ((SELECT id FROM user_accounts WHERE name=username), 1);
      SELECT id, current_session FROM user_accounts WHERE name=username;
    $$
    LANGUAGE SQL;

CREATE OR REPLACE FUNCTION ping_session_and_get_user_with_access_token(token VARCHAR)
  RETURNS TABLE (
    userId INTEGER,
    sessionId INTEGER
  )
    AS $$
      DECLARE
        found_userId INTEGER := 0;
        found_sessionId INTEGER := 0;
        session_timeout INTERVAL := 10 * INTERVAL '1 MINUTE';
        found_user user_accounts;
        user_session user_sessions;
      BEGIN
        IF EXISTS (SELECT id FROM user_accounts WHERE access_token=token) THEN
          SELECT * FROM user_accounts WHERE access_token=token INTO found_user;
          SELECT * FROM user_sessions WHERE id=found_user.current_session
                                            AND user_id=found_user.id INTO user_session;

          found_userId = found_user.id;

          IF (now() - user_session.last_check) > session_timeout THEN
            found_sessionId = user_session.id+1;

            INSERT INTO user_sessions (user_id, id) VALUES (found_user.id, found_sessionId);

            UPDATE user_accounts SET current_session=found_sessionId
            WHERE id=found_user.id;
          ELSE
            UPDATE user_sessions SET last_check=now()
            WHERE id=found_user.current_session AND user_id=found_user.id;

            found_sessionId = user_session.id;
          END IF;
        END IF;
        RETURN QUERY SELECT found_userId, found_sessionId;
      END;
    $$
    LANGUAGE "plpgsql" VOLATILE;

CREATE OR REPLACE FUNCTION update_access_token_and_ping_session(token VARCHAR, newToken VARCHAR)
  RETURNS TABLE (
    userId INTEGER,
    sessionId INTEGER
  )
    AS $$
      DECLARE

      BEGIN
        UPDATE user_accounts SET access_token=newToken WHERE access_token=token;
        RETURN QUERY SELECT * FROM ping_session_and_get_user_with_access_token(newToken);
      END;
    $$
    LANGUAGE "plpgsql" VOLATILE;

CREATE OR REPLACE FUNCTION clear_access_token(token VARCHAR)
  RETURNS TABLE (
    userId INTEGER,
    sessionId INTEGER
  )
    AS $$
      DECLARE
        found_user_id INTEGER = 0;
        found_session_id INTEGER = 0;
      BEGIN
        IF EXISTS(SELECT access_token FROM user_accounts WHERE access_token=token) THEN
          UPDATE user_accounts SET access_token='' WHERE access_token=token
          RETURNING id, current_session INTO found_user_id, found_session_id;
        END IF;

        RETURN QUERY SELECT found_user_id, found_session_id;
      END;
    $$
    LANGUAGE "plpgsql" VOLATILE;

CREATE OR REPLACE FUNCTION get_session_with_username_and_pass(username VARCHAR, pass VARCHAR, token VARCHAR)
  RETURNS TABLE (
    userId INTEGER,
    sessionId INTEGER,
    accessToken VARCHAR(64)
  )
    AS $$
      DECLARE
        found_userId INTEGER := 0;
        found_sessionId INTEGER := 0;
        found_token VARCHAR(64) := '';
      BEGIN
        IF EXISTS (SELECT id FROM user_accounts WHERE name=username AND password=pass) THEN

          UPDATE user_accounts SET access_token=token
          WHERE name=username
                AND password=pass
                AND access_token=''
          RETURNING access_token INTO found_token;

          if found_token IS NULL THEN
            SELECT access_token FROM user_accounts WHERE name=username AND password=pass INTO found_token;
          END IF;

          SELECT * FROM ping_session_and_get_user_with_access_token(found_token) INTO found_userId, found_sessionId;
          RETURN QUERY SELECT found_userId, found_sessionId, found_token;
        ELSE
          RETURN QUERY SELECT found_userId, found_sessionId, found_token;
        END IF;
      END;
    $$
    LANGUAGE "plpgsql" VOLATILE;