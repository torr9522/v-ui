package controller

import (
	"github.com/gin-gonic/gin"
)

type XUIController struct {
	BaseController

	inboundController  *InboundController
	outboundController *OutboundController
	routingController  *RoutingController
	settingController  *SettingController
	accessIPController *AccessIPController
}

func NewXUIController(g *gin.RouterGroup) *XUIController {
	a := &XUIController{}
	a.initRouter(g)
	return a
}

func (a *XUIController) initRouter(g *gin.RouterGroup) {
	g = g.Group("/xui")
	g.Use(a.checkLogin)

	g.GET("/", a.index)
	g.GET("/inbounds", a.inbounds)
	g.GET("/outbounds", a.outbounds)
	g.GET("/routing", a.routing)
	g.GET("/access-ips", a.accessIPs)
	g.GET("/setting", a.setting)

	a.inboundController = NewInboundController(g)
	a.outboundController = NewOutboundController(g)
	a.routingController = NewRoutingController(g)
	a.settingController = NewSettingController(g)
	a.accessIPController = NewAccessIPController(g)
}

func (a *XUIController) index(c *gin.Context) {
	html(c, "index.html", "C-UI - 系统状态", nil)
}

func (a *XUIController) inbounds(c *gin.Context) {
	html(c, "inbounds.html", "C-UI - 入站列表", nil)
}

func (a *XUIController) outbounds(c *gin.Context) {
	html(c, "outbounds.html", "C-UI - 出站", nil)
}

func (a *XUIController) routing(c *gin.Context) {
	html(c, "routing.html", "C-UI - 路由", nil)
}

func (a *XUIController) accessIPs(c *gin.Context) {
	html(c, "access_ips.html", "C-UI - 访问 IP", nil)
}

func (a *XUIController) setting(c *gin.Context) {
	html(c, "setting.html", "C-UI - 设置", nil)
}
