package handler

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/zaharinea/go-example/config"
	"github.com/zaharinea/go-example/pkg/repository"
	"github.com/zaharinea/go-example/pkg/service"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/context"
)

func SetupHandlers() *Handler {
	gin.SetMode(gin.ReleaseMode)

	c := config.NewTestingConfig()
	dbClient := repository.InitDbClient(c)
	repos := repository.NewRepository(dbClient.Database(c.MongoDbName))
	services := service.NewService(repos)
	handlers := NewHandler(c, services)

	return handlers
}

func TestListUsers(t *testing.T) {
	h := SetupHandlers()
	router := gin.New()
	h.InitRoutes(router)

	user1 := repository.User{
		ID:        primitive.NewObjectID(),
		Name:      "User1",
		CreatedAt: time.Date(2020, 11, 23, 23, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2020, 11, 23, 23, 0, 0, 0, time.UTC),
	}
	user2 := repository.User{
		ID:        primitive.NewObjectID(),
		Name:      "User2",
		CreatedAt: time.Date(2020, 11, 23, 23, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2020, 11, 23, 23, 0, 0, 0, time.UTC),
	}

	t.Run("Create", func(t *testing.T) {
		err := h.services.User.DeleteAll(context.Background())
		assert.NoError(t, err)

		w := performRequest(router, "POST", "/api/users", "{\"name\": \"user\"}")
		assert.Equal(t, http.StatusCreated, w.Code)

		response := ResponseUser{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "user", response.Name)
	})

	t.Run("Create Invalid request", func(t *testing.T) {
		err := h.services.User.DeleteAll(context.Background())
		assert.NoError(t, err)

		w := performRequest(router, "POST", "/api/users", "{}")
		assert.Equal(t, http.StatusBadRequest, w.Code)

		response := gin.H{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, gin.H{"message": "Key: 'RequestCreateUser.Name' Error:Field validation for 'Name' failed on the 'required' tag"}, response)
	})

	t.Run("Update", func(t *testing.T) {
		err := h.services.User.DeleteAll(context.Background())
		assert.NoError(t, err)
		err = h.services.User.Create(context.Background(), &user1)
		assert.NoError(t, err)

		w := performRequest(router, "PUT", "/api/users/"+user1.ID.Hex(), "{\"name\": \"user\"}")
		assert.Equal(t, http.StatusOK, w.Code)

		response := ResponseUser{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "user", response.Name)

		updatedUser, err := h.services.User.GetByID(context.Background(), user1.ID.Hex())
		assert.NoError(t, err)
		assert.Equal(t, "user", updatedUser.Name)

	})

	t.Run("Update Invalid request", func(t *testing.T) {
		err := h.services.User.DeleteAll(context.Background())
		assert.NoError(t, err)

		w := performRequest(router, "PUT", "/api/users/"+user1.ID.Hex(), "{}")
		assert.Equal(t, http.StatusBadRequest, w.Code)

		response := gin.H{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, gin.H{"message": "Key: 'RequestUpdateUser.Name' Error:Field validation for 'Name' failed on the 'required' tag"}, response)
	})

	t.Run("GetByID NotFound", func(t *testing.T) {
		err := h.services.User.DeleteAll(context.Background())
		assert.NoError(t, err)

		w := performRequest(router, "GET", "/api/users/5fbaeab741e97bef8525d6ab", "")
		assert.Equal(t, http.StatusNotFound, w.Code)

		response := errorResponse{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "mongo: no documents in result", response.Message)
	})

	t.Run("GetByID InvalidID", func(t *testing.T) {
		err := h.services.User.DeleteAll(context.Background())
		assert.NoError(t, err)

		w := performRequest(router, "GET", "/api/users/1", "")
		assert.Equal(t, http.StatusNotFound, w.Code)

		response := errorResponse{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "mongo: no documents in result", response.Message)
	})

	t.Run("GetByID", func(t *testing.T) {
		err := h.services.User.DeleteAll(context.Background())
		assert.NoError(t, err)
		err = h.services.User.Create(context.Background(), &user1)
		assert.NoError(t, err)

		w := performRequest(router, "GET", "/api/users/"+user1.ID.Hex(), "")
		assert.Equal(t, http.StatusOK, w.Code)

		response := ResponseUser{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, user1.ID.Hex(), response.ID)
		assert.Equal(t, user1.Name, response.Name)
	})

	t.Run("List Empty", func(t *testing.T) {
		err := h.services.User.DeleteAll(context.Background())
		assert.NoError(t, err)

		w := performRequest(router, "GET", "/api/users", "")
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "{\"items\":[]}", w.Body.String())
	})

	t.Run("List Ok", func(t *testing.T) {
		err := h.services.User.DeleteAll(context.Background())
		assert.NoError(t, err)
		err = h.services.User.Create(context.Background(), &user1)
		assert.NoError(t, err)
		err = h.services.User.Create(context.Background(), &user2)
		assert.NoError(t, err)

		w := performRequest(router, "GET", "/api/users", "")
		assert.Equal(t, http.StatusOK, w.Code)

		response := ResponseUsers{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(response.Items))
		assert.Equal(t, user1.ID.Hex(), response.Items[0].ID)
		assert.Equal(t, user1.Name, response.Items[0].Name)
		assert.Equal(t, user2.ID.Hex(), response.Items[1].ID)
		assert.Equal(t, user2.Name, response.Items[1].Name)
	})

	t.Run("List Ok with limit and offset", func(t *testing.T) {
		err := h.services.User.DeleteAll(context.Background())
		assert.NoError(t, err)
		err = h.services.User.Create(context.Background(), &user1)
		assert.NoError(t, err)
		err = h.services.User.Create(context.Background(), &user2)
		assert.NoError(t, err)

		w := performRequest(router, "GET", "/api/users?limit=1&offset=1", "")
		assert.Equal(t, http.StatusOK, w.Code)

		response := ResponseUsers{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(response.Items))
		assert.Equal(t, user2.ID.Hex(), response.Items[0].ID)
		assert.Equal(t, user2.Name, response.Items[0].Name)
	})

	t.Run("Delete Ok", func(t *testing.T) {
		err := h.services.User.DeleteAll(context.Background())
		assert.NoError(t, err)
		err = h.services.User.Create(context.Background(), &user1)
		assert.NoError(t, err)

		w := performRequest(router, "DELETE", "/api/users/"+user1.ID.Hex(), "")
		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Equal(t, "", w.Body.String())

		_, err = h.services.User.GetByID(context.Background(), user1.ID.Hex())
		assert.Error(t, err, mongo.ErrNoDocuments)
	})

	t.Run("Delete NotFound", func(t *testing.T) {
		err := h.services.User.DeleteAll(context.Background())
		assert.NoError(t, err)

		w := performRequest(router, "DELETE", "/api/users/5fbaeab741e97bef8525d6ab", "")
		assert.Equal(t, http.StatusNoContent, w.Code) // FIXME should 404
		assert.Equal(t, "", w.Body.String())
	})

	t.Run("Delete InvalidID", func(t *testing.T) {
		err := h.services.User.DeleteAll(context.Background())
		assert.NoError(t, err)

		w := performRequest(router, "DELETE", "/api/users/1", "")
		assert.Equal(t, http.StatusNotFound, w.Code)

		response := errorResponse{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "mongo: no documents in result", response.Message)
	})

}
