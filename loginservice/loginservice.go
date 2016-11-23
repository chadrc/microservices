package main

import (
	"fmt"
	"net/http"
	"crypto/sha256"
	"time"
	"math/rand"
	"os"
	"strconv"
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

func main() {

	http.HandleFunc("/register", registerNewUser)
	http.HandleFunc("/login", loginUser)
	http.HandleFunc("/logout", logoutUser)
	http.HandleFunc("/checktoken", checkAccessToken)
	fmt.Printf("%s: Login service listening on port 8080\n", strconv.Itoa(os.Getpid()))
	http.ListenAndServe(":8080", nil)
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

	_, userExists := usersByName[name]
	if userExists {
		http.Error(response, "User with that name already exists.", http.StatusBadRequest)
		return;
	}

	var user User
	user.name = name

	passSha := sha256.New()
	passSha.Write([]byte(password))
	user.password = fmt.Sprintf("%x",string(passSha.Sum(nil)))

	tokenSha := sha256.New()
	tokenSha.Write([]byte(time.Now().String() + string(rand.Int63())))
	user.accessToken = fmt.Sprintf("%x", string(tokenSha.Sum(nil)))

	usersByName[user.name] = &user
	loggedInUsers[user.accessToken] = &user

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