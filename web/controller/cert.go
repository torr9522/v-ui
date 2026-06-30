package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"x-ui/web/entity"
	"x-ui/web/service"
)

type CertController struct {
	BaseController

	certService service.CertService
}

type certDomainRequest struct {
	Domain string `json:"domain" form:"domain"`
}

type certUploadRequest struct {
	CertPEM string `json:"certPem" form:"certPem"`
	KeyPEM  string `json:"keyPem" form:"keyPem"`
}

type certIssueHTTPRequest struct {
	Domain  string `json:"domain" form:"domain"`
	Email   string `json:"email" form:"email"`
	Staging bool   `json:"staging" form:"staging"`
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
	g.POST("/upload", a.upload)
	g.POST("/issueHttp", a.issueHTTP)
	g.POST("/enableHttps", a.enableHTTPS)
	g.POST("/disableHttps", a.disableHTTPS)
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

func (a *CertController) upload(c *gin.Context) {
	req := &certUploadRequest{}
	if err := c.ShouldBind(req); err != nil {
		jsonMsg(c, "upload certificate", err)
		return
	}
	if err := a.certService.UploadCertificate(req.CertPEM, req.KeyPEM); err != nil {
		c.JSON(http.StatusOK, entity.Msg{Success: false, Msg: err.Error(), Obj: nil})
		return
	}
	status, err := a.certService.GetStatus()
	if err != nil {
		jsonMsg(c, "get cert status", err)
		return
	}
	c.JSON(http.StatusOK, entity.Msg{Success: true, Msg: "upload certificate success", Obj: gin.H{"status": status, "applied": false}})
}

func (a *CertController) issueHTTP(c *gin.Context) {
	req := &certIssueHTTPRequest{}
	if err := c.ShouldBind(req); err != nil {
		jsonMsg(c, "issue http certificate", err)
		return
	}
	result, err := a.certService.IssueHTTP(req.Domain, req.Email, req.Staging)
	if err != nil {
		c.JSON(http.StatusOK, entity.Msg{Success: false, Msg: err.Error(), Obj: nil})
		return
	}
	c.JSON(http.StatusOK, entity.Msg{Success: true, Msg: "issue http certificate success", Obj: result})
}

func (a *CertController) enableHTTPS(c *gin.Context) {
	applied, err := a.certService.EnableHTTPS()
	if err != nil {
		c.JSON(http.StatusOK, entity.Msg{Success: false, Msg: err.Error(), Obj: nil})
		return
	}
	status, err := a.certService.GetStatus()
	if err != nil {
		jsonMsg(c, "get cert status", err)
		return
	}
	c.JSON(http.StatusOK, entity.Msg{Success: true, Msg: "enable https success", Obj: gin.H{"status": status, "applied": applied}})
}

func (a *CertController) disableHTTPS(c *gin.Context) {
	applied, err := a.certService.DisableHTTPS()
	if err != nil {
		c.JSON(http.StatusOK, entity.Msg{Success: false, Msg: err.Error(), Obj: nil})
		return
	}
	status, err := a.certService.GetStatus()
	if err != nil {
		jsonMsg(c, "get cert status", err)
		return
	}
	c.JSON(http.StatusOK, entity.Msg{Success: true, Msg: "disable https success", Obj: gin.H{"status": status, "applied": applied}})
}
