package config

import "time"

// default const
const (
	DefaultConnMaxLifeTime      time.Duration = 1 * time.Hour
	DefaultConnMaxIdleTime      time.Duration = 15 * time.Minute
	DefaultAccessTokenDuration  time.Duration = 1 * time.Hour
	DefaultRefreshTokenDuration time.Duration = 24 * time.Hour * 7 // 7 days
	DefaultRedisExpiredDuration time.Duration = 5 * time.Minute
)
