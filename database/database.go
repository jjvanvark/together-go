package database

import (
	"log"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func InitDatabase(host, name string) (Db, error) {
	sess, err := mgo.Dial(host)
	if err != nil {
		return nil, err
	}

	sess.SetMode(mgo.Monotonic, true)

	return &db{name, sess}, nil
}

type Db interface {
	Close()
	Drop(dbName string) error
	AddUser(user *User) error
	RemoveUser(id bson.ObjectId) error
	GetUser(id bson.ObjectId) (*User, error)
	GetCollaborators(ids []bson.ObjectId) ([]User, error)
	GetCollaborating(id bson.ObjectId) ([]User, error)
	GetUserByEmail(email string) (*User, error)
	UpdateUser(id bson.ObjectId, updater interface{}) error
	SetContent(id bson.ObjectId, value string) error
	GetContent(id bson.ObjectId) (string, error)
	AddCollaborator(id bson.ObjectId, collaboratorId bson.ObjectId) error
	RemoveCollaborator(id bson.ObjectId, collaboratorId bson.ObjectId) error
}

type db struct {
	name string
	sess *mgo.Session
}

func (d *db) Close() {
	log.Println("Closing db")
	d.sess.Close()
}

func (d *db) Drop(dbName string) error {
	sess := d.sess.Clone()
	defer sess.Close()

	return sess.DB(dbName).DropDatabase()
}

func (d *db) withCollection(name string, fn func(c *mgo.Collection) error) error {
	s := d.sess.Clone()
	defer s.Close()

	return fn(s.DB(d.name).C(name))
}

func (d *db) single(name string, value interface{}, query interface{}) error {
	return d.withCollection(name, func(c *mgo.Collection) error {
		return c.Find(query).One(value)
	})
}

func (d *db) multi(name string, value interface{}, query interface{}) error {
	return d.withCollection(name, func(c *mgo.Collection) error {
		return c.Find(query).All(value)
	})
}

func (d *db) insert(name string, value interface{}) error {
	return d.withCollection(name, func(c *mgo.Collection) error {
		return c.Insert(value)
	})
}

func (d *db) update(name string, query interface{}, updater interface{}) error {
	return d.withCollection(name, func(c *mgo.Collection) error {
		return c.Update(query, updater)
	})
}

func (d *db) remove(name string, query interface{}) error {
	return d.withCollection(name, func(c *mgo.Collection) error {
		return c.Remove(query)
	})
}
