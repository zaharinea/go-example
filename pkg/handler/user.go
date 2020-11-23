package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zaharinea/go-example/pkg/repository"
	"go.mongodb.org/mongo-driver/mongo"
)

// RequestCreateUser struct
type RequestCreateUser struct {
	Name string `json:"name" binding:"required"`
}

// RequestUpdateUser struct
type RequestUpdateUser struct {
	Name string `json:"name" binding:"required"`
}

// RequestListUsers struct
type RequestListUsers struct {
	Limit  int64 `form:"limit"`
	Offset int64 `form:"offset"`
}

// RequestGetUser struct
type RequestGetUser struct {
	ID string `uri:"id" binding:"required"`
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

func newResponseUser(user repository.User) *ResponseUser {
	return &ResponseUser{
		ID:        user.ID.Hex(),
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func newResponseUsers(users []repository.User) ResponseUsers {
	items := make([]*ResponseUser, len(users))
	for idx, user := range users {
		items[idx] = newResponseUser(user)
	}
	return ResponseUsers{Items: items}
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
	var req RequestCreateUser
	if err := c.ShouldBindJSON(&req); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	newUser := repository.User{Name: req.Name}
	_, err := h.services.User.Create(c, &newUser)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusCreated, newResponseUser(newUser))
}

// ListUsers handler
// @Summary List users
// @Description get users
// @Tags users
// @Accept  json
// @Produce  json
// @Param limit query int false "limit" mininum(1) maxinum(100) default(25)
// @Param offset query int false "offset" mininum(0) default(0)
// @Success 200 {object} ResponseUsers
// @Router /api/users [get]
func (h *Handler) ListUsers(c *gin.Context) {
	var req RequestListUsers
	if err := c.ShouldBindQuery(&req); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = h.config.PageSize
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	users, err := h.services.User.List(c, req.Limit, req.Offset)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, newResponseUsers(users))
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
	var req RequestGetUser
	if err := c.ShouldBindUri(&req); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.services.User.GetByID(c, req.ID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			newErrorResponse(c, http.StatusNotFound, err.Error())
			return
		}

		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, newResponseUser(user))
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
	var reqURI RequestGetUser
	if err := c.ShouldBindUri(&reqURI); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	var reqData RequestUpdateUser
	if err := c.ShouldBindJSON(&reqData); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	updateUser := repository.UpdateUser{Name: reqData.Name}
	updatedUser, err := h.services.User.UpdateAndReturn(c, reqURI.ID, updateUser)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			newErrorResponse(c, http.StatusNotFound, err.Error())
			return
		}

		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, newResponseUser(updatedUser))
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
		if err == mongo.ErrNoDocuments {
			newErrorResponse(c, http.StatusNotFound, err.Error())
			return
		}

		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusNoContent, gin.H{})
}
