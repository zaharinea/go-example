package handler

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/zaharinea/go-example/config"
	"github.com/zaharinea/go-example/pkg/repository"
	"github.com/zaharinea/go-example/pkg/service"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/context"
)

type UsersSuite struct {
	suite.Suite
	ctx      context.Context
	config   *config.Config
	db       *mongo.Database
	router   *gin.Engine
	repos    *repository.Repository
	services *service.Service
	handlers *Handler
	user1    repository.User
	user2    repository.User
}

func (s *UsersSuite) SetupSuite() {
	gin.SetMode(gin.ReleaseMode)
	s.ctx = context.Background()
	s.config = config.NewTestingConfig()
	dbClient := repository.InitDbClient(s.config)
	s.db = dbClient.Database(s.config.MongoDbName)
	s.repos = repository.NewRepository(s.db)
	s.services = service.NewService(s.repos)
	s.handlers = NewHandler(s.config, s.services)

	s.router = gin.New()
	s.handlers.InitRoutes(s.router)

	s.user1 = repository.User{
		ID:        primitive.NewObjectID(),
		Name:      "User1",
		CreatedAt: time.Date(2020, 11, 23, 23, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2020, 11, 23, 23, 0, 0, 0, time.UTC),
	}
	s.user2 = repository.User{
		ID:        primitive.NewObjectID(),
		Name:      "User2",
		CreatedAt: time.Date(2020, 11, 23, 23, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2020, 11, 23, 23, 0, 0, 0, time.UTC),
	}
}

func (s *UsersSuite) SetupTest() {
	err := s.repos.User.DeleteAll(s.ctx)
	s.Require().NoError(err)
}

func (s *UsersSuite) TearDownTest() {}

func (s *UsersSuite) TearSuite() {}

