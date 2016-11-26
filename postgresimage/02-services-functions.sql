CREATE OR REPLACE FUNCTION create_user(username VARCHAR, pass VARCHAR, token VARCHAR)
  RETURNS user_accounts
    as $$
      INSERT INTO user_accounts ("name", "password", "accesstoken", "currentsession") VALUES (username, pass, token, 1);
      INSERT INTO user_sessions ("userid", "id") VALUES ((SELECT id FROM user_accounts WHERE "name"=username), 1);
      SELECT * FROM user_accounts WHERE "name"=username;
    $$
    LANGUAGE SQL;

CREATE OR REPLACE FUNCTION ping_session_and_get_user_with_access_token(token VARCHAR)
  RETURNS TABLE (
    user_id INTEGER,
    user_name VARCHAR(255),
    session_id INTEGER
  )
    as $$
      DECLARE
        found_userId INTEGER := 0;
        found_sessionId INTEGER := 0;
        found_userName VARCHAR(255) := '';
        session_timeout INTERVAL := 10 * INTERVAL '1 MINUTE';
        found_user user_accounts;
        user_session user_sessions;
      BEGIN
        IF EXISTS (SELECT id FROM user_accounts WHERE accesstoken=token) THEN
          SELECT * FROM user_accounts WHERE user_accounts.accesstoken=token INTO found_user;
          SELECT * FROM user_sessions WHERE user_sessions.id=found_user.currentsession
                                            AND user_sessions.userid=found_user.id INTO user_session;

          found_userId = found_user.id;
          found_userName = found_user.name;

          IF (now() - user_session.lastcheck) > session_timeout THEN
            found_sessionId = user_session.id+1;

            INSERT INTO user_sessions ("userid", "id") VALUES (found_user.id, found_sessionId);

            UPDATE user_accounts SET currentSession=found_sessionId
            WHERE id=found_user.id;
          ELSE
            UPDATE user_sessions SET lastcheck=now()
            WHERE id=found_user.currentsession AND userid=found_user.id;

            found_sessionId = user_session.id;
          END IF;
        END IF;
        RETURN QUERY SELECT found_userId, found_userName, found_sessionId;
      END;
    $$
    LANGUAGE "plpgsql" VOLATILE;