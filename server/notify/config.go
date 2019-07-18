package notify

import (
	"doomsday/server/notify/backend"
	"doomsday/server/notify/schedule"
)

type Config struct {
	Backend     backend.Config  `yaml:"backend"`
	Schedule    schedule.Config `yaml:"schedule"`
	DoomsdayURL string          `yaml:"doomsday_url"`
}
