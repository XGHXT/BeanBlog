package blog

import (
	"BeanBlog/internal/config"
	"BeanBlog/internal/model"
	"github.com/panjf2000/ants"
	"github.com/patrickmn/go-cache"
	"go.uber.org/dig"
	"golang.org/x/sync/singleflight"
	"gopkg.in/yaml.v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"os"
	"sync"
	"time"
)

func newCache() *cache.Cache {
	return cache.New(5*time.Minute, 10*time.Minute)
}

func newPool() *ants.Pool {
	p, err := ants.NewPool(20000)
	if err != nil {
		panic(err)
	}
	return p
}

func newDatabase(conf *config.Config) *gorm.DB {
	db, err := gorm.Open(mysql.Open(conf.Database), &gorm.Config{
		PrepareStmt: true, // 缓存每一条sql语句，提高执行速度
	})
	if err != nil {
		panic(err)
	}
	sqlDb, err := db.DB()
	if err != nil {
		panic(err)
	}
	sqlDb.SetConnMaxLifetime(time.Hour)
	sqlDb.SetMaxOpenConns(10)
	sqlDb.SetMaxIdleConns(30)
	return db
}

func newConfig() *config.Config {
	configFile := "data/conf.yml"
	content, err := os.ReadFile(configFile)
	if err != nil {
		panic(err)
	}
	var c config.Config
	err = yaml.Unmarshal(content, &c)
	if err != nil {
		panic(err)
	}
	c.ConfigFilePath = configFile
	return &c
}

func newSystem(c *config.Config, d *gorm.DB, h *cache.Cache, p *ants.Pool) *SysVariable {
	return &SysVariable{
		Config:    c,
		DB:        d,
		Cache:     h,
		SafeCache: new(singleflight.Group),
		Pool:      p,
	}
}

func provide() {
	var providers = []interface{}{
		newCache,
		newConfig,
		newDatabase,
		newSystem,
		newPool,
	}
	var err error
	for i := 0; i < len(providers); i++ {
		err = Injector.Provide(providers[i])
		if err != nil {
			panic(err)
		}
	}
	err = Injector.Invoke(func(s *SysVariable) {
		System = s
	})
	if err != nil {
		panic(err)
	}
}

func checkPoolSubmit(wg *sync.WaitGroup, err error) {
	if err != nil {
		log.Println(err)
		if wg != nil {
			wg.Done()
		}
	}
}

func init() {
	Injector = dig.New()
	provide()
	if System.DB != nil {
		if err := System.DB.AutoMigrate(model.Article{}, model.ArticleHistory{}, model.Comment{}); err != nil {
			panic(err)
		}
	}
}
