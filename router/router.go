package router

import (
	"BeanBlog/pkg/blog"
	"BeanBlog/pkg/log"
	"BeanBlog/pkg/trans"
	gv "github.com/go-playground/validator"
	"github.com/gofiber/fiber/v2"
	"net/http"
	"sync"
)

var validator = gv.New()

func RegisterRoutes(app *fiber.App) {
	app.Static("/static", "resource/static")

	//app.Get("/", page404)
	//app.Get("/feed/:format?", feedHandler)
	//app.Get("/archives/:page?", archive)
	//app.Get("/search/", search)
	//app.Get("/tags/:tag/:page?", tags)
	//app.Get("/tags/", tagsCloud)

	app.Get("/login", guestRequired, login)
	app.Post("/login", guestRequired, loginHandler)
	app.Post("/logout", loginRequired, logoutHandler)
	//app.Post("/count", count)
	//app.Post("/comment", commentHandler)
	//app.Static("/static", "resource/static")
	//app.Static("/upload", "data/upload")

	admin := app.Group("/admin", loginRequired)
	admin.Get("/", manager)
	admin.Get("/publish", publish)
	admin.Post("/publish", publishHandler)
	//admin.Get("/rebuild-full-text-search", rebuildFullTextSearch)
	//admin.Post("/upload", upload)
	//admin.Post("/fetch", fetch)
	admin.Get("/comments", comments)
	admin.Delete("/comments", deleteComment)
	admin.Get("/articles", manageArticle)
	admin.Delete("/articles", deleteArticle)
	//admin.Get("/media", media)
	//admin.Delete("/media", mediaHandler)
	admin.Get("/settings", settings)
	admin.Post("/settings", settingsHandler)
	admin.Get("/tags", tagsManagePage)
	admin.Delete("/tags", deleteTag)
	admin.Patch("/tags", renameTag)

	app.Use(page404)
}

func guestRequired(c *fiber.Ctx) error {
	if c.Locals(blog.CtxAuthorized).(bool) {
		c.Redirect("/admin/", http.StatusFound)
		return nil
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
		log.Error(err.Error())
		if wg != nil {
			wg.Done()
		}
	}
}

// injectSiteData 渲染数据
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
