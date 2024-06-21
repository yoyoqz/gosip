package main

import (
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/panjjo/gosip/api"
	"github.com/panjjo/gosip/api/middleware"
	"github.com/panjjo/gosip/docs"
	"github.com/panjjo/gosip/m"
	sipapi "github.com/panjjo/gosip/sip"
	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title          GoSIP
// @version        2.0
// @description    GB28181 SIP服务端.
// @termsOfService https://github.com/panjjo/gosip

// @contact.name  GoSIP
// @contact.url   https://github.com/panjjo/gosip
// @contact.email panjjo@vip.qq.com

// @license.name Apache 2.0
// @license.url  http://www.apache.org/licenses/LICENSE-2.0.html

// @host     localhost:8090
// @BasePath /

// @securityDefinitions.basic BasicAuth

func main() {
	//pprof
	go func() {
		http.ListenAndServe("0.0.0.0:6060", nil)
	}()

	sipapi.Start()

	r := gin.Default()
	r.Use(middleware.Recovery)

	// programatically set swagger info
	docs.SwaggerInfo.Title = "GoSIP"
	docs.SwaggerInfo.Description = "GB28181 SIP服务端"
	docs.SwaggerInfo.Version = "1.0"

	if os.Getenv("GOSIPSERVER") != "" {
		docs.SwaggerInfo.Host = os.Getenv("GOSIPSERVER")
	} else {
		docs.SwaggerInfo.Host = "localhost:8090"
	}

	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	api.Init(r)

	logrus.Fatal(r.Run(m.MConfig.API))
	// restapi.RestfulAPI()
}

func init() {
	m.LoadConfig()
	_cron()
}

func _cron() {
	fixTime := "@every 15m"
	if os.Getenv("CRONTIME") != "" {
		fixTime = os.Getenv("CRONTIME")
	}

	c := cron.New() // 新建一个定时任务对象
	//c.AddFunc("0 */60 * * * *", sipapi.CheckStreams) // 定时关闭推送流
	//c.AddFunc("@every 1h30m"
	c.AddFunc(fixTime, sipapi.CheckStreams)
	//c.AddFunc("0 */5 * * * *", sipapi.ClearFiles)    // 定时清理录制文件
	c.Start()
}
