package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"x-ui/database/model"
	"x-ui/web/entity"
	"x-ui/web/service"
)

type OutboundController struct {
	outboundService service.OutboundService
}

type outboundListRequest struct {
	IncludeSystem bool `json:"includeSystem" form:"includeSystem"`
}

type outboundToggleRequest struct {
	Enabled bool `json:"enabled" form:"enabled"`
}

func NewOutboundController(g *gin.RouterGroup) *OutboundController {
	a := &OutboundController{}
	a.initRouter(g)
	return a
}

func (a *OutboundController) initRouter(g *gin.RouterGroup) {
	g = g.Group("/outbound")

	g.POST("/list", a.getOutbounds)
	g.POST("/add", a.addOutbound)
	g.POST("/update/:id", a.updateOutbound)
	g.POST("/del/:id", a.delOutbound)
	g.POST("/toggle/:id", a.toggleOutbound)
}

func (a *OutboundController) getOutbounds(c *gin.Context) {
	req := &outboundListRequest{
		IncludeSystem: true,
	}
	if err := c.ShouldBind(req); err != nil {
		jsonMsg(c, "get outbound list", err)
		return
	}

	outbounds, err := a.outboundService.GetOutbounds(req.IncludeSystem)
	if err != nil {
		jsonMsg(c, "get outbound list", err)
		return
	}

	jsonObj(c, gin.H{
		"items": outbounds,
	}, nil)
}

func (a *OutboundController) addOutbound(c *gin.Context) {
	outbound := &model.Outbound{}
	if err := c.ShouldBind(outbound); err != nil {
		jsonMsg(c, "add outbound", err)
		return
	}

	applied, err := a.outboundService.AddOutbound(outbound)
	if err != nil {
		c.JSON(http.StatusOK, entity.Msg{Success: false, Msg: err.Error(), Obj: nil})
		return
	}

	c.JSON(http.StatusOK, entity.Msg{Success: true, Msg: "", Obj: gin.H{"item": outbound, "applied": applied}})
}

func (a *OutboundController) updateOutbound(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, "update outbound", err)
		return
	}

	outbound := &model.Outbound{Id: id}
	if err := c.ShouldBind(outbound); err != nil {
		jsonMsg(c, "update outbound", err)
		return
	}

	applied, err := a.outboundService.UpdateOutbound(outbound)
	if err != nil {
		c.JSON(http.StatusOK, entity.Msg{Success: false, Msg: err.Error(), Obj: nil})
		return
	}

	current, err := a.outboundService.GetOutbound(id)
	if err != nil {
		c.JSON(http.StatusOK, entity.Msg{Success: false, Msg: err.Error(), Obj: nil})
		return
	}

	c.JSON(http.StatusOK, entity.Msg{Success: true, Msg: "", Obj: gin.H{"item": current, "applied": applied}})
}

func (a *OutboundController) delOutbound(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, "delete outbound", err)
		return
	}

	applied, err := a.outboundService.DeleteOutbound(id)
	if err != nil {
		c.JSON(http.StatusOK, entity.Msg{Success: false, Msg: err.Error(), Obj: nil})
		return
	}

	c.JSON(http.StatusOK, entity.Msg{Success: true, Msg: "", Obj: gin.H{"id": id, "applied": applied}})
}

func (a *OutboundController) toggleOutbound(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, "toggle outbound", err)
		return
	}

	req := &outboundToggleRequest{}
	if err := c.ShouldBind(req); err != nil {
		jsonMsg(c, "toggle outbound", err)
		return
	}

	applied, err := a.outboundService.ToggleOutbound(id, req.Enabled)
	if err != nil {
		c.JSON(http.StatusOK, entity.Msg{Success: false, Msg: err.Error(), Obj: nil})
		return
	}

	current, err := a.outboundService.GetOutbound(id)
	if err != nil {
		c.JSON(http.StatusOK, entity.Msg{Success: false, Msg: err.Error(), Obj: nil})
		return
	}

	c.JSON(http.StatusOK, entity.Msg{Success: true, Msg: "", Obj: gin.H{"item": current, "applied": applied}})
}
