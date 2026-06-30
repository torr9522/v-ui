package controller

import (
	"github.com/gin-gonic/gin"
	"x-ui/web/entity"
	"x-ui/web/service"
)

type CloudflareController struct {
	BaseController

	cloudflareService service.CloudflareService
}

type cloudflareTestRequest struct {
	Domain string `json:"domain" form:"domain"`
}

type cloudflareRecordsRequest struct {
	ZoneID string `json:"zoneId" form:"zoneId"`
	Name   string `json:"name" form:"name"`
	Type   string `json:"type" form:"type"`
}

func NewCloudflareController(g *gin.RouterGroup) *CloudflareController {
	a := &CloudflareController{}
	a.initRouter(g)
	return a
}

func (a *CloudflareController) initRouter(g *gin.RouterGroup) {
	g = g.Group("/cloudflare")
	g.Use(a.checkLogin)

	g.POST("/status", a.status)
	g.POST("/save", a.save)
	g.POST("/test", a.test)
	g.POST("/zones", a.zones)
	g.POST("/records", a.records)
}

func (a *CloudflareController) status(c *gin.Context) {
	status, err := a.cloudflareService.GetStatus()
	if err != nil {
		jsonMsg(c, "get cloudflare status", err)
		return
	}
	jsonObj(c, status, nil)
}

func (a *CloudflareController) save(c *gin.Context) {
	cfg := &entity.CloudflareConfig{}
	if err := c.ShouldBind(cfg); err != nil {
		jsonMsg(c, "save cloudflare config", err)
		return
	}
	if err := a.cloudflareService.SaveConfig(cfg); err != nil {
		jsonMsg(c, "save cloudflare config", err)
		return
	}
	status, err := a.cloudflareService.GetStatus()
	if err != nil {
		jsonMsg(c, "get cloudflare status", err)
		return
	}
	jsonObj(c, status, nil)
}

func (a *CloudflareController) test(c *gin.Context) {
	req := &cloudflareTestRequest{}
	if err := c.ShouldBind(req); err != nil {
		jsonMsg(c, "test cloudflare config", err)
		return
	}
	result, err := a.cloudflareService.TestConnection(req.Domain)
	if err != nil {
		jsonMsg(c, "test cloudflare config", err)
		return
	}
	jsonObj(c, result, nil)
}

func (a *CloudflareController) zones(c *gin.Context) {
	zones, err := a.cloudflareService.GetZones()
	if err != nil {
		jsonMsg(c, "get cloudflare zones", err)
		return
	}
	jsonObj(c, zones, nil)
}

func (a *CloudflareController) records(c *gin.Context) {
	req := &cloudflareRecordsRequest{}
	if err := c.ShouldBind(req); err != nil {
		jsonMsg(c, "get cloudflare records", err)
		return
	}
	records, err := a.cloudflareService.GetDNSRecords(req.ZoneID, req.Name, req.Type)
	if err != nil {
		jsonMsg(c, "get cloudflare records", err)
		return
	}
	jsonObj(c, records, nil)
}
