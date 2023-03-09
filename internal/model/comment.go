package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

// Comment 评论表
type Comment struct {
	ID        string    `gorm:"primary_key"`
	CreatedAt time.Time `gorm:"column:created_at;index;" json:"created_at,omitempty"`

	ReplyTo   *string `gorm:"type:uuid;index;default:NULL" form:"reply_to"`
	Nickname  string  `form:"nickname" validate:"required"`
	Content   string  `form:"content" validate:"required" gorm:"text"`
	Website   string  `form:"website"`
	Version   uint    `form:"-"`
	Email     string  `form:"email"`
	IP        string  `gorm:"inet"`
	UserAgent string
	IsAdmin   bool

	ArticleID     *string `gorm:"type:uuid;index;default:NULL" form:"article_id" validate:"required,uuid"`
	Article       *Article
	ChildComments []*Comment `gorm:"foreignkey:ReplyTo" form:"-" validate:"-"`
}

func (c *Comment) BeforeCreate(tx *gorm.DB) error {
	c.ID = uuid.New().String()
	return nil
}
