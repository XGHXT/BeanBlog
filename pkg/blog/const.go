package blog

import (
	"BeanBlog/internal/config"
	"github.com/panjf2000/ants"
	"github.com/patrickmn/go-cache"
	"go.uber.org/dig"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

const (
	// CtxAuthorized 用户已认证
	CtxAuthorized = "bean_auth"
	// AuthCookie 用户认证使用的Cookie名
	AuthCookie = "bean_bean"
	// CacheKeyPrefixRelatedChapters 缓存键前缀：章节
	CacheKeyPrefixRelatedChapters = "ckprc"
	// CacheKeyPrefixRelatedArticle 缓存键前缀：文章
	CacheKeyPrefixRelatedArticle = "ckpra"
	// CacheKeyPrefixRelatedSiblingArticle 缓存键前缀：相邻文章
	CacheKeyPrefixRelatedSiblingArticle = "ckprsa"
	// RequestId 请求id名称
	RequestId = "request_id"
	// TimeLayout 时间格式
	TimeLayout   = "2006-01-02 15:04:05"
	TimeLayoutMs = "2006-01-02 15:04:05.000"
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
