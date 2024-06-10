package models

type Player struct {
	PlayerID uint   `json:"player_id" gorm:"primaryKey"`
	Username string `json:"username" gorm:"size:50; not null; uniqueIndex"`
	TeamName string `json:"team_name" gorm:"size:50; not null"`
	Ranking  int    `json:"ranking" gorm:"not null"`
	Score    int    `json:"score" gorm:"not null"`
}
