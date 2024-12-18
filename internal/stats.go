package internal

import (
	"log"
	"math"
	"strconv"
)

type Stats struct {
	Min string
	Avg string
	Max string
}

func GetVotedUsers(um *Manager) int {
	votedUsers := 0
	users := um.GetAll()

	for _, user := range users {
		if user.Vote != 0 {
			votedUsers++
		}
	}
	return votedUsers
}

func CalculateMinAvgMax(um *Manager, showResult bool) Stats {
	votes := make([]float64, 0)
	users := um.GetAll()

	if len(users) == 0 || !showResult || GetVotedUsers(um) == 0 {
		return Stats{}
	}

	for _, user := range users {
		if user.Vote != 0 {
			log.Println("Vote found", user.Vote)
			votes = append(votes, float64(user.Vote))
		}
	}

	min := math.Inf(1)
	max := math.Inf(-1)
	sum := 0.0

	for _, vote := range votes {
		if vote < min {
			min = vote
		}
		if vote > max {
			max = vote
		}
		sum += vote
	}

	avg := sum / float64(len(votes))

	return Stats{
		Min: strconv.Itoa(int(min)),
		Avg: strconv.Itoa(int(avg)),
		Max: strconv.Itoa(int(max)),
	}
}
