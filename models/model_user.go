package models

type User struct {
	UserID   uint   `json:"user_id" gorm:"primaryKey"`
	Email    string `json:"email" gorm:"size:255; not null; uniqueIndex"`
	Username string `json:"username" gorm:"size:255; not null"`
	Password string `json:"password" gorm:"size:255; not null"`
}
