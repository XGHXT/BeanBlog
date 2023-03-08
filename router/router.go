package router

import (
	"BeanBlog/pkg/trans"
	"BeanBlog/tools/blog"
	gv "github.com/go-playground/validator"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html"
	"github.com/jinzhu/gorm"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

var validator = gv.New()

func RegisterRoutes(engine *html.Engine) *fiber.App {
	dbErrors := map[error]bool{
		gorm.ErrCantStartTransaction: true,
		gorm.ErrInvalidSQL:           true,
		gorm.ErrInvalidTransaction:   true,
		gorm.ErrUnaddressable:        true,
	}
	app := fiber.New(fiber.Config{
		EnableTrustedProxyCheck: blog.System.Config.EnableTrustedProxyCheck,
		TrustedProxies:          blog.System.Config.TrustedProxies,
		ProxyHeader:             blog.System.Config.ProxyHeader,
		Views:                   engine,
		ErrorHandler: func(c *fiber.Ctx, e error) error {
			// 404 页面
			if e == gorm.ErrRecordNotFound {
				return page404(c)
			}
			title := "Unknown error"
			errMsg := e.Error()
			if dbErrors[e] {
				title = "DB error"
				errMsg = "Please contact the webmaster"
			}
			if strings.Contains(string(c.Request().Header.Peek("Accept")), "html") {
				return c.Status(http.StatusInternalServerError).Render("default/error", injectSiteData(c, fiber.Map{
					"title": title,
					"msg":   errMsg,
				}))
			}
			_, e = c.Status(http.StatusInternalServerError).WriteString(errMsg)
			return e
		},
	})
	if blog.System.Config.Debug {
		app.Use(logger.New())
		engine.Reload(true)
		engine.Debug(true)
	}
	app.Use(auth)

	//app.Get("/", index)
	//app.Get("/feed/:format?", feedHandler)
	//app.Get("/archives/:page?", archive)
	//app.Get("/search/", search)
	//app.Get("/tags/:tag/:page?", tags)
	//app.Get("/tags/", tagsCloud)
	//app.Get("/login", guestRequired, login)
	//app.Post("/login", guestRequired, loginHandler)
	//app.Post("/logout", loginRequired, logoutHandler)
	//app.Post("/count", count)
	//app.Post("/comment", commentHandler)
	//app.Static("/static", "resource/static")
	//app.Static("/upload", "data/upload")

	admin := app.Group("/admin", loginRequired)
	admin.Get("/", manager)
	//admin.Get("/publish", publish)
	//admin.Post("/publish", publishHandler)
	//admin.Get("/rebuild-full-text-search", rebuildFullTextSearch)
	//admin.Post("/upload", upload)
	//admin.Post("/fetch", fetch)
	//admin.Get("/comments", comments)
	//admin.Delete("/comments", deleteComment)
	//admin.Get("/articles", manageArticle)
	//admin.Delete("/articles", deleteArticle)
	//admin.Get("/media", media)
	//admin.Delete("/media", mediaHandler)
	//admin.Get("/settings", settings)
	//admin.Post("/settings", settingsHandler)
	//admin.Get("/tags", tagsManagePage)
	//admin.Delete("/tags", deleteTag)
	//admin.Patch("/tags", renameTag)

	app.Use(page404)

	return app
}

func auth(c *fiber.Ctx) error {
	token := c.Cookies(blog.AuthCookie)
	if len(token) > 0 && token == blog.System.Config.User.Token && blog.System.Config.User.TokenExpires > time.Now().Unix() {
		c.Locals(blog.CtxAuthorized, true)
	} else {
		c.Locals(blog.CtxAuthorized, false)
	}
	return c.Next()
}

func page404(c *fiber.Ctx) error {
	c.Status(http.StatusNotFound).Render("default/error", injectSiteData(c, fiber.Map{
		"title": trans.WordTrans["404_title"],
		"msg":   trans.WordTrans["404_msg"],
	}))
	return nil
}

func loginRequired(c *fiber.Ctx) error {
	if !c.Locals(blog.CtxAuthorized).(bool) {
		c.Redirect("/login", http.StatusFound)
		return nil
	}
	return c.Next()
}

func checkPoolSubmit(wg *sync.WaitGroup, err error) {
	if err != nil {
		log.Println(err)
		if wg != nil {
			wg.Done()
		}
	}
}

func injectSiteData(c *fiber.Ctx, data fiber.Map) fiber.Map {
	var title, keywords, desc string

	// custom title
	if k, ok := data["title"]; ok && k.(string) != "" {
		title = data["title"].(string) + " | " + blog.System.Config.Site.SpaceName
	} else {
		title = blog.System.Config.Site.SpaceName
	}
	// custom keywords
	if k, ok := data["keywords"]; ok && k.(string) != "" {
		keywords = data["keywords"].(string)
	} else {
		keywords = blog.System.Config.Site.SpaceKeywords
	}
	// custom desc
	if k, ok := data["desc"]; ok && k.(string) != "" {
		desc = data["desc"].(string)
	} else {
		desc = blog.System.Config.Site.SpaceDesc
	}

	var soli = make(map[string]interface{})
	soli["Conf"] = blog.System.Config
	soli["Title"] = title
	soli["Keywords"] = keywords
	soli["Desc"] = desc
	soli["Login"] = c.Locals(blog.CtxAuthorized)
	soli["Data"] = data

	return soli
}
