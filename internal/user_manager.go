package internal

import (
	"log"
	"sync"

	"github.com/gofiber/contrib/websocket"
)

type User struct {
	Id         string
	Name       string
	Vote       int
	Connection *websocket.Conn
}

type Option func(*User)

func WithName(name string) Option {
	return func(u *User) {
		u.Name = name
	}
}

func WithVote(vote int) Option {
	return func(u *User) {
		u.Vote = vote
	}
}

func WithConnection(conn *websocket.Conn) Option {
	return func(u *User) {
		u.Connection = conn
	}
}

type Manager struct {
	users sync.Map
}

func CreateUserManager() *Manager {
	return &Manager{}
}

func (m *Manager) New(user User) User {
	log.Printf("Creating new user with ID: %s", user.Id)
	m.users.Store(user.Id, user)

	return user
}

func (m *Manager) Get(id string) (User, bool) {
	log.Printf("Retrieving user with ID: %s", id)
	value, ok := m.users.Load(id)

	if !ok {
		m.notFound(id)
		return User{}, false
	}

	user := value.(User)
	return user, true
}

func (m *Manager) GetAll() map[string]User {
	log.Printf("Retrieving all users")
	users := make(map[string]User)

	m.users.Range(func(key, value interface{}) bool {
		users[key.(string)] = value.(User)
		return true
	})

	return users
}

func (m *Manager) Update(id string, options ...Option) (User, bool) {
	user, ok := m.Get(id)

	if !ok {
		m.notFound(id)
		return User{}, false
	}

	for _, option := range options {
		option(&user)
	}

	m.users.Store(id, user)

	return user, true
}

func (m *Manager) ResetVotes() {
	log.Printf("Resetting votes")
	m.users.Range(func(key, value interface{}) bool {
		user := value.(User)
		user.Vote = 0
		m.users.Store(key.(string), user)
		return true
	})
}

func (m *Manager) Broadcast(message []byte) {

	m.users.Range(func(key, value interface{}) bool {
		user := value.(User)

		log.Printf("Broadcasting message to user %v", user)

		if user.Connection != nil {
			if err := user.Connection.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("Failed to write message to user %s: %v", user.Id, err)
			}
		}

		return true
	})
}

func (m *Manager) notFound(id string) {
	log.Printf("User not found for ID: %s", id)
}
