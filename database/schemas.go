package database

import "gorm.io/gorm"

type User struct {
	gorm.Model
	GoogleID string  `gorm:"uniqueIndex;not null"`
	Email    string  `gorm:"uniqueIndex;not null"`
	Images   []Image `gorm:"foreignKey:UserGoogleID;references:GoogleID"`
}

type Image struct {
	gorm.Model
	UserGoogleID string `gorm:"index"`
	Base64String string `gorm:"type:text;not null"`
}

func GetModels() []interface{} {
	return []interface{}{
		&User{},
		&Image{},
	}
}
