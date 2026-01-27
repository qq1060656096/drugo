package tpl

// ModuleApi templates for generating API structure within an existing module.

const ModuleApiApiTpl = `package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"{{.ModPath}}/internal/{{.ModuleName}}/biz"
	"{{.ModPath}}/internal/{{.ModuleName}}/data"
	"{{.ModPath}}/internal/{{.ModuleName}}/service"
	"github.com/qq1060656096/drugo/pkg/router"
)

func init() {
	// 自动注册{{.NameTitle}}路由
	router.Default().Register(func(r *gin.Engine) {
		api := New{{.NameTitle}}Handler()
		api.RegisterRoutes(r)
	})
}

// {{.NameTitle}}Handler {{.Name}} API 处理器
type {{.NameTitle}}Handler struct {
	svc *service.{{.NameTitle}}Service
}

// New{{.NameTitle}}Handler 创建 {{.NameTitle}}Handler 实例
func New{{.NameTitle}}Handler() *{{.NameTitle}}Handler {
	// 依赖注入: data -> biz -> service
	repo := data.New{{.NameTitle}}Repo()
	uc := biz.New{{.NameTitle}}Usecase(repo)
	svc := service.New{{.NameTitle}}Service(uc)
	return &{{.NameTitle}}Handler{svc: svc}
}

// RegisterRoutes 注册{{.Name}}相关路由
func (h *{{.NameTitle}}Handler) RegisterRoutes(r gin.IRouter) {
	group := r.Group("/{{.ModuleName}}/{{.Name}}")
	{
		group.POST("", h.Create)
		group.GET("", h.List)
		group.GET("/:id", h.Get)
		group.PUT("/:id", h.Update)
		group.DELETE("/:id", h.Delete)
	}
}

// Create 创建{{.Name}}
// POST /{{.Name}}
func (h *{{.NameTitle}}Handler) Create(c *gin.Context) {
	var req service.Create{{.NameTitle}}Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "invalid request: " + err.Error(),
		})
		return
	}

	resp, err := h.svc.Create(c.Request.Context(), &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "success",
		"data":    resp,
	})
}

// Get 获取{{.Name}}详情
// GET /{{.Name}}/:id
func (h *{{.NameTitle}}Handler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "invalid id",
		})
		return
	}

	resp, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    resp,
	})
}

// Update 更新{{.Name}}
// PUT /{{.Name}}/:id
func (h *{{.NameTitle}}Handler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "invalid id",
		})
		return
	}

	var req service.Update{{.NameTitle}}Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "invalid request: " + err.Error(),
		})
		return
	}

	resp, err := h.svc.Update(c.Request.Context(), id, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    resp,
	})
}

// Delete 删除{{.Name}}
// DELETE /{{.Name}}/:id
func (h *{{.NameTitle}}Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "invalid id",
		})
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// List 获取{{.Name}}列表
// GET /{{.Name}}
func (h *{{.NameTitle}}Handler) List(c *gin.Context) {
	var req service.List{{.NameTitle}}Request
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "invalid request: " + err.Error(),
		})
		return
	}

	resp, err := h.svc.List(c.Request.Context(), &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    resp,
	})
}

// handleError 统一错误处理
// 注意：作为方法挂载在 Handler 上以避免与其他 Handler 的辅助函数冲突
func (h *{{.NameTitle}}Handler) handleError(c *gin.Context, err error) {
	if service.Is{{.NameTitle}}NotFound(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "not found",
		})
		return
	}
	if service.Is{{.NameTitle}}InvalidParams(err) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "invalid params",
		})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{
		"code":    500,
		"message": "internal server error",
	})
}
`

