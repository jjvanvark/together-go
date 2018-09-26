package routing

import (
	"errors"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"gopkg.in/mgo.v2/bson"
)

var (
	key []byte = []byte("MySuperSigningKey")
)

type CustomClaims struct {
	Id bson.ObjectId `json:"id"`
	jwt.StandardClaims
}

func getToken(id bson.ObjectId) (string, error) {

	claims := CustomClaims{
		id,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(365 * 24 * time.Hour).Unix(),
			Issuer:    "maus",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(key)
	if err != nil {
		return "", err
	}

	return tokenString, nil

}

func parseToken(tokenString string) (*bson.ObjectId, error) {

	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("Signing method error")
		}
		return key, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return &claims.Id, nil
	}

	return nil, errors.New("Parsetoken failed")

}
