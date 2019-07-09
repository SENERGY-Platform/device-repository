package model

type AuthAction int

const (
	READ AuthAction = iota
	WRITE
	EXECUTE
	ADMINISTRATE
)
