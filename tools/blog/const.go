package blog

import (
	"BeanBlog/internal/config"
	"github.com/jinzhu/gorm"
	"github.com/panjf2000/ants"
	"github.com/patrickmn/go-cache"
	"go.uber.org/dig"
	"golang.org/x/sync/singleflight"
)

const (
	// CtxAuthorized 用户已认证
	CtxAuthorized = "cazed"
	// AuthCookie 用户认证使用的Cookie名
	AuthCookie = "bean_bean"
	// CacheKeyPrefixRelatedChapters 缓存键前缀：章节
	CacheKeyPrefixRelatedChapters = "ckprc"
	// CacheKeyPrefixRelatedArticle 缓存键前缀：文章
	CacheKeyPrefixRelatedArticle = "ckpra"
	// CacheKeyPrefixRelatedSiblingArticle 缓存键前缀：相邻文章
	CacheKeyPrefixRelatedSiblingArticle = "ckprsa"
)

// SysVariable 全局变量
type SysVariable struct {
	Config    *config.Config
	DB        *gorm.DB
	Cache     *cache.Cache
	SafeCache *singleflight.Group
	Pool      *ants.Pool
}

// Injector 运行时依赖注入
var Injector *dig.Container

// System 全局变量
var System *SysVariable

// Templates 文章模板
var Templates = map[byte]string{
	1: "Article template",
	2: "Page template",
}
