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
	// Name of user
	UserName string
	// App Token for user
	Token string
	// Time when the session was created (for expiration)
	Created time.Time
}

// NewSession creates a new session for a user with the given key and a new
// unique ID.
func NewSession(username, token string) *Session {
	sess := new(Session)
	sess.ID = uuid.New().String()
	sess.Token = token
	sess.UserName = username
	sess.Created = time.Now()
	return sess
}

// InsertJob inserts a new Job into the database.  Upon successful return, the
// Job has a new unique ID.
func (conn *Connection) InsertSession(sess *Session) error {
	_, err := conn.engine.Insert(sess)
	return err
}

// GetSession retrieves a session from the database given its ID.
func (conn *Connection) GetSession(id string) (*Session, error) {
	sess := new(Session)
	sess.ID = id
	if has, err := conn.engine.Get(sess); err != nil {
		return nil, err
	} else if !has {
		return nil, fmt.Errorf("not found")
	}
	return sess, nil
}
