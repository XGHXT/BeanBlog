package router

import (
	"BeanBlog/internal/model"
	"BeanBlog/pkg/blog"
	"BeanBlog/pkg/paginator"

	"errors"
	"github.com/gofiber/fiber/v2"
	"net/http"
	"strconv"
	"unicode/utf8"
)

func manageArticle(c *fiber.Ctx) error {
	rawPage := c.Query("page")
	var page int64
	page, _ = strconv.ParseInt(rawPage, 10, 32)
	var as []model.Article
	pg := paginator.Paging(&paginator.Param{
		DB:      blog.System.DB,
		Page:    int(page),
		Limit:   15,
		OrderBy: []string{"created_at DESC"},
	}, &as)
	for i := 0; i < len(as); i++ {
		as[i].RelatedCount(blog.System.DB, blog.System.Pool, checkPoolSubmit)
	}
	c.Status(http.StatusOK).Render("admin/articles", injectSiteData(c, fiber.Map{
		"title":    "文章管理",
		"articles": as,
		"page":     pg,
	}))
	return nil
}

func publish(c *fiber.Ctx) error {
	id := c.Query("id")
	var article model.Article
	if id != "" {
		blog.System.DB.Take(&article, "id = ?", id)
	}
	c.Status(http.StatusOK).Render("admin/publish", injectSiteData(c, fiber.Map{
		"title":     "文章发布",
		"templates": blog.Templates,
		"article":   article,
	}))
	return nil
}

func deleteArticle(c *fiber.Ctx) error {
	id := c.Query("id")
	if len(id) < 10 {
		return errors.New("error article id")
	}
	var a model.Article
	if err := blog.System.DB.Select("id").Preload("ArticleHistories").Take(&a, "id = ?", id).Error; err != nil {
		return err
	}
	var indexIDs []string
	indexIDs = append(indexIDs, a.GetIndexID())
	tx := blog.System.DB.Begin()
	if err := tx.Delete(model.Article{}, "id = ?", a.ID).Error; err != nil {
		tx.Rollback()
		return err
	}
	// delete article history
	for i := 0; i < len(a.ArticleHistories); i++ {
		indexIDs = append(indexIDs, a.ArticleHistories[i].GetIndexID())
	}
	if err := tx.Delete(model.ArticleHistory{}, "article_id = ?", a.ID).Error; err != nil {
		tx.Rollback()
		return err
	}
	// delete comments
	if err := tx.Delete(model.Comment{}, "article_id = ?", a.ID).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}
	// delete full-text search data
	//for i := 0; i < len(indexIDs); i++ {
	//	blog.System.Search.Delete(indexIDs[i])
	//}
	return nil
}

type publishArticle struct {
	ID         string `form:"id"`
	Title      string `form:"title"`
	Slug       string `form:"slug"`
	Content    string `form:"content"`
	Template   byte   `form:"template"`
	Tags       string `form:"tags"`
	IsBook     bool   `form:"is_book"`
	IsPrivate  bool   `form:"is_private"`
	BookRefer  string `form:"book_refer"`
	NewVersion bool   `form:"new_version"`
}

func publishHandler(c *fiber.Ctx) error {
	var pa publishArticle
	if err := c.BodyParser(&pa); err != nil {
		return err
	}
	if err := validator.StructCtx(c.Context(), &pa); err != nil {
		return err
	}
	var bookRefer *string
	if pa.BookRefer != "" {
		bookRefer = &pa.BookRefer
	}
	// edit article
	newArticle := &model.Article{
		ID:         pa.ID,
		Title:      pa.Title,
		Slug:       pa.Slug,
		Content:    clearNonUTF8Chars(pa.Content),
		NewVersion: pa.NewVersion,
		TemplateID: pa.Template,
		IsBook:     pa.IsBook,
		IsPrivate:  pa.IsPrivate,
		RawTags:    pa.Tags,
		BookRefer:  bookRefer,
		Version:    1,
	}
	if originalArticle, err := fetchOriginArticle(newArticle); err != nil {
		return err
	} else {
		// save edit history && article
		tx := blog.System.DB.Begin()
		err = tx.Save(&newArticle).Error
		if pa.NewVersion && err == nil {
			var history model.ArticleHistory
			history.Content = originalArticle.Content
			history.Version = originalArticle.Version
			history.ArticleID = originalArticle.ID
			err = tx.Save(&history).Error
		}
		if err != nil {
			tx.Rollback()
			return err
		}
		if err = tx.Commit().Error; err != nil {
			return err
		}
		// indexing serch engine
		//numBefore, _ := blog.System.Search.DocCount()
		//errIndex := blog.System.Search.Index(newArticle.GetIndexID(), newArticle)
		//numAfter, _ := blog.System.Search.DocCount()
		//log.Info("Doc %s indexed %d --> %d %+v\n", newArticle.GetIndexID(), numBefore, numAfter, errIndex)
	}
	return nil
}

func fetchOriginArticle(af *model.Article) (model.Article, error) {
	if af.ID == "" {
		return model.Article{}, nil
	}
	var originArticle model.Article
	if err := blog.System.DB.Take(&originArticle, "id = ?", af.ID).Error; err != nil {
		return model.Article{}, err
	}
	af.Version = originArticle.Version
	af.CommentNum = originArticle.CommentNum
	af.ReadNum = originArticle.ReadNum
	if af.NewVersion {
		af.Version = originArticle.Version + 1
	}
	return originArticle, nil
}

func clearNonUTF8Chars(s string) string {
	v := make([]rune, 0, len(s))
	for i, r := range s {
		// 清理非 UTF-8 字符
		if r == utf8.RuneError {
			_, size := utf8.DecodeRuneInString(s[i:])
			if size == 1 {
				continue
			}
		}
		// 清理 backspace
		if r == '\b' {
			continue
		}
		v = append(v, r)
	}
	return string(v)
}
