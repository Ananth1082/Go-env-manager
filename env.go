package main

import "time"

// define the env structs here

// example
type Cloudinary struct {
	CloudName    string `env:"CLOUD_NAME"`
	ClientSecret string `env:"CLIENT_SECRET"`
	CleintId     string `env:"CLIENT_ID"`
}

type Twilio struct {
	ApiKey     string
	ApiSeceret string
	Channel    int
}

type JWT struct {
	RefreshTokenSecret     string
	RefreshTokenExpiryTime time.Duration
	AcessTokenSecret       string
	AcessTokenExpiryTime   time.Duration
}
