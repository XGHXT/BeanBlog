package server

import (
	"BeanBlog/internal/model"
	"BeanBlog/router"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/88250/lute"
	"github.com/gofiber/template/html"
	"github.com/microcosm-cc/bluemonday"
	"html/template"
	"reflect"
	"strings"
	"time"
)

var bluemondayPolicy = bluemonday.UGCPolicy()
var luteEngine = lute.New()

func init() {
	luteEngine.SetCodeSyntaxHighlight(false)
	luteEngine.SetHeadingAnchor(true)
	luteEngine.SetHeadingID(true)
	luteEngine.SetSub(true)
	luteEngine.SetSup(true)
}

func Serve(endRun chan error) {
	engine := html.New("resource/theme", ".html")
	setFuncMap(engine)
	app := router.RegisterRoutes(engine)

	//app.Get("/:slug/:version?", article)

	go func() {
		endRun <- app.Listen(":8080")
	}()
}

func setFuncMap(engine *html.Engine) {
	funcMap := template.FuncMap{
		"md5": func(origin string) string {
			hasher := md5.New()
			hasher.Write([]byte(origin))
			return hex.EncodeToString(hasher.Sum(nil))
		},
		"add": func(a, b int) int {
			return a + b
		},
		"uint2str": func(i uint) string {
			return fmt.Sprintf("%d", i)
		},
		"int2str": func(i int) string {
			return fmt.Sprintf("%d", i)
		},
		"json": func(x interface{}) string {
			b, _ := json.Marshal(x)
			return string(b)
		},
		"unsafe": func(raw string) template.HTML {
			return template.HTML(raw)
		},
		"tf": func(t time.Time, f string) string {
			return t.Format(f)
		},
		"ugcPolicy": ugcPolicy,
		"md":        mdRender,
		"articleIdx": func(t model.Article) string {
			return t.GetIndexID()
		},
		"last": func(x int, a interface{}) bool {
			return x == reflect.ValueOf(a).Len()-1
		},
		"trim": strings.TrimSpace,
	}
	for name, fn := range funcMap {
		engine.AddFunc(name, fn)
	}
}

func mdRender(id string, raw string) string {
	return luteEngine.MarkdownStr(id, raw)
}

func ugcPolicy(raw string) string {
	return bluemondayPolicy.Sanitize(raw)
}

