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

func (m *Manager) SetConnection(id string, conn *websocket.Conn) (User, bool) {
	log.Printf("Setting connection for user with ID: %s", id)
	value, ok := m.users.Load(id)

	if !ok {
		m.notFound(id)
		return User{}, false
	}

	user := value.(User)
	user.Connection = conn
	m.users.Store(id, user)

	return user, true
}

func (m *Manager) RemoveConnection(id string) (User, bool) {
	log.Printf("Removing connection for user with ID %s", id)
	value, ok := m.users.Load(id)

	if !ok {
		m.notFound(id)
		return User{}, false
	}

	user := value.(User)
	user.Connection = nil
	m.users.Store(id, user)

	return user, true
}

func (m *Manager) SetVote(id string, vote int) (User, bool) {
	log.Printf("Setting vote %d for user with ID: %s", vote, id)
	value, ok := m.users.Load(id)

	if !ok {
		m.notFound(id)
		return User{}, false
	}

	user := value.(User)
	user.Vote = vote
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
	log.Printf("Broadcasting message to all users")

	m.users.Range(func(key, value interface{}) bool {
		user := value.(User)

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
