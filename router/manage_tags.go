package router

import (
	"BeanBlog/pkg/blog"
	"errors"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func tagsManagePage(c *fiber.Ctx) error {
	var tags []string
	rows, err := blog.System.DB.Raw(`select count(*), unnest(articles.tags) t from articles group by t order by count desc`).Rows()
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var line string
			var count int
			rows.Scan(&count, &line)
			tags = append(tags, line)
		}
	}
	c.Status(http.StatusOK).Render("admin/tags", injectSiteData(c, fiber.Map{
		"title": "标签管理",
		"tags":  tags,
	}))
	return nil
}

func deleteTag(c *fiber.Ctx) error {
	tagName := c.Query("tagName")
	return blog.System.DB.Exec("UPDATE articles SET tags = array_remove(tags, ?);", tagName).Error
}

func renameTag(c *fiber.Ctx) error {
	oldTagName := c.Query("oldTagName")
	newTagName := strings.TrimSpace(c.Query("newTagName"))
	if newTagName == "" {
		return errors.New("empty tag name")
	}
	return blog.System.DB.Exec("UPDATE articles SET tags = array_replace(tags, ?, ?);", oldTagName, newTagName).Error
}
