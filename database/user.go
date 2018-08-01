package database

import "gopkg.in/mgo.v2/bson"

type User struct {
	Id       bson.ObjectId `bson:"_id,omitempty" json:"id"`
	Email    string        `bson:"email" json:"email"`
	Password []byte        `bson:"password" json"-"`
}

const COLL_USERS string = "users"

func (d *db) AddUser(user *User) error {
	return d.insert(COLL_USERS, user)
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

func (d *db) UpdateUser(id bson.ObjectId, updater interface{}) error {
	return d.update(COLL_USERS, bson.M{"_id": id}, updater)
}
