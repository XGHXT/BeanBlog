package router

import (
	"BeanBlog/internal/model"
	"BeanBlog/pkg/blog"
	"errors"
	"gorm.io/gorm"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

type loginForm struct {
	Email    string `form:"email" validate:"required,email"`
	Password string `form:"password" validate:"required"`
	Remember string `form:"remember"`
}

func loginHandler(c *fiber.Ctx) error {
	var lf loginForm
	if err := c.BodyParser(&lf); err != nil {
		return err
	}
	if err := validator.StructCtx(c.Context(), &lf); err != nil {
		return err
	}
	if lf.Email != blog.System.Config.User.Email ||
		bcrypt.CompareHashAndPassword([]byte(blog.System.Config.User.Password),
			[]byte(lf.Password)) != nil {
		return errors.New("invalid email or password")
	}
	token, err := bcrypt.GenerateFromPassword([]byte(lf.Password+time.Now().String()), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	blog.System.Config.User.Token = string(token)
	var expires time.Time
	if lf.Remember == "on" {
		expires = time.Now().AddDate(0, 0, 3)
	} else {
		expires = time.Now().Add(time.Hour * 5)
	}
	blog.System.Config.User.TokenExpires = expires.Unix()
	c.Cookie(&fiber.Cookie{
		Name:    blog.AuthCookie,
		Value:   string(token),
		Expires: expires,
	})
	blog.System.Config.Save()
	c.Redirect("/admin/", http.StatusFound)
	return nil
}

func login(c *fiber.Ctx) error {
	c.Status(http.StatusOK).Render("admin/login", injectSiteData(c, fiber.Map{}))
	return nil
}

func logoutHandler(c *fiber.Ctx) error {
	blog.System.Config.User.TokenExpires = time.Now().Unix()
	blog.System.Config.User.Token = ""
	blog.System.Config.Save()
	c.Redirect("/", http.StatusFound)
	return nil
}

func index(c *fiber.Ctx) error {
	var as []model.Article
	blog.System.DB.Order("created_at DESC").Limit(10).Find(&as)
	for i := 0; i < len(as); i++ {
		as[i].RelatedCount(blog.System.DB, blog.System.Pool, checkPoolSubmit)
	}
	c.Status(http.StatusOK).Render("default/index", injectSiteData(c, fiber.Map{
		"title":    "首页",
		"articles": as,
	}))
	return nil
}

func count(c *fiber.Ctx) error {
	if c.Query("slug") == "" {
		return nil
	}
	// key := c.IP() + c.Query("slug")
	// if _, ok := blog.System.Cache.Get(key); ok {
	// 	return nil
	// }
	// blog.System.Cache.Set(key, nil, time.Hour*20)
	blog.System.DB.Model(model.Article{}).
		Where("slug = ?", c.Query("slug")).
		UpdateColumn("read_num", gorm.Expr("read_num + ?", 1))
	return nil
}
