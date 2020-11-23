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
	db := dbClient.Database(c.MongoDbName)
	repository.ApplyDbMigrations(c, dbClient)
	repos := repository.NewRepository(db)
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

	t.Run("GetByID NotFound", func(t *testing.T) {
		h.services.User.DeleteAll(context.Background())

		w := performRequest(router, "GET", "/api/users/5fbaeab741e97bef8525d6ab")
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("GetByID InvalidID", func(t *testing.T) {
		h.services.User.DeleteAll(context.Background())

		w := performRequest(router, "GET", "/api/users/1")
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("GetByID", func(t *testing.T) {
		h.services.User.DeleteAll(context.Background())
		h.services.User.Create(context.Background(), &user1)

		w := performRequest(router, "GET", "/api/users/"+user1.ID.Hex())
		assert.Equal(t, http.StatusOK, w.Code)

		response := ResponseUser{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, response.ID, user1.ID.Hex())
		assert.Equal(t, response.Name, user1.Name)
	})

	t.Run("List Empty", func(t *testing.T) {
		h.services.User.DeleteAll(context.Background())

		w := performRequest(router, "GET", "/api/users")
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, w.Body.String(), "{\"items\":[]}")
	})

	t.Run("List Ok", func(t *testing.T) {
		h.services.User.DeleteAll(context.Background())
		h.services.User.Create(context.Background(), &user1)

		w := performRequest(router, "GET", "/api/users")
		assert.Equal(t, http.StatusOK, w.Code)

		response := ResponseUsers{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, len(response.Items), 1)
		assert.Equal(t, response.Items[0].ID, user1.ID.Hex())
		assert.Equal(t, response.Items[0].Name, user1.Name)
	})

	t.Run("Delete Ok", func(t *testing.T) {
		h.services.User.DeleteAll(context.Background())
		h.services.User.Create(context.Background(), &user1)

		w := performRequest(router, "DELETE", "/api/users/"+user1.ID.Hex())
		assert.Equal(t, http.StatusNoContent, w.Code)

		_, err := h.services.User.GetByID(context.Background(), user1.ID.Hex())
		assert.Error(t, err, mongo.ErrNoDocuments)
	})

	t.Run("Delete NotFound", func(t *testing.T) {
		h.services.User.DeleteAll(context.Background())

		w := performRequest(router, "DELETE", "/api/users/5fbaeab741e97bef8525d6ab")
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("Delete InvalidID", func(t *testing.T) {
		h.services.User.DeleteAll(context.Background())

		w := performRequest(router, "DELETE", "/api/users/1")
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

}