const ModuleApiBizTpl = `package biz

import (
	"context"
	"errors"
)

// 业务错误定义
var (
	Err{{.NameTitle}}NotFound  = errors.New("{{.Name}} not found")
	Err{{.NameTitle}}InvalidParams = errors.New("invalid params")
)

// {{.NameTitle}} {{.Name}}实体
type {{.NameTitle}} struct {
	ID   int64  ` + "`json:\"id\"`" + `
	Name string ` + "`json:\"name\"`" + `
	// TODO: 添加更多字段
}

// {{.NameTitle}}Repo {{.Name}}数据仓库接口
type {{.NameTitle}}Repo interface {
	Create(ctx context.Context, entity *{{.NameTitle}}) (*{{.NameTitle}}, error)
	Get(ctx context.Context, id int64) (*{{.NameTitle}}, error)
	Update(ctx context.Context, entity *{{.NameTitle}}) (*{{.NameTitle}}, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, page, pageSize int) ([]*{{.NameTitle}}, int64, error)
}

// {{.NameTitle}}Usecase {{.Name}}业务逻辑
type {{.NameTitle}}Usecase struct {
	repo {{.NameTitle}}Repo
}

// New{{.NameTitle}}Usecase 创建 {{.NameTitle}}Usecase 实例
func New{{.NameTitle}}Usecase(repo {{.NameTitle}}Repo) *{{.NameTitle}}Usecase {
	return &{{.NameTitle}}Usecase{repo: repo}
}

// Create 创建{{.Name}}
func (uc *{{.NameTitle}}Usecase) Create(ctx context.Context, name string) (*{{.NameTitle}}, error) {
	if name == "" {
		return nil, Err{{.NameTitle}}InvalidParams
	}
	entity := &{{.NameTitle}}{
		Name: name,
	}
	return uc.repo.Create(ctx, entity)
}

// Get 获取{{.Name}}详情
func (uc *{{.NameTitle}}Usecase) Get(ctx context.Context, id int64) (*{{.NameTitle}}, error) {
	if id <= 0 {
		return nil, Err{{.NameTitle}}InvalidParams
	}
	return uc.repo.Get(ctx, id)
}

// Update 更新{{.Name}}
func (uc *{{.NameTitle}}Usecase) Update(ctx context.Context, id int64, name string) (*{{.NameTitle}}, error) {
	if id <= 0 {
		return nil, Err{{.NameTitle}}InvalidParams
	}
	entity := &{{.NameTitle}}{
		ID:   id,
		Name: name,
	}
	return uc.repo.Update(ctx, entity)
}

// Delete 删除{{.Name}}
func (uc *{{.NameTitle}}Usecase) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return Err{{.NameTitle}}InvalidParams
	}
	return uc.repo.Delete(ctx, id)
}

// List 获取{{.Name}}列表
func (uc *{{.NameTitle}}Usecase) List(ctx context.Context, page, pageSize int) ([]*{{.NameTitle}}, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return uc.repo.List(ctx, page, pageSize)
}
`

const ModuleApiDataTpl = `package data

import (
	"context"
	"sync"

	"{{.ModPath}}/internal/{{.ModuleName}}/biz"
)

// {{.Name}}Repo 实现 biz.{{.NameTitle}}Repo 接口，使用内存存储
type {{.Name}}Repo struct {
	mu    sync.RWMutex
	items map[int64]*biz.{{.NameTitle}}
	maxID int64
}

// New{{.NameTitle}}Repo 创建 {{.NameTitle}}Repo 实例
func New{{.NameTitle}}Repo() biz.{{.NameTitle}}Repo {
	return &{{.Name}}Repo{
		items: make(map[int64]*biz.{{.NameTitle}}),
		maxID: 0,
	}
}

// Create 创建{{.Name}}
func (r *{{.Name}}Repo) Create(ctx context.Context, entity *biz.{{.NameTitle}}) (*biz.{{.NameTitle}}, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.maxID++
	entity.ID = r.maxID
	r.items[entity.ID] = entity
	return entity, nil
}

// Get 根据 ID 获取{{.Name}}
func (r *{{.Name}}Repo) Get(ctx context.Context, id int64) (*biz.{{.NameTitle}}, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entity, ok := r.items[id]
	if !ok {
		return nil, biz.Err{{.NameTitle}}NotFound
	}
	return entity, nil
}

// Update 更新{{.Name}}
func (r *{{.Name}}Repo) Update(ctx context.Context, entity *biz.{{.NameTitle}}) (*biz.{{.NameTitle}}, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.items[entity.ID]; !ok {
		return nil, biz.Err{{.NameTitle}}NotFound
	}
	r.items[entity.ID] = entity
	return entity, nil
}

// Delete 删除{{.Name}}
func (r *{{.Name}}Repo) Delete(ctx context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.items[id]; !ok {
		return biz.Err{{.NameTitle}}NotFound
	}
	delete(r.items, id)
	return nil
}

// List 获取{{.Name}}列表
func (r *{{.Name}}Repo) List(ctx context.Context, page, pageSize int) ([]*biz.{{.NameTitle}}, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	total := int64(len(r.items))
	items := make([]*biz.{{.NameTitle}}, 0, len(r.items))
	for _, item := range r.items {
		items = append(items, item)
	}

	// 简单分页
	start := (page - 1) * pageSize
	if start >= len(items) {
		return []*biz.{{.NameTitle}}{}, total, nil
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}
	return items[start:end], total, nil
}
`

