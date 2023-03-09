package model

import (
	"fmt"
	"github.com/88250/lute/editor"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/panjf2000/ants"
	"gorm.io/gorm"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode"
)

// ArticleTOC 文章标题
type ArticleTOC struct {
	Title     string
	Slug      string
	SubTitles []*ArticleTOC
	Parent    *ArticleTOC `gorm:"-"`
	Level     int         `gorm:"-"`
}

// SibilingArticle 相邻文章
type SibilingArticle struct {
	Next Article
	Prev Article
}

// Article 文章表
type Article struct {
	ID        string    `gorm:"primary_key"`
	CreatedAt time.Time `gorm:"column:created_at;index;" json:"created_at,omitempty"`
	UpdatedAt time.Time `gorm:"column:updated_at;index;" json:"updated_at,omitempty"`

	Slug       string         `form:"slug" validate:"required" gorm:"unique_index"`
	Title      string         `form:"title" validate:"required"`
	Content    string         `form:"content" validate:"required" gorm:"text"`
	TemplateID byte           `form:"template" validate:"required"`
	IsBook     bool           `form:"is_book"`
	RawTags    string         `form:"tags" gorm:"-"`
	Tags       pq.StringArray `gorm:"type:json" validate:"-" form:"-"`
	ReadNum    uint           `gorm:"default:0;"`
	CommentNum uint           `gorm:"default:0;"`
	Version    uint           `gorm:"default:1;"`
	BookRefer  *string        `form:"book_refer" validate:"omitempty,uuid4" gorm:"type:uuid;index;default:NULL"`
	IsPrivate  bool           `form:"is_private"`

	Comments         []*Comment
	ArticleHistories []*ArticleHistory
	Toc              []*ArticleTOC    `gorm:"-" validate:"-" form:"-"`
	Chapters         []*Article       `gorm:"foreignkey:BookRefer" form:"-" validate:"-"`
	Book             *Article         `gorm:"-" validate:"-" form:"-"`
	SibilingArticle  *SibilingArticle `gorm:"-" validate:"-" form:"-"`

	// for form
	NewVersion bool `gorm:"-" form:"new_version"`
}

// ArticleIndex index data
type ArticleIndex struct {
	Slug    string
	Version float64
	Title   string
}

// GetIndexID get index data id
func (t *Article) GetIndexID() string {
	return fmt.Sprintf("%s.%d", t.ID, t.Version)
}

func (t *Article) BeforeCreate(tx *gorm.DB) error {
	t.ID = uuid.New().String()
	return nil
}

// BeforeSave hook
func (t *Article) BeforeSave(tx *gorm.DB) error {
	t.Tags = strings.Split(t.RawTags, ",")
	return nil
}

// AfterFind hook
func (t *Article) AfterFind(tx *gorm.DB) error {
	t.RawTags = strings.Join(t.Tags, ",")
	return nil
}

var titleRegex = regexp.MustCompile(`^\s{0,2}(#{1,6})\s(.*)$`)

// GenTOC 生成标题树
func (t *Article) GenTOC() {
	lines := strings.Split(t.Content, "\n")
	uniqueHeadingID := make(map[string]int)
	var matches []string
	var currentToc *ArticleTOC
	for j := 0; j < len(lines); j++ {
		matches = titleRegex.FindStringSubmatch(lines[j])
		if len(matches) != 3 {
			continue
		}
		var toc ArticleTOC
		toc.Level = len(matches[1])
		toc.Title = string(matches[2])
		toc.Slug = sanitizedAnchorName(uniqueHeadingID, string(matches[2]))
		if currentToc == nil {
			t.Toc = append(t.Toc, &toc)
			currentToc = &toc
			continue
		}
		parent := currentToc
		if currentToc.Level > toc.Level {
			// 父节点
			for i := -1; i < currentToc.Level-toc.Level; i++ {
				parent = parent.Parent
				if parent == nil || parent.Level < toc.Level {
					break
				}
			}
			if parent == nil {
				t.Toc = append(t.Toc, &toc)
			} else {
				toc.Parent = parent
				parent.SubTitles = append(parent.SubTitles, &toc)
			}
		} else if currentToc.Level == toc.Level {
			// 兄弟节点
			if parent.Parent == nil {
				t.Toc = append(t.Toc, &toc)
			} else {
				toc.Parent = parent.Parent
				toc.Parent.SubTitles = append(toc.Parent.SubTitles, &toc)
			}
		} else {
			// 子节点
			toc.Parent = parent
			parent.SubTitles = append(parent.SubTitles, &toc)
		}
		currentToc = &toc
	}
}

// 生成标题 ID
func sanitizedAnchorName(unique map[string]int, text string) (ret string) {
	text = strings.TrimSpace(text)
	text = strings.ReplaceAll(text, editor.Caret, "")
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			ret += string(r)
		} else {
			ret += "-"
		}
	}
	for 0 < unique[ret] {
		ret += "-"
	}
	unique[ret] = 1
	return
}

// RelatedCount 合计专栏下文章计数
func (t *Article) RelatedCount(db *gorm.DB, pool *ants.Pool, checkPoolSubmit func(*sync.WaitGroup, error)) {
	if !t.IsBook {
		return
	}
	var wg sync.WaitGroup
	wg.Add(1)
	checkPoolSubmit(&wg, pool.Submit(func() {
		innerRelatedCount(db, t, &wg, true)
	}))
	wg.Wait()
}

func innerRelatedCount(db *gorm.DB, p *Article, wg *sync.WaitGroup, root bool) {
	var chapters []*Article
	db.Select("id,is_book,read_num,comment_num").Where("book_refer = ?", p.ID).Find(&chapters)
	for i := 0; i < len(chapters); i++ {
		if chapters[i].IsBook {
			innerRelatedCount(db, chapters[i], nil, false)
		}
		p.ReadNum += chapters[i].ReadNum
		p.CommentNum += chapters[i].CommentNum
	}
	if root {
		wg.Done()
	}
}
