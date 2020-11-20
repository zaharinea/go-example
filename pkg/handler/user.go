package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zaharinea/go-example/pkg/repository"
)

// RequestCreateUser struct
type RequestCreateUser struct {
	Name string `json:"name" binding:"required"`
}

// RequestUpdateUser struct
type RequestUpdateUser struct {
	Name string `json:"name" binding:"required"`
}

// ResponseUser struct
type ResponseUser struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ResponseUsers struct
type ResponseUsers struct {
	Items []*ResponseUser `json:"items"`
}

// CreateUser handler
// @Summary Create user
// @Tags users
// @Accept  json
// @Produce  json
// @Param user body RequestCreateUser true "Add user"
// @Success 201 {object} ResponseUser
// @Router /api/users [post]
func (h *Handler) CreateUser(c *gin.Context) {
	var requestData RequestCreateUser

	if err := c.BindJSON(&requestData); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	newUser := repository.User{Name: requestData.Name}
	_, err := h.services.User.Create(c, &newUser)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusCreated, ResponseUser{
		ID:        newUser.ID.Hex(),
		Name:      newUser.Name,
		CreatedAt: newUser.CreatedAt,
		UpdatedAt: newUser.UpdatedAt,
	})
}

// ListUsers handler
// @Summary List users
// @Description get users
// @Tags users
// @Accept  json
// @Produce  json
// @Param limit query int false "limit" mininum(1) maxinum(1000) default(25)
// @Param offset query int false "offset" mininum(0) default(0)
// @Success 200 {object} ResponseUsers
// @Router /api/users [get]
func (h *Handler) ListUsers(c *gin.Context) {
	limit := Limit(c.Query("limit"), h.config.PageSize)
	offset := Offset(c.Query("offset"), 0)

	users, err := h.services.User.List(c, limit, offset)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	items := make([]*ResponseUser, len(users))
	for idx, user := range users {
		items[idx] = &ResponseUser{
			ID:        user.ID.Hex(),
			Name:      user.Name,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, ResponseUsers{Items: items})
}

// GetUserByID handler
// @Summary Get user by ID
// @Description get user by ID
// @Tags users
// @Accept  json
// @Produce  json
// @Param  id path string true "User ID"
// @Success 200 {object} ResponseUser
// @Router /api/users/{id} [get]
func (h *Handler) GetUserByID(c *gin.Context) {
	userID := c.Param("id")
	user, err := h.services.User.GetByID(c, userID)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, ResponseUser{
		ID:        user.ID.Hex(),
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	})
}

// UpdateUser handler
// @Summary Update user
// @Description Update by json user
// @Tags users
// @Accept  json
// @Produce  json
// @Param  id path string true "User ID"
// @Param user body RequestUpdateUser true "Update user"
// @Success 200 {object} ResponseUser
// @Router /api/users/{id} [put]
func (h *Handler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	var requestData RequestUpdateUser

	if err := c.BindJSON(&requestData); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	updateUser := repository.UpdateUser{Name: requestData.Name}
	user, err := h.services.User.UpdateAndReturn(c, userID, updateUser)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, ResponseUser{
		ID:        user.ID.Hex(),
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	})
}

// DeleteUserByID handler
// @Summary Delete user
// @Description Delete by user ID
// @Tags users
// @Accept  json
// @Produce  json
// @Param  id path string true "User ID"
// @Success 204 {object} emptyResponse
// @Router /api/users/{id} [delete]
func (h *Handler) DeleteUserByID(c *gin.Context) {
	userID := c.Param("id")
	err := h.services.User.DeleteByID(c, userID)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusNoContent, gin.H{})
}
