package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"x-ui/database/model"
	"x-ui/web/entity"
	"x-ui/web/service"
)

type RoutingController struct {
	routingService service.RoutingService
}

type routingListRequest struct {
	IncludeSystem bool `json:"includeSystem" form:"includeSystem"`
}

type routingRuleUpsertRequest struct {
	Type        string `json:"type" form:"type"`
	Domain      string `json:"domain" form:"domain"`
	IP          string `json:"ip" form:"ip"`
	Port        string `json:"port" form:"port"`
	Protocol    string `json:"protocol" form:"protocol"`
	Network     string `json:"network" form:"network"`
	Source      string `json:"source" form:"source"`
	InboundTag  string `json:"inboundTag" form:"inboundTag"`
	OutboundTag string `json:"outboundTag" form:"outboundTag"`
	Enabled     *bool  `json:"enabled" form:"enabled"`
	Remark      string `json:"remark" form:"remark"`
	Sort        int    `json:"sort" form:"sort"`
}

type routingToggleRequest struct {
	Enabled *bool `json:"enabled" form:"enabled"`
}

type routingReorderRequest struct {
	Items []service.RoutingReorderItem `json:"items" form:"items"`
}

type routingAITemplateRequest struct {
	OutboundTag  string `json:"outboundTag" form:"outboundTag"`
	SortBase     int    `json:"sortBase" form:"sortBase"`
	RemarkPrefix string `json:"remarkPrefix" form:"remarkPrefix"`
}

func NewRoutingController(g *gin.RouterGroup) *RoutingController {
	a := &RoutingController{}
	a.initRouter(g)
	return a
}

func (a *RoutingController) initRouter(g *gin.RouterGroup) {
	g = g.Group("/routing")

	g.POST("/list", a.getRules)
	g.POST("/add", a.addRule)
	g.POST("/update/:id", a.updateRule)
	g.POST("/del/:id", a.delRule)
	g.POST("/toggle/:id", a.toggleRule)
	g.POST("/reorder", a.reorderRules)
	g.POST("/applyAiTemplate", a.applyAITemplate)
}

func (a *RoutingController) getRules(c *gin.Context) {
	req := &routingListRequest{
		IncludeSystem: true,
	}
	if err := c.ShouldBind(req); err != nil {
		jsonMsg(c, "get routing list", err)
		return
	}

	rules, err := a.routingService.GetRules(req.IncludeSystem)
	if err != nil {
		jsonMsg(c, "get routing list", err)
		return
	}

	jsonObj(c, gin.H{
		"items": rules,
	}, nil)
}

func (r *routingRuleUpsertRequest) toModel(id int, defaultEnabled bool) *model.RoutingRule {
	enabled := defaultEnabled
	if r.Enabled != nil {
		enabled = *r.Enabled
	}
	return &model.RoutingRule{
		Id:          id,
		Type:        r.Type,
		Domain:      r.Domain,
		IP:          r.IP,
		Port:        r.Port,
		Protocol:    r.Protocol,
		Network:     r.Network,
		Source:      r.Source,
		InboundTag:  r.InboundTag,
		OutboundTag: r.OutboundTag,
		Enabled:     enabled,
		Remark:      r.Remark,
		Sort:        r.Sort,
	}
}

func (a *RoutingController) addRule(c *gin.Context) {
	req := &routingRuleUpsertRequest{}
	if err := c.ShouldBind(req); err != nil {
		jsonMsg(c, "add routing rule", err)
		return
	}
	rule := req.toModel(0, true)
	applied, err := a.routingService.AddRule(rule)
	if err != nil {
		c.JSON(http.StatusOK, entity.Msg{Success: false, Msg: err.Error(), Obj: nil})
		return
	}
	c.JSON(http.StatusOK, entity.Msg{Success: true, Msg: "", Obj: gin.H{"item": rule, "applied": applied}})
}

