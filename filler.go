package main

import (
	"log"
	"maus/together-go/database"

	"golang.org/x/crypto/bcrypt"

	"gopkg.in/mgo.v2/bson"
)

func filler(db database.Db, dbName string, withDrop bool) {
	if withDrop {
		err := db.Drop(dbName)
		if err != nil {
			log.Fatal(err)
		}
	}

	var err error

	// Insert three users

	joost := createUser("joost.van.vark@gmail.com", "Joost", "joost")

	err = db.AddUser(joost)
	if err != nil {
		log.Fatal(err)
	}

	err = db.AddUser(joost)
	if err != database.ErrUserExists {
		if err != nil {
			log.Fatal(err)
		} else {
			log.Fatal("User is added but should not be allowed to")
		}
	}

	err = db.AddUser(createUser("kimkreffer@gmail.com", "Kim", "kim"))
	if err != nil {
		log.Fatal(err)
	}

	err = db.AddUser(createUser("mprudon@gmail.com", "Mathieu", "thieu"))
	if err != nil {
		log.Fatal(err)
	}

	// Update user

	var user *database.User

	user, err = db.GetUserByEmail("joost.van.vark@gmail.com")
	if err != nil {
		log.Fatal(err)
	}

	err = db.UpdateUser(user.Id, bson.M{"$set": bson.M{"email": "joostvanvark@gmail.com"}})
	if err != nil {
		log.Fatal(err)
	}

	// Add Collaborator

	var kim *database.User

	kim, err = db.GetUserByEmail("kimkreffer@gmail.com")
	if err != nil {
		log.Fatal(err)
	}

	user, err = db.GetUserByEmail("joostvanvark@gmail.com")
	if err != nil {
		log.Fatal(err)
	}

	err = db.AddCollaborator(user.Id, kim.Id)
	if err != nil {
		log.Fatal(err)
	}

	err = db.AddCollaborator(user.Id, kim.Id)
	if err != nil {
		log.Fatal(err)
	}

	err = db.RemoveCollaborator(user.Id, kim.Id)
	if err != nil {
		log.Fatal(err)
	}

	err = db.AddCollaborator(user.Id, kim.Id)
	if err != nil {
		log.Fatal(err)
	}

	// Content

	err = db.SetContent(kim.Id, "Dit is content van Kim")
	if err != nil {
		log.Fatal(err)
	}

}

func createUser(email, name, password string) *database.User {
	encryptedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}

	return &database.User{
		Id:            bson.NewObjectId(),
		Email:         email,
		Name:          name,
		Password:      encryptedPassword,
		Content:       "",
		Collaborators: []bson.ObjectId{},
	}
}
