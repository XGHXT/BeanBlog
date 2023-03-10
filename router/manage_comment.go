package router

import (
	"BeanBlog/internal/model"
	"BeanBlog/pkg/blog"
	"BeanBlog/pkg/paginator"
	"errors"
	"gorm.io/gorm"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func comments(c *fiber.Ctx) error {
	rawPage := c.Query("page")
	var page int64
	page, _ = strconv.ParseInt(rawPage, 10, 32)
	var cs []model.Comment
	pg := paginator.Paging(&paginator.Param{
		DB:      blog.System.DB.Preload("Article"),
		Page:    int(page),
		Limit:   15,
		OrderBy: []string{"created_at DESC"},
	}, &cs)
	c.Status(http.StatusOK).Render("admin/comments", injectSiteData(c, fiber.Map{
		"title":    "评论管理",
		"comments": cs,
		"page":     pg,
	}))
	return nil
}

func deleteComment(c *fiber.Ctx) error {
	id := c.Query("id")
	articleID := c.Query("aid")

	if len(id) < 10 || len(articleID) < 10 {
		return errors.New("error id")
	}

	tx := blog.System.DB.Begin()
	if err := tx.Delete(&model.Comment{}, "id =?", id).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Model(model.Article{}).Where("id = ?", articleID).
		UpdateColumn("comment_num", gorm.Expr("comment_num - ?", 1)).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}
	return nil
}
