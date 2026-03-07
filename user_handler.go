package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/yourorg/users-api/internal/middleware"
	"github.com/yourorg/users-api/internal/models"
	"github.com/yourorg/users-api/internal/service"
)

// UserHandler handles HTTP requests for user operations.
type UserHandler struct {
	svc *service.UserService
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// RegisterRoutes mounts all user routes under the given router group.
func (h *UserHandler) RegisterRoutes(rg *gin.RouterGroup) {
	users := rg.Group("/users")
	{
		users.GET("", h.List)
		users.POST("", h.Create)
		users.GET("/:id", h.GetByID)
		users.PATCH("/:id", h.Update)
		users.DELETE("/:id", h.Delete)
	}
}

// List godoc
//
//	@Summary		List users
//	@Description	Returns a paginated list of users scoped to the authenticated tenant.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			page	query		int	false	"Page number (default 1)"	minimum(1)
//	@Param			limit	query		int	false	"Items per page (default 20, max 100)"	minimum(1)	maximum(100)
//	@Success		200		{object}	models.ListUsersResponse
//	@Failure		401		{object}	models.ErrorResponse
//	@Failure		500		{object}	models.ErrorResponse
//	@Security		BearerAuth
//	@Router			/v1/users [get]
func (h *UserHandler) List(c *gin.Context) {
	tenantID := middleware.MustTenantID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	users, total, err := h.svc.ListUsers(c.Request.Context(), tenantID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to list users", Code: "INTERNAL"})
		return
	}

	resp := models.ListUsersResponse{
		Total: total,
		Page:  page,
		Limit: limit,
	}
	for _, u := range users {
		resp.Data = append(resp.Data, u.ToResponse())
	}
	if resp.Data == nil {
		resp.Data = []models.UserResponse{}
	}

	c.JSON(http.StatusOK, resp)
}

// GetByID godoc
//
//	@Summary		Get user by ID
//	@Description	Returns a single user by UUID, scoped to the authenticated tenant.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"User UUID"	format(uuid)
//	@Success		200	{object}	models.UserResponse
//	@Failure		400	{object}	models.ErrorResponse
//	@Failure		401	{object}	models.ErrorResponse
//	@Failure		404	{object}	models.ErrorResponse
//	@Security		BearerAuth
//	@Router			/v1/users/{id} [get]
func (h *UserHandler) GetByID(c *gin.Context) {
	tenantID := middleware.MustTenantID(c)

	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid user id", Code: "INVALID_ID"})
		return
	}

	user, err := h.svc.GetUser(c.Request.Context(), tenantID, userID)
	if errors.Is(err, service.ErrUserNotFound) {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "user not found", Code: "NOT_FOUND"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to get user", Code: "INTERNAL"})
		return
	}

	c.JSON(http.StatusOK, user.ToResponse())
}

// Create godoc
//
//	@Summary		Create user
//	@Description	Creates a new user scoped to the authenticated tenant.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			body	body		models.CreateUserRequest	true	"User payload"
//	@Success		201		{object}	models.UserResponse
//	@Failure		400		{object}	models.ErrorResponse
//	@Failure		401		{object}	models.ErrorResponse
//	@Failure		409		{object}	models.ErrorResponse	"Email already in use"
//	@Failure		500		{object}	models.ErrorResponse
//	@Security		BearerAuth
//	@Router			/v1/users [post]
func (h *UserHandler) Create(c *gin.Context) {
	tenantID := middleware.MustTenantID(c)

	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error(), Code: "VALIDATION"})
		return
	}

	user, err := h.svc.CreateUser(c.Request.Context(), tenantID, &req)
	if errors.Is(err, service.ErrEmailTaken) {
		c.JSON(http.StatusConflict, models.ErrorResponse{Error: "email already in use", Code: "EMAIL_TAKEN"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to create user", Code: "INTERNAL"})
		return
	}

	c.JSON(http.StatusCreated, user.ToResponse())
}

// Update godoc
//
//	@Summary		Update user
//	@Description	Applies partial updates to a user. Only provided fields are changed.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string						true	"User UUID"	format(uuid)
//	@Param			body	body		models.UpdateUserRequest	true	"Fields to update"
//	@Success		200		{object}	models.UserResponse
//	@Failure		400		{object}	models.ErrorResponse
//	@Failure		401		{object}	models.ErrorResponse
//	@Failure		404		{object}	models.ErrorResponse
//	@Failure		500		{object}	models.ErrorResponse
//	@Security		BearerAuth
//	@Router			/v1/users/{id} [patch]
func (h *UserHandler) Update(c *gin.Context) {
	tenantID := middleware.MustTenantID(c)

	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid user id", Code: "INVALID_ID"})
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error(), Code: "VALIDATION"})
		return
	}

	user, err := h.svc.UpdateUser(c.Request.Context(), tenantID, userID, &req)
	if errors.Is(err, service.ErrUserNotFound) {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "user not found", Code: "NOT_FOUND"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to update user", Code: "INTERNAL"})
		return
	}

	c.JSON(http.StatusOK, user.ToResponse())
}

// Delete godoc
//
//	@Summary		Delete user
//	@Description	Soft-deletes a user. The record is retained in the database with a deleted_at timestamp.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id	path	string	true	"User UUID"	format(uuid)
//	@Success		204	"No Content"
//	@Failure		400	{object}	models.ErrorResponse
//	@Failure		401	{object}	models.ErrorResponse
//	@Failure		404	{object}	models.ErrorResponse
//	@Failure		500	{object}	models.ErrorResponse
//	@Security		BearerAuth
//	@Router			/v1/users/{id} [delete]
func (h *UserHandler) Delete(c *gin.Context) {
	tenantID := middleware.MustTenantID(c)

	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid user id", Code: "INVALID_ID"})
		return
	}

	if err := h.svc.DeleteUser(c.Request.Context(), tenantID, userID); err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "user not found", Code: "NOT_FOUND"})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to delete user", Code: "INTERNAL"})
		return
	}

	c.Status(http.StatusNoContent)
}