func (s *UsersSuite) TestCreateOk() {
	w := performRequest(s.router, "POST", "/api/users", "{\"name\": \"user\"}")
	s.Require().Equal(http.StatusCreated, w.Code)

	response := ResponseUser{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.Require().NoError(err)
	s.Require().Equal("user", response.Name)
}

func (s *UsersSuite) TestCreateErrorInvalidRequest() {
	w := performRequest(s.router, "POST", "/api/users", "{}")
	s.Require().Equal(http.StatusBadRequest, w.Code)

	response := gin.H{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.Require().NoError(err)
	s.Require().Equal(gin.H{"message": "Key: 'RequestCreateUser.Name' Error:Field validation for 'Name' failed on the 'required' tag"}, response)
}

func (s *UsersSuite) TestUpdateOk() {
	err := s.services.User.Create(s.ctx, &s.user1)
	s.Require().NoError(err)

	w := performRequest(s.router, "PUT", "/api/users/"+s.user1.ID.Hex(), "{\"name\": \"user\"}")
	s.Require().Equal(http.StatusOK, w.Code)

	response := ResponseUser{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	s.Require().NoError(err)
	s.Require().Equal("user", response.Name)

	updatedUser, err := s.services.User.GetByID(s.ctx, s.user1.ID.Hex())
	s.Require().NoError(err)
	s.Require().Equal("user", updatedUser.Name)
}

func (s *UsersSuite) TestUpdateErrorInvalidRequest() {
	w := performRequest(s.router, "PUT", "/api/users/"+s.user1.ID.Hex(), "{}")
	s.Require().Equal(http.StatusBadRequest, w.Code)

	response := gin.H{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.Require().NoError(err)
	s.Require().Equal(gin.H{"message": "Key: 'RequestUpdateUser.Name' Error:Field validation for 'Name' failed on the 'required' tag"}, response)
}
func (s *UsersSuite) TestGetByIDErrorNotFound() {
	w := performRequest(s.router, "GET", "/api/users/5fbaeab741e97bef8525d6ab", "")
	s.Require().Equal(http.StatusNotFound, w.Code)

	response := errorResponse{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.Require().NoError(err)
	s.Require().Equal("mongo: no documents in result", response.Message)
}

func (s *UsersSuite) TestGetByIDErrorInvalidID() {
	w := performRequest(s.router, "GET", "/api/users/1", "")
	s.Require().Equal(http.StatusNotFound, w.Code)

	response := errorResponse{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.Require().NoError(err)
	s.Require().Equal("mongo: no documents in result", response.Message)
}

func (s *UsersSuite) TestGetByIDOk() {
	err := s.services.User.Create(s.ctx, &s.user1)
	s.Require().NoError(err)

	w := performRequest(s.router, "GET", "/api/users/"+s.user1.ID.Hex(), "")
	s.Require().Equal(http.StatusOK, w.Code)

	response := ResponseUser{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	s.Require().NoError(err)
	s.Require().Equal(s.user1.ID.Hex(), response.ID)
	s.Require().Equal(s.user1.Name, response.Name)
}

func (s *UsersSuite) TestListOkEmpty() {
	w := performRequest(s.router, "GET", "/api/users", "")
	s.Require().Equal(http.StatusOK, w.Code)
	s.Require().Equal("{\"items\":[]}", w.Body.String())
}

func (s *UsersSuite) TestListOk() {
	err := s.services.User.Create(s.ctx, &s.user1)
	s.Require().NoError(err)
	err = s.services.User.Create(s.ctx, &s.user2)
	s.Require().NoError(err)

	w := performRequest(s.router, "GET", "/api/users", "")
	s.Require().Equal(http.StatusOK, w.Code)

	response := ResponseUsers{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	s.Require().NoError(err)
	s.Require().Equal(2, len(response.Items))
	s.Require().Equal(s.user1.ID.Hex(), response.Items[0].ID)
	s.Require().Equal(s.user1.Name, response.Items[0].Name)
	s.Require().Equal(s.user2.ID.Hex(), response.Items[1].ID)
	s.Require().Equal(s.user2.Name, response.Items[1].Name)
}

func (s *UsersSuite) TestListOkWithLimitAndOffset() {
	err := s.services.User.Create(s.ctx, &s.user1)
	s.Require().NoError(err)
	err = s.services.User.Create(s.ctx, &s.user2)
	s.Require().NoError(err)

	w := performRequest(s.router, "GET", "/api/users?limit=1&offset=1", "")
	s.Require().Equal(http.StatusOK, w.Code)

	response := ResponseUsers{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	s.Require().NoError(err)
	s.Require().Equal(1, len(response.Items))
	s.Require().Equal(s.user2.ID.Hex(), response.Items[0].ID)
	s.Require().Equal(s.user2.Name, response.Items[0].Name)
}

func (s *UsersSuite) TestDeleteOk() {
	err := s.services.User.Create(s.ctx, &s.user1)
	s.Require().NoError(err)

	w := performRequest(s.router, "DELETE", "/api/users/"+s.user1.ID.Hex(), "")
	s.Require().Equal(http.StatusNoContent, w.Code)
	s.Require().Equal("", w.Body.String())

	_, err = s.services.User.GetByID(s.ctx, s.user1.ID.Hex())
	assert.Error(s.T(), err, mongo.ErrNoDocuments)
}

func (s *UsersSuite) TestDeleteErrorNotFound() {
	w := performRequest(s.router, "DELETE", "/api/users/5fbaeab741e97bef8525d6ab", "")
	s.Require().Equal(http.StatusNoContent, w.Code) // FIXME should 404
	s.Require().Equal("", w.Body.String())
}
func (s *UsersSuite) TestDeleteErrorInvalidID() {
	w := performRequest(s.router, "DELETE", "/api/users/1", "")
	s.Require().Equal(http.StatusNotFound, w.Code)

	response := errorResponse{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.Require().NoError(err)
	s.Require().Equal("mongo: no documents in result", response.Message)
}

func TestUsersSuite(t *testing.T) {
	suite.Run(t, new(UsersSuite))
}
