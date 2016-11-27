package main

import (
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

type User struct {
	name         string
	password     string
	sessionToken string
}

func (user *User) GetName() (username string) {
	username = user.name
	return
}

var db *sql.DB
var dockerBridgeAddr string
var postgresPort = "5300"
var postgresConnected bool
var postgresConnectionMessage string

func main() {
	dockerBridgeAddr = os.Getenv("DOCKER_BRIDGE_IP")

	conn, dbErr := sql.Open("postgres",
		fmt.Sprintf("sslmode=disable user=postgres password=postgres dbname=micro_services port=%s host=%s",
			postgresPort, dockerBridgeAddr))
	if dbErr != nil {
		postgresConnected = false
		postgresConnectionMessage = dbErr.Error()
	} else {
		postgresConnected = true
		postgresConnectionMessage = "Connected"
		db = conn
	}

	http.HandleFunc("/", getInfo)
	http.HandleFunc("/register", registerNewUser)
	http.HandleFunc("/login", loginUser)
	http.HandleFunc("/logout", logoutUser)
	http.HandleFunc("/checktoken", checkSessionToken)
	http.ListenAndServe(":8080", nil)
}

func getInfo(response http.ResponseWriter, _ *http.Request) {
	response.Write([]byte(
		fmt.Sprintf("{dockerBridgeAddr: %s, postgresConnected: %t, postgresConnectionMessage: %s}",
			dockerBridgeAddr, postgresConnected, postgresConnectionMessage)))
}

func resolveUserRow(rows *sql.Row) (user *User, err error) {
	user = new(User)
	err = rows.Scan(&user.name, &user.password, &user.sessionToken)
	if err != nil {
		user = nil
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.New("DB Error: " + err.Error())
		}
		return
	}
	user.sessionToken = strings.Trim(user.sessionToken, " ")
	return
}

func getUserByName(name string) (user *User, err error) {
	userRow := db.QueryRow("SELECT name, password, sessionToken FROM users WHERE name = $1", name)
	user, err = resolveUserRow(userRow)
	return
}

func getUserBySessionToken(token string) (user *User, err error) {
	userRow := db.QueryRow("SELECT name, password, sessionToken FROM users WHERE sessionToken = $1", token)
	user, err = resolveUserRow(userRow)
	return
}

func registerNewUser(response http.ResponseWriter, request *http.Request) {
	name := request.URL.Query().Get("name")
	if name == "" {
		http.Error(response, "Name required.", http.StatusBadRequest)
		return
	}

	password := request.URL.Query().Get("password")
	if password == "" {
		http.Error(response, "Password required.", http.StatusBadRequest)
		return
	}

	var dbUser string
	err := db.QueryRow("SELECT * FROM username_exists($1)", name).Scan(&dbUser)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}

	if dbUser != "" {
		http.Error(response, "User with that name already exists.", http.StatusBadRequest)
		return
	}

	var user User
	user.name = name

	passSha := sha256.New()
	passSha.Write([]byte(password))
	user.password = fmt.Sprintf("%x", string(passSha.Sum(nil)))

	tokenSha := sha256.New()
	tokenSha.Write([]byte(time.Now().String() + string(rand.Int63())))
	user.sessionToken = fmt.Sprintf("%x", string(tokenSha.Sum(nil)))

	var userId int32
	var sessionId int32
	err = db.QueryRow("SELECT * FROM create_user($1, $2, $3)", name, user.password, user.sessionToken).Scan(&userId, &sessionId)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}

	response.Write([]byte("{message: 'Registration successful.', sessionToken: '" + user.sessionToken + "'}"))
}

func loginUser(response http.ResponseWriter, request *http.Request) {
	name := request.URL.Query().Get("name")
	if name == "" {
		http.Error(response, "Name required.", http.StatusBadRequest)
		return
	}

	password := request.URL.Query().Get("password")
	if password == "" {
		http.Error(response, "Password required.", http.StatusBadRequest)
		return
	}

	user, err := getUserByName(name)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}

	if user == nil {
		http.Error(response, "Invalid login credentials.", http.StatusBadRequest)
		return
	}

	passSha := sha256.New()
	passSha.Write([]byte(password))
	hashedPass := fmt.Sprintf("%x", string(passSha.Sum(nil)))

	if user.password != hashedPass {
		http.Error(response, "Invalid login credentials.", http.StatusBadRequest)
		return
	}

	if user.sessionToken == "" {
		tokenSha := sha256.New()
		tokenSha.Write([]byte(time.Now().String() + string(rand.Int63())))
		user.sessionToken = fmt.Sprintf("%x", string(tokenSha.Sum(nil)))
		_, err = db.Query("INSERT INTO session_tokens (username, token) VALUES($1, $2)", user.name, user.sessionToken)
		if err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	tx, err := db.Begin()
	_, err = tx.Exec(`UPDATE users SET sessionToken = $1 WHERE name = $2;`, user.sessionToken, user.name)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		tx.Rollback()
		return
	}

	_, err = tx.Exec(`UPDATE session_tokens SET lastcheck = now() WHERE token = $1 AND username = $2;`,
		user.sessionToken, user.name)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		tx.Rollback()
		return
	}

	err = tx.Commit()
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		tx.Rollback()
		return
	}

	response.Write([]byte("{message: 'Login successful', sessionToken: '" + user.sessionToken + "'}"))
}

func logoutUser(response http.ResponseWriter, request *http.Request) {
	token := request.URL.Query().Get("sessionToken")
	if token == "" {
		http.Error(response, "Access token required.", http.StatusBadRequest)
		return
	}

	user, err := getUserBySessionToken(token)
	if err != nil {
		http.Error(response, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = db.Query("UPDATE users SET sessionToken = $1 WHERE name = $2", "", user.name)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
	}
	response.Write([]byte("{message: 'Logout successful.'}"))
}

func checkSessionToken(response http.ResponseWriter, request *http.Request) {
	token := request.URL.Query().Get("sessionToken")
	if token == "" {
		http.Error(response, "Access token required.", http.StatusBadRequest)
		return
	}

	user, err := getUserBySessionToken(token)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
	}

	_, err = db.Exec(`UPDATE session_tokens SET lastcheck = now() WHERE username = $1 AND token = $2`, user.name, user.sessionToken)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}

	if user != nil {
		response.Write([]byte("{valid: true}"))
	} else {
		response.Write([]byte("{valid: false}"))
	}
}
