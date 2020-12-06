package rmq

import (
	"testing"
	"time"

	"github.com/streadway/amqp"
	"github.com/stretchr/testify/suite"
	"github.com/zaharinea/go-example/config"
	"github.com/zaharinea/go-example/pkg/repository"
	"github.com/zaharinea/go-example/pkg/service"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/context"
)

type RmqHanlersSuite struct {
	suite.Suite
	ctx         context.Context
	config      *config.Config
	db          *mongo.Database
	repos       *repository.Repository
	services    *service.Service
	rmqHandlers *Handler
}

func (s *RmqHanlersSuite) SetupSuite() {
	s.ctx = context.Background()
	s.config = config.NewTestingConfig()
	dbClient := repository.InitDbClient(s.config)
	s.db = dbClient.Database(s.config.MongoDbName)
	s.repos = repository.NewRepository(s.db)
	s.services = service.NewService(s.repos)
	s.rmqHandlers = NewHandler(s.config, s.repos)
}

func (s *RmqHanlersSuite) SetupTest() {
	err := s.repos.Account.DeleteAll(s.ctx)
	s.Require().NoError(err)
}

func (s *RmqHanlersSuite) TearDownTest() {}

func (s *RmqHanlersSuite) TearSuite() {}

func (s *RmqHanlersSuite) TestHandlerCompanyEvent() {
	msg := amqp.Delivery{Body: []byte("test")}
	result := s.rmqHandlers.HandlerCompanyEvent(s.ctx, msg)
	s.Require().Equal(true, result)
}

func (s *RmqHanlersSuite) TestHandlerAccountEvent() {
	accountEvent := `{
		"external_id":"1",
		"name":"account1",
		"created_at":"2020-11-20T00:03:00.000+00:03",
		"updated_at":"2020-11-21T00:03:00.000+00:03"
	}`

	msg := amqp.Delivery{Body: []byte(accountEvent)}
	result := s.rmqHandlers.HandlerAccountEvent(s.ctx, msg)
	s.Require().Equal(true, result)

	dbAccount, err := s.repos.Account.GetByExternalID(s.ctx, "1")
	s.Require().NoError(err)
	s.Require().Equal("account1", dbAccount.Name)
	s.Require().Equal(time.Date(2020, 11, 20, 0, 0, 0, 0, time.UTC), dbAccount.CreatedAt)
	s.Require().Equal(time.Date(2020, 11, 21, 0, 0, 0, 0, time.UTC), dbAccount.UpdatedAt)
}

func (s *RmqHanlersSuite) TestHandlerAccountEventSkipOld() {
	_, err := s.repos.Account.CreateOrUpdate(s.ctx, repository.Account{
		ID:         primitive.NewObjectID(),
		ExternalID: "1",
		Name:       "account1",
		CreatedAt:  time.Date(2020, 11, 20, 0, 0, 0, 0, time.UTC),
		UpdatedAt:  time.Date(2020, 11, 23, 0, 0, 0, 0, time.UTC),
	}, true)
	s.Require().NoError(err)

	oldAccountEvent := `{
		"external_id":"1",
		"name":"account0",
		"created_at":"2020-11-20T00:00:00.000Z",
		"updated_at":"2020-11-22T23:59:59.999Z"
	}`
	msg := amqp.Delivery{Body: []byte(oldAccountEvent)}
	result := s.rmqHandlers.HandlerAccountEvent(s.ctx, msg)
	s.Require().Equal(true, result)

	dbAccount, err := s.repos.Account.GetByExternalID(s.ctx, "1")
	s.Require().NoError(err)
	s.Require().Equal("account1", dbAccount.Name)
}

func TestRmqHanlersSuite(t *testing.T) {
	suite.Run(t, new(RmqHanlersSuite))
}
