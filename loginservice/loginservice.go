package main

import (
	"fmt"
	"net/http"
	"crypto/sha256"
	"time"
	"math/rand"
	"database/sql"
	_ "github.com/lib/pq"
	"os"
	"errors"
	"strings"
)

type User struct {
	name        string
	password    string
	accessToken string
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
	http.HandleFunc("/checktoken", checkAccessToken)
	http.ListenAndServe(":8080", nil)
}

func getInfo(response http.ResponseWriter, request *http.Request) {
	response.Write([]byte(
		fmt.Sprintf("{dockerBridgeAddr: %s, postgresConnected: %t, postgresConnectionMessage: %s}",
			dockerBridgeAddr, postgresConnected, postgresConnectionMessage)))
}

func resolveUserRow(rows *sql.Row) (user *User, err error) {
	user = new(User)
	err = rows.Scan(&user.name, &user.password, &user.accessToken)
	if err != nil {
		user = nil
		if (err == sql.ErrNoRows) {
			err = nil
		} else {
			err = errors.New("DB Error: " + err.Error())
		}
		return
	}
	user.accessToken = strings.Trim(user.accessToken, " ")
	return
}

func getUserByName(name string) (user *User, err error) {
	userRow := db.QueryRow("SELECT name, password, accessToken FROM users WHERE name = $1", name)
	user, err = resolveUserRow(userRow)
	return
}

func getUserByAccessToken(token string) (user *User, err error) {
	userRow := db.QueryRow("SELECT name, password, accessToken FROM users WHERE accessToken = $1", token)
	user, err = resolveUserRow(userRow)
	return
}

func registerNewUser(response http.ResponseWriter, request *http.Request) {
	name := request.URL.Query().Get("name")
	if name == "" {
		http.Error(response, "Name required.", http.StatusBadRequest)
		return;
	}

	password := request.URL.Query().Get("password")
	if password == "" {
		http.Error(response, "Password required.", http.StatusBadRequest)
		return
	}

	dbUser, err := getUserByName(name)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}

	if dbUser != nil {
		http.Error(response, "User with that name already exists.", http.StatusBadRequest)
		return
	}

	var user User
	user.name = name

	passSha := sha256.New()
	passSha.Write([]byte(password))
	user.password = fmt.Sprintf("%x",string(passSha.Sum(nil)))

	tokenSha := sha256.New()
	tokenSha.Write([]byte(time.Now().String() + string(rand.Int63())))
	user.accessToken = fmt.Sprintf("%x", string(tokenSha.Sum(nil)))

	_, err = db.Query(`INSERT INTO users (name, password, accessToken)
				VALUES($1, $2, $3)`, user.name, user.password, user.accessToken)
	if err != nil {
		http.Error(response, "DB error: " + err.Error(), http.StatusInternalServerError)
		return
	}

	response.Write([]byte("{message: 'Registration successful.', accessToken: '" + user.accessToken + "'}"))
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
	hashedPass := fmt.Sprintf("%x",string(passSha.Sum(nil)))

	if user.password != hashedPass {
		http.Error(response, "Invalid login credentials.", http.StatusBadRequest)
		return
	}

	if user.accessToken == "" {
		tokenSha := sha256.New()
		tokenSha.Write([]byte(time.Now().String() + string(rand.Int63())))
		user.accessToken = fmt.Sprintf("%x", string(tokenSha.Sum(nil)))
	}

	_, err = db.Query("UPDATE users SET accessToken = $1 WHERE name = $2", user.accessToken, user.name)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}

	response.Write([]byte("{message: 'Login successful', accessToken: '" + user.accessToken +"'}"))
}

func logoutUser(response http.ResponseWriter, request *http.Request) {
	token := request.URL.Query().Get("accessToken")
	if token == "" {
		http.Error(response, "Access token required.", http.StatusBadRequest)
		return
	}

	user, err := getUserByAccessToken(token)
	if err != nil {
		http.Error(response, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = db.Query("UPDATE users SET accessToken = $1 WHERE name = $2", "", user.name)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
	}
	response.Write([]byte("{message: 'Logout successful.'}"))
}

func checkAccessToken(response http.ResponseWriter, request *http.Request) {
	token := request.URL.Query().Get("accessToken")
	if token == "" {
		http.Error(response, "Access token required.", http.StatusBadRequest)
		return
	}

	user, err := getUserByAccessToken(token)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
	}

	if user != nil {
		response.Write([]byte("{valid: true}"))
	} else {
		response.Write([]byte("{valid: false}"))
	}
}