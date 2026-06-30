package controller

import (
	"github.com/gin-gonic/gin"
	"x-ui/web/service"
)

type CertController struct {
	BaseController

	certService service.CertService
}

type certDomainRequest struct {
	Domain string `json:"domain" form:"domain"`
}

func NewCertController(g *gin.RouterGroup) *CertController {
	a := &CertController{}
	a.initRouter(g)
	return a
}

func (a *CertController) initRouter(g *gin.RouterGroup) {
	g = g.Group("/cert")
	g.Use(a.checkLogin)

	g.POST("/status", a.getStatus)
	g.POST("/setDomain", a.setDomain)
	g.POST("/checkDomain", a.checkDomain)
}

func (a *CertController) getStatus(c *gin.Context) {
	status, err := a.certService.GetStatus()
	if err != nil {
		jsonMsg(c, "get cert status", err)
		return
	}
	jsonObj(c, status, nil)
}

func (a *CertController) setDomain(c *gin.Context) {
	req := &certDomainRequest{}
	if err := c.ShouldBind(req); err != nil {
		jsonMsg(c, "set cert domain", err)
		return
	}
	err := a.certService.SetDomain(req.Domain)
	if err != nil {
		jsonMsg(c, "set cert domain", err)
		return
	}
	status, err := a.certService.GetStatus()
	if err != nil {
		jsonMsg(c, "get cert status", err)
		return
	}
	jsonObj(c, status, nil)
}

func (a *CertController) checkDomain(c *gin.Context) {
	req := &certDomainRequest{}
	if err := c.ShouldBind(req); err != nil {
		jsonMsg(c, "check cert domain", err)
		return
	}
	result, err := a.certService.CheckDomain(req.Domain)
	if err != nil {
		jsonMsg(c, "check cert domain", err)
		return
	}
	jsonObj(c, result, nil)
}
