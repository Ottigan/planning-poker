package internal

import (
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
)

type User struct {
	ID         string
	Name       string
	Vote       int
	Connection *websocket.Conn
}

type Manager struct {
	users map[string]User
	mu    sync.Mutex
}

func CreateUserManager() *Manager {
	return &Manager{
		users: make(map[string]User),
	}
}

func (u *Manager) New() string {
	u.mu.Lock()
	defer u.mu.Unlock()

	id := strconv.FormatInt(time.Now().UnixNano(), 10)
	log.Printf("Creating new user with ID: %s", id)
	u.users[id] = User{
		ID:   id,
		Name: "",
		Vote: 0,
	}

	return id
}

func (u *Manager) Get(id string) (User, bool) {
	u.mu.Lock()
	defer u.mu.Unlock()

	user, ok := u.users[id]
	return user, ok
}

func (u *Manager) GetAll() map[string]User {
	u.mu.Lock()
	defer u.mu.Unlock()

	return u.users
}

func (u *Manager) SetConnection(id string, conn *websocket.Conn) {
	u.mu.Lock()
	defer u.mu.Unlock()

	user := u.users[id]
	user.Connection = conn
	u.users[id] = user
}

func (u *Manager) RemoveConnection(id string) {
	u.mu.Lock()
	defer u.mu.Unlock()

	user := u.users[id]
	user.Connection.Close()
	user.Connection = nil
}

func (u *Manager) SetVote(id string, vote int) {
	u.mu.Lock()
	defer u.mu.Unlock()

	user := u.users[id]
	user.Vote = vote
	u.users[id] = user
}

func (u *Manager) ResetVotes() {
	u.mu.Lock()
	defer u.mu.Unlock()

	for id, user := range u.users {
		user.Vote = 0
		u.users[id] = user
	}
}

func (u *Manager) Remove(id string) {
	u.mu.Lock()
	defer u.mu.Unlock()

	delete(u.users, id)
}

func (u *Manager) Broadcast(message []byte) {
	u.mu.Lock()
	defer u.mu.Unlock()

	for _, user := range u.users {
		if user.Connection != nil {
			if err := user.Connection.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("Failed to write message to user %s: %v", user.ID, err)
			}
		}
	}
}
