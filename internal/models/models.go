package models

import "time"

type PVZInfo struct {
	PVZ        PVZ
	Receptions []ReceptionInfo
}

type PVZ struct {
	Id               string
	RegistrationDate time.Time
	City             string
}

type Reception struct {
	Id              string
	DateTime        time.Time
	PvzId           string
	ReceptionStatus string
}

type User struct {
	Id           string
	Email        string
	PasswordHash string
	Role         string
}

type ReceptionInfo struct {
	Reception Reception
	Products  []Product
}

type Product struct {
	Id          string
	DateTime    time.Time
	Type        string
	ReceptionId string
}
