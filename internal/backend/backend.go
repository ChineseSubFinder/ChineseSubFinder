package backend

import (
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/backend/routers"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// StartBackEnd 开启后端的服务器
func StartBackEnd(httpPort int) {

	log_helper.SetLoggerName(log_helper.LogNameBackEnd)

	engine := gin.Default()
	// 默认所有都通过
	engine.Use(cors.Default())
	routers.InitRouter(engine)

	// listen and serve on 0.0.0.0:8080(default)
	log_helper.GetLogger().Infoln("Try Start Server At Port", httpPort)
	err := engine.Run(":" + fmt.Sprintf("%d", httpPort))
	if err != nil {
		log_helper.GetLogger().Errorln("Start Server At Port", httpPort, "Error", err)
	}
}
