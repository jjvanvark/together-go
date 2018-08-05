package routing

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/securecookie"

	"golang.org/x/crypto/bcrypt"
	mgo "gopkg.in/mgo.v2"
)

var hashKey = []byte("xxxxxxxxxxxxxxxx")
var blockKey = []byte("xxxxxxxxxxxxxxxx")
var sc = securecookie.New(hashKey, blockKey)

const cookieName string = "cookiename"

func handleLogin(rw http.ResponseWriter, req *http.Request) {
	result := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Println(err)
		http.Error(rw, "Internal server error", http.StatusInternalServerError)
		return
	}

	defer req.Body.Close()

	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Println(err)
		http.Error(rw, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !emailRegEx.MatchString(result.Email) {
		log.Println("Login :: Not a valid email address")
		http.Error(rw, "Not a valid email or password", http.StatusUnprocessableEntity)
		return
	}

	user, err := db.GetUserByEmail(result.Email)
	if err == mgo.ErrNotFound {
		log.Println("Login :: email not found")
		http.Error(rw, "Not a valid email or password", http.StatusUnprocessableEntity)
		return
	} else if err != nil {
		log.Println(err)
		http.Error(rw, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = bcrypt.CompareHashAndPassword(user.Password, []byte(result.Password))
	if err != nil {
		log.Println(err)
		http.Error(rw, "Not a valid email or password", http.StatusUnprocessableEntity)
		return
	}

	encoded, err := sc.Encode(cookieName, user.Id.Hex())
	if err != nil {
		log.Println(err)
		http.Error(rw, "Internal server error", http.StatusInternalServerError)
		return
	}

	expire := time.Now().Add(24 * time.Hour)
	http.SetCookie(rw, &http.Cookie{
		Name:    cookieName,
		Value:   encoded,
		Path:    prefix,
		MaxAge:  0,
		Expires: expire,
	})
}