const ModuleApiServiceTpl = `package service

import (
	"context"
	"errors"

	"{{.ModPath}}/internal/{{.ModuleName}}/biz"
)

// Create{{.NameTitle}}Request 创建{{.Name}}请求
type Create{{.NameTitle}}Request struct {
	Name string ` + "`json:\"name\" binding:\"required\"`" + `
}

// Update{{.NameTitle}}Request 更新{{.Name}}请求
type Update{{.NameTitle}}Request struct {
	Name string ` + "`json:\"name\"`" + `
}

// List{{.NameTitle}}Request {{.Name}}列表请求
type List{{.NameTitle}}Request struct {
	Page     int ` + "`form:\"page\"`" + `
	PageSize int ` + "`form:\"page_size\"`" + `
}

// {{.NameTitle}}Response {{.Name}}响应
type {{.NameTitle}}Response struct {
	ID   int64  ` + "`json:\"id\"`" + `
	Name string ` + "`json:\"name\"`" + `
}

// List{{.NameTitle}}Response {{.Name}}列表响应
type List{{.NameTitle}}Response struct {
	Total int64                ` + "`json:\"total\"`" + `
	List  []*{{.NameTitle}}Response ` + "`json:\"list\"`" + `
}

// {{.NameTitle}}Service {{.Name}}服务
type {{.NameTitle}}Service struct {
	uc *biz.{{.NameTitle}}Usecase
}

// New{{.NameTitle}}Service 创建 {{.NameTitle}}Service 实例
func New{{.NameTitle}}Service(uc *biz.{{.NameTitle}}Usecase) *{{.NameTitle}}Service {
	return &{{.NameTitle}}Service{uc: uc}
}

// Create 创建{{.Name}}
func (s *{{.NameTitle}}Service) Create(ctx context.Context, req *Create{{.NameTitle}}Request) (*{{.NameTitle}}Response, error) {
	entity, err := s.uc.Create(ctx, req.Name)
	if err != nil {
		return nil, err
	}
	return to{{.NameTitle}}Response(entity), nil
}

// Get 获取{{.Name}}
func (s *{{.NameTitle}}Service) Get(ctx context.Context, id int64) (*{{.NameTitle}}Response, error) {
	entity, err := s.uc.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return to{{.NameTitle}}Response(entity), nil
}

// Update 更新{{.Name}}
func (s *{{.NameTitle}}Service) Update(ctx context.Context, id int64, req *Update{{.NameTitle}}Request) (*{{.NameTitle}}Response, error) {
	entity, err := s.uc.Update(ctx, id, req.Name)
	if err != nil {
		return nil, err
	}
	return to{{.NameTitle}}Response(entity), nil
}

// Delete 删除{{.Name}}
func (s *{{.NameTitle}}Service) Delete(ctx context.Context, id int64) error {
	return s.uc.Delete(ctx, id)
}

// List 获取{{.Name}}列表
func (s *{{.NameTitle}}Service) List(ctx context.Context, req *List{{.NameTitle}}Request) (*List{{.NameTitle}}Response, error) {
	items, total, err := s.uc.List(ctx, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	list := make([]*{{.NameTitle}}Response, 0, len(items))
	for _, item := range items {
		list = append(list, to{{.NameTitle}}Response(item))
	}
	return &List{{.NameTitle}}Response{
		Total: total,
		List:  list,
	}, nil
}

// to{{.NameTitle}}Response 转换为响应结构
func to{{.NameTitle}}Response(entity *biz.{{.NameTitle}}) *{{.NameTitle}}Response {
	return &{{.NameTitle}}Response{
		ID:   entity.ID,
		Name: entity.Name,
	}
}

// Is{{.NameTitle}}NotFound 判断是否为未找到错误
func Is{{.NameTitle}}NotFound(err error) bool {
	return errors.Is(err, biz.Err{{.NameTitle}}NotFound)
}

// Is{{.NameTitle}}InvalidParams 判断是否为参数错误
func Is{{.NameTitle}}InvalidParams(err error) bool {
	return errors.Is(err, biz.Err{{.NameTitle}}InvalidParams)
}
`
