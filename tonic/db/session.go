package db

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Session holds all the information for a given user Session
type Session struct {
	// Session ID (stored in the cookie)
	ID string `xorm:"pk"`
	// App Token for user
	Token string
	// Time when the session was created (for expiration)
	Created time.Time
}

// NewSession creates a new session for a user with the given token and a new
// unique ID.
func NewSession(token string) *Session {
	sess := new(Session)
	sess.ID = uuid.New().String()
	sess.Token = token
	sess.Created = time.Now()
	return sess
}

// InsertSession inserts a new Session into the database.  Upon successful
// return, the Job has a new unique ID.
func (conn *Connection) InsertSession(sess *Session) error {
	_, err := conn.engine.Insert(sess)
	return err
}

// GetSession retrieves a session from the database given its ID.
func (conn *Connection) GetSession(id string) (*Session, error) {
	sess := new(Session)
	if has, err := conn.engine.ID(id).Get(sess); err != nil {
		return nil, err
	} else if !has {
		return nil, fmt.Errorf("not found")
	}
	return sess, nil
}

// DeleteSession removes the Session matching the given ID from the database.
func (conn *Connection) DeleteSession(id string) error {
	_, err := conn.engine.ID(id).Delete(new(Session))
	return err
}
