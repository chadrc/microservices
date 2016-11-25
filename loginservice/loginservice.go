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

var usersByName map[string]*User = make(map[string]*User)
var loggedInUsers map[string]*User = make(map[string]*User)

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

func registerNewUser(response http.ResponseWriter, request *http.Request) {
	name := request.URL.Query().Get("name")
	if name == "" {
		http.Error(response, "Name required.", http.StatusBadRequest)
		return;
	}

	password := request.URL.Query().Get("password")
	if password == "" {
		http.Error(response, "Password required.", http.StatusBadRequest)
		return;
	}

	var dataName string
	userRow := db.QueryRow("SELECT name FROM users WHERE name = $1", name)
	err := userRow.Scan(&dataName)
	if err != nil && err != sql.ErrNoRows{
		http.Error(response, "DB Error: " + err.Error(), http.StatusInternalServerError)
		return
	}

	if err != sql.ErrNoRows {
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

	//usersByName[user.name] = &user
	//loggedInUsers[user.accessToken] = &user

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
		return;
	}

	password := request.URL.Query().Get("password")
	if password == "" {
		http.Error(response, "Password required.", http.StatusBadRequest)
		return;
	}

	user, userExists := usersByName[name]
	if !userExists {
		http.Error(response, "Invalid login credentials.", http.StatusBadRequest)
		return;
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

	loggedInUsers[user.accessToken] = user

	response.Write([]byte("{message: 'Login successful', accessToken: '" + user.accessToken +"'}"))
}

func logoutUser(response http.ResponseWriter, request *http.Request) {
	token := request.URL.Query().Get("accessToken")
	if token == "" {
		http.Error(response, "Access token required.", http.StatusBadRequest)
		return
	}

	user, exists := loggedInUsers[token]
	if !exists {
		http.Error(response, "Invalid access token.", http.StatusBadRequest)
		return
	}

	user.accessToken = ""
	delete(loggedInUsers, token)
	response.Write([]byte("{message: 'Logout successful.'}"))
}

func checkAccessToken(response http.ResponseWriter, request *http.Request) {
	token := request.URL.Query().Get("accessToken")
	if token == "" {
		http.Error(response, "Access token required.", http.StatusBadRequest)
		return
	}

	_, exists := loggedInUsers[token]
	if exists {
		response.Write([]byte("{valid: true}"))
	} else {
		response.Write([]byte("{valid: false}"))
	}
}