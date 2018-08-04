package database

import (
	"errors"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type User struct {
	Id            bson.ObjectId   `bson:"_id,omitempty" json:"id"`
	Email         string          `bson:"email" json:"email"`
	Name          string          `bson:"name" json:"name"`
	Password      []byte          `bson:"password" json:"-"`
	Content       string          `bson:"content" json:"content"`
	Collaborators []bson.ObjectId `bson:"collaborators" json:"-"`
}

const COLL_USERS string = "users"

var ErrUserExists error = errors.New("User already exists with this email id")

func (d *db) AddUser(user *User) error {
	_, err := d.GetUserByEmail(user.Email)
	if err == mgo.ErrNotFound {
		return d.insert(COLL_USERS, user)
	} else if err != nil {
		return err
	} else {
		return ErrUserExists
	}
}

func (d *db) RemoveUser(id bson.ObjectId) error {
	return d.remove(COLL_USERS, bson.M{"_id": id})
}

func (d *db) GetUser(id bson.ObjectId) (*User, error) {
	var user User

	err := d.single(COLL_USERS, &user, bson.M{"_id": id})
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (d *db) GetCollaborators(ids []bson.ObjectId) ([]User, error) {
	var result []User = make([]User, 0)

	err := d.withCollection(COLL_USERS, func(c *mgo.Collection) error {
		return c.Find(bson.M{"_id": bson.M{"$in": ids}}).Select(bson.M{"collaborators": 0}).All(&result)
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (d *db) GetCollaborating(id bson.ObjectId) ([]User, error) {
	var result []User = make([]User, 0)

	err := d.withCollection(COLL_USERS, func(c *mgo.Collection) error {
		return c.Find(bson.M{"collaborators": id}).All(&result)
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (d *db) GetUserByEmail(email string) (*User, error) {
	var user User

	err := d.single(COLL_USERS, &user, bson.M{"email": email})
	if err != nil {
		return nil, err
	}

	return &user, nil

}

func (d *db) UpdateUser(id bson.ObjectId, updater interface{}) error {
	return d.update(COLL_USERS, bson.M{"_id": id}, updater)
}

func (d *db) SetContent(id bson.ObjectId, value string) error {
	return d.update(COLL_USERS, bson.M{"_id": id}, bson.M{"$set": bson.M{"content": value}})
}

func (d *db) GetContent(id bson.ObjectId) (string, error) {
	var content string

	err := d.withCollection(COLL_USERS, func(c *mgo.Collection) error {
		return c.Find(bson.M{"_id": id}).Select(bson.M{"content": 1}).One(&content)
	})
	if err != nil {
		return "", err
	}

	return content, nil
}

func (d *db) AddCollaborator(id bson.ObjectId, collaboratorId bson.ObjectId) error {
	if id == collaboratorId {
		return errors.New("You can't be collaborating with yourself")
	}

	updater := bson.M{
		"$addToSet": bson.M{
			"collaborators": collaboratorId,
		}}
	return d.update(COLL_USERS, bson.M{"_id": id}, updater)
}

func (d *db) RemoveCollaborator(id bson.ObjectId, collaboratorId bson.ObjectId) error {
	updater := bson.M{
		"$pull": bson.M{
			"collaborators": collaboratorId,
		}}
	return d.update(COLL_USERS, bson.M{"_id": id}, updater)
}
