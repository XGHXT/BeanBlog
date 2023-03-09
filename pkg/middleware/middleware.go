package middleware

import (
	"BeanBlog/pkg/blog"
	"github.com/gofiber/fiber/v2"
	"time"
)

func AuthAdmin(c *fiber.Ctx) error {
	token := c.Cookies(blog.AuthCookie)
	if len(token) > 0 && token == blog.System.Config.User.Token && blog.System.Config.User.TokenExpires > time.Now().Unix() {
		c.Locals(blog.CtxAuthorized, true)
	} else {
		c.Locals(blog.CtxAuthorized, false)
	}
	return c.Next()
}
