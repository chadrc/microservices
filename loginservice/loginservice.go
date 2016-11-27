package main

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"math/rand"
	"net/http"
	"os"
	"time"
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

func getInfo(response http.ResponseWriter, _ *http.Request) {
	response.Write([]byte(
		fmt.Sprintf("{dockerBridgeAddr: %s, postgresConnected: %t, postgresConnectionMessage: %s}",
			dockerBridgeAddr, postgresConnected, postgresConnectionMessage)))
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
	user.accessToken = fmt.Sprintf("%x", string(tokenSha.Sum(nil)))

	var userId int32
	var sessionId int32
	err = db.QueryRow("SELECT * FROM create_user($1, $2, $3)", name, user.password, user.accessToken).Scan(&userId, &sessionId)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
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

	passSha := sha256.New()
	passSha.Write([]byte(password))
	hashedPass := fmt.Sprintf("%x", string(passSha.Sum(nil)))

	tokenSha := sha256.New()
	tokenSha.Write([]byte(time.Now().String() + string(rand.Int63())))
	newAccessToken := fmt.Sprintf("%x", string(tokenSha.Sum(nil)))

	var userId int32;
	var sessionId int32;
	var accessToken string;
	err := db.QueryRow("SELECT * FROM get_session_with_username_and_pass($1, $2, $3)", name, hashedPass, newAccessToken).Scan(&userId, &sessionId, &accessToken)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}

	if userId == 0 {
		http.Error(response, "Invalid login credentials.", http.StatusBadRequest)
		return
	}

	response.Write([]byte("{message: 'Login successful', accessToken: '" + accessToken + "'}"))
}

func logoutUser(response http.ResponseWriter, request *http.Request) {
	token := request.URL.Query().Get("accessToken")
	if token == "" {
		http.Error(response, "Access token required.", http.StatusBadRequest)
		return
	}

	var userId int32;
	var sessionId int32;
	err := db.QueryRow("SELECT * FROM clear_access_token($1)", token).Scan(&userId, &sessionId)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}

	if userId == 0 {
		http.Error(response, "Invalid access token.", http.StatusBadRequest)
		return
	}
	response.Write([]byte("{message: 'Logout successful.'}"))
}

func checkAccessToken(response http.ResponseWriter, request *http.Request) {
	token := request.URL.Query().Get("accessToken")
	if token == "" {
		http.Error(response, "Access token required.", http.StatusBadRequest)
		return
	}

	var userId int32;
	var sessionId int32;
	err := db.QueryRow("SELECT * FROM ping_session_and_get_user_with_access_token($1)", token).Scan(&userId, &sessionId)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
	}

	if userId != 0 {
		response.Write([]byte("{valid: true}"))
	} else {
		response.Write([]byte("{valid: false}"))
	}
}