func (a *RoutingController) updateRule(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, "update routing rule", err)
		return
	}
	req := &routingRuleUpsertRequest{}
	if err := c.ShouldBind(req); err != nil {
		jsonMsg(c, "update routing rule", err)
		return
	}
	current, err := a.routingService.GetRule(id)
	if err != nil {
		c.JSON(http.StatusOK, entity.Msg{Success: false, Msg: err.Error(), Obj: nil})
		return
	}
	rule := req.toModel(id, current.Enabled)
	applied, err := a.routingService.UpdateRule(rule)
	if err != nil {
		c.JSON(http.StatusOK, entity.Msg{Success: false, Msg: err.Error(), Obj: nil})
		return
	}
	current, err = a.routingService.GetRule(id)
	if err != nil {
		c.JSON(http.StatusOK, entity.Msg{Success: false, Msg: err.Error(), Obj: nil})
		return
	}
	c.JSON(http.StatusOK, entity.Msg{Success: true, Msg: "", Obj: gin.H{"item": current, "applied": applied}})
}

func (a *RoutingController) delRule(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, "delete routing rule", err)
		return
	}
	applied, err := a.routingService.DeleteRule(id)
	if err != nil {
		c.JSON(http.StatusOK, entity.Msg{Success: false, Msg: err.Error(), Obj: nil})
		return
	}
	c.JSON(http.StatusOK, entity.Msg{Success: true, Msg: "", Obj: gin.H{"id": id, "applied": applied}})
}

func (a *RoutingController) toggleRule(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, "toggle routing rule", err)
		return
	}
	req := &routingToggleRequest{}
	if err := c.ShouldBind(req); err != nil {
		jsonMsg(c, "toggle routing rule", err)
		return
	}
	if req.Enabled == nil {
		c.JSON(http.StatusOK, entity.Msg{Success: false, Msg: "ROUTING_RULE_ENABLED_REQUIRED: enabled is required", Obj: nil})
		return
	}
	applied, err := a.routingService.ToggleRule(id, *req.Enabled)
	if err != nil {
		c.JSON(http.StatusOK, entity.Msg{Success: false, Msg: err.Error(), Obj: nil})
		return
	}
	current, err := a.routingService.GetRule(id)
	if err != nil {
		c.JSON(http.StatusOK, entity.Msg{Success: false, Msg: err.Error(), Obj: nil})
		return
	}
	c.JSON(http.StatusOK, entity.Msg{Success: true, Msg: "", Obj: gin.H{"item": current, "applied": applied}})
}

func (a *RoutingController) reorderRules(c *gin.Context) {
	req := &routingReorderRequest{}
	if err := c.ShouldBind(req); err != nil {
		jsonMsg(c, "reorder routing rules", err)
		return
	}
	applied, err := a.routingService.ReorderRules(req.Items)
	if err != nil {
		c.JSON(http.StatusOK, entity.Msg{Success: false, Msg: err.Error(), Obj: nil})
		return
	}
	c.JSON(http.StatusOK, entity.Msg{Success: true, Msg: "", Obj: gin.H{"items": req.Items, "applied": applied}})
}

func (a *RoutingController) applyAITemplate(c *gin.Context) {
	req := &routingAITemplateRequest{}
	if err := c.ShouldBind(req); err != nil {
		jsonMsg(c, "apply ai template", err)
		return
	}

	created, skipped, applied, err := a.routingService.ApplyAITemplate(req.OutboundTag, req.SortBase, req.RemarkPrefix)
	if err != nil {
		c.JSON(http.StatusOK, entity.Msg{Success: false, Msg: err.Error(), Obj: nil})
		return
	}

	c.JSON(http.StatusOK, entity.Msg{
		Success: true,
		Msg:     "apply ai template success",
		Obj: gin.H{
			"created":     created,
			"skipped":     skipped,
			"outboundTag": req.OutboundTag,
			"applied":     applied,
		},
	})
}
