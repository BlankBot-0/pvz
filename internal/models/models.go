package models

import "time"

type PVZInfo struct {
	PVZ        PVZ
	Receptions []ReceptionInfo
}

type PVZ struct {
	ID               string
	RegistrationDate time.Time
	City             string
}

type Reception struct {
	ID              string
	DateTime        time.Time
	PvzID           string
	ReceptionStatus string
}

type User struct {
	ID           string
	Email        string
	PasswordHash string
	Role         string
}

type ReceptionInfo struct {
	Reception Reception
	Products  []Product
}

type Product struct {
	ID          string
	DateTime    time.Time
	Type        string
	ReceptionID string
}

const (
	ReceptionStatusInProgress = "in_progress"
	ReceptionStatusClosed     = "closed"
)
