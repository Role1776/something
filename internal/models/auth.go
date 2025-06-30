package models

import "time"

type FirstAuth struct {
	Login    string `json:"login" binding:"required,min=2,max=86"`
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=8,max=86"`
	Verified bool   `json:"verified" binding:"-"`
}

type SecondAuth struct {
	Login    string `json:"login" binding:"required,min=2,max=86"`
	Password string `json:"password" binding:"required,min=8,max=86"`
	DeviceID string `json:"device_id" binding:"required"`
}

type Email struct {
	Email string `json:"email" binding:"required,email,max=255"`
}

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshToken struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type RefreshTokenData struct {
	UserID    int
	DeviceID  string
	ExpiresAt time.Time
	CreatedAt time.Time
}

type VerificationCode struct {
	Code string `json:"code" binding:"required,min=6,max=6"`
}
