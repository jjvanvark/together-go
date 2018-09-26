package routing

import (
	"errors"
	"log"
	"maus/together-go/database"
	"net/http"

	"gopkg.in/mgo.v2/bson"
)

func secure(fn func(rw http.ResponseWriter, req *http.Request, user *database.User)) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		cookie, err := req.Cookie(cookieName)
		if err == http.ErrNoCookie {
			http.Error(rw, "Unauthorized", http.StatusUnauthorized)
			return
		} else if err != nil {
			log.Println(err)
			http.Error(rw, "Internal server error", http.StatusInternalServerError)
			return
		}

		var hex string
		err = sc.Decode(cookieName, cookie.Value, &hex)
		if err != nil {
			log.Println(err)
			http.Error(rw, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if !bson.IsObjectIdHex(hex) {
			log.Println(errors.New("Illegal objectId Hex in cookie: " + hex))
			http.Error(rw, "Unauthorized", http.StatusUnauthorized)
			return
		}

		user, err := db.GetUser(bson.ObjectIdHex(hex))
		if err != nil {
			log.Println(err)
			http.Error(rw, "Unauthorized", http.StatusUnauthorized)
			return
		}

		fn(rw, req, user)
	}
}

func handleCheck(rw http.ResponseWriter, req *http.Request) {
	cookie, err := req.Cookie(cookieName)
	if err == http.ErrNoCookie {
		http.Error(rw, "Unauthorized", http.StatusUnauthorized)
		return
	} else if err != nil {
		log.Println(err)
		http.Error(rw, "Internal server error", http.StatusInternalServerError)
		return
	}

	var hex string
	err = sc.Decode(cookieName, cookie.Value, &hex)
	if err != nil {
		log.Println(err)
		http.Error(rw, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !bson.IsObjectIdHex(hex) {
		log.Println(errors.New("Illegal objectId Hex in cookie: " + hex))
		http.Error(rw, "Unauthorized", http.StatusUnauthorized)
		return
	}

	rw.Header().Set("Content-Type", "text/plain")
	rw.Write([]byte(cookie.Value))

}
