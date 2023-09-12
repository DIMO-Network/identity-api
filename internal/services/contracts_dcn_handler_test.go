package services

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/DIMO-Network/identity-api/internal/config"
	test "github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"
	"github.com/Shopify/sarama"
	"github.com/Shopify/sarama/mocks"
	"github.com/goccy/go-json"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type DCNConsumerTestSuite struct {
	suite.Suite
	ctx       context.Context
	pdb       db.Store
	container testcontainers.Container
	settings  config.Settings
	logger    zerolog.Logger
}

func (o *DCNConsumerTestSuite) SetupSuite() {
	o.ctx = context.Background()
	o.pdb, o.container = test.StartContainerDatabase(o.ctx, o.T(), "../../migrations")

	o.settings = config.Settings{
		DCNRegistryAddr:     contractEventData.Contract.String(),
		DIMORegistryChainID: 80001,
	}

	o.logger = zerolog.New(os.Stdout).With().Timestamp().Str("app", test.DBSettings.Name).Logger()
}

// TearDownTest after each test truncate tables
func (s *DCNConsumerTestSuite) TearDownTest() {
	test.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
}

// TearDownSuite cleanup at end by terminating container
func (o *DCNConsumerTestSuite) TearDownSuite() {
	fmt.Printf("shutting down postgres at with session: %s \n", o.container.SessionID())

	if err := o.container.Terminate(o.ctx); err != nil {
		o.T().Fatal(err)
	}
}

// Test Runner
func TestDCNConsumerTestSuite(t *testing.T) {
	suite.Run(t, new(DCNConsumerTestSuite))
}

func (o *DCNConsumerTestSuite) Test_NewNode_Consume_Success() {
	contractEventData.EventName = NewNode.String()
	_, wallet, err := test.GenerateWallet()
	o.NoError(err)

	var eventData = NewDCNNodeEventData{
		Node:  test.GenerateDCNNode(),
		Owner: *wallet,
	}

	config := mocks.NewTestConfig()
	consumer := mocks.NewConsumer(o.T(), config)

	contractEventConsumer := NewContractsEventsConsumer(o.pdb, &o.logger, &o.settings)
	expectedBytes := eventBytes(eventData, contractEventData, o.T())

	consumer.ExpectConsumePartition(o.settings.ContractsEventTopic, 0, 0).YieldMessage(&sarama.ConsumerMessage{Value: expectedBytes})

	outputTest, err := consumer.ConsumePartition(o.settings.ContractsEventTopic, 0, 0)
	o.NoError(err)

	m := <-outputTest.Messages()
	var e shared.CloudEvent[json.RawMessage]
	err = json.Unmarshal(m.Value, &e)
	o.NoError(err)

	err = contractEventConsumer.Process(o.ctx, &e)
	o.NoError(err)

	dcn, err := models.DCNS().All(o.ctx, o.pdb.DBS().Reader.DB)
	o.NoError(err)

	o.Len(dcn, 1)
	o.Equal(eventData.Node, dcn[0].Node)
	o.Equal(eventData.Owner.Bytes(), dcn[0].OwnerAddress)
}

func (o *DCNConsumerTestSuite) Test_NewDCNExpiration_Consume_Success() {
	contractEventData.EventName = NewExpiration.String()

	_, owner, err := test.GenerateWallet()
	o.NoError(err)

	currTime := time.Now().UTC().Truncate(time.Second)

	var eventData = NewDCNExpirationEventData{
		Node:       test.GenerateDCNNode(),
		Expiration: int(currTime.Unix()),
	}

	config := mocks.NewTestConfig()
	consumer := mocks.NewConsumer(o.T(), config)

	d := models.DCN{
		Node:         eventData.Node,
		OwnerAddress: owner.Bytes(),
	}

	err = d.Insert(o.ctx, o.pdb.DBS().Writer.DB, boil.Infer())
	o.NoError(err)

	contractEventConsumer := NewContractsEventsConsumer(o.pdb, &o.logger, &o.settings)
	expectedBytes := eventBytes(eventData, contractEventData, o.T())

	consumer.ExpectConsumePartition(o.settings.ContractsEventTopic, 0, 0).YieldMessage(&sarama.ConsumerMessage{Value: expectedBytes})

	outputTest, err := consumer.ConsumePartition(o.settings.ContractsEventTopic, 0, 0)
	o.NoError(err)

	m := <-outputTest.Messages()
	var e shared.CloudEvent[json.RawMessage]
	err = json.Unmarshal(m.Value, &e)
	o.NoError(err)

	err = contractEventConsumer.Process(o.ctx, &e)
	o.NoError(err)

	dcn, err := models.DCNS().All(o.ctx, o.pdb.DBS().Reader.DB)
	o.NoError(err)

	o.Len(dcn, 1)
	o.Equal(eventData.Node, dcn[0].Node)
	o.Equal(owner.Bytes(), dcn[0].OwnerAddress)
	o.Equal(currTime, dcn[0].Expiration.Time)
}

func (o *DCNConsumerTestSuite) Test_DCNNameChanged_Consume_Success() {
	contractEventData.EventName = NameChanged.String()

	_, owner, err := test.GenerateWallet()
	o.NoError(err)

	newName := "SomeMockName"
	var eventData = DCNNameChangedEventData{
		Node: test.GenerateDCNNode(),
		Name: newName,
	}

	config := mocks.NewTestConfig()
	consumer := mocks.NewConsumer(o.T(), config)

	d := models.DCN{
		Node:         eventData.Node,
		OwnerAddress: owner.Bytes(),
	}

	err = d.Insert(o.ctx, o.pdb.DBS().Writer.DB, boil.Infer())
	o.NoError(err)

	contractEventConsumer := NewContractsEventsConsumer(o.pdb, &o.logger, &o.settings)
	expectedBytes := eventBytes(eventData, contractEventData, o.T())

	consumer.ExpectConsumePartition(o.settings.ContractsEventTopic, 0, 0).YieldMessage(&sarama.ConsumerMessage{Value: expectedBytes})

	outputTest, err := consumer.ConsumePartition(o.settings.ContractsEventTopic, 0, 0)
	o.NoError(err)

	m := <-outputTest.Messages()
	var e shared.CloudEvent[json.RawMessage]
	err = json.Unmarshal(m.Value, &e)
	o.NoError(err)

	err = contractEventConsumer.Process(o.ctx, &e)
	o.NoError(err)

	dcn, err := models.DCNS().All(o.ctx, o.pdb.DBS().Reader.DB)
	o.NoError(err)

	o.Len(dcn, 1)
	o.Equal(eventData.Node, dcn[0].Node)
	o.Equal(owner.Bytes(), dcn[0].OwnerAddress)
	o.Equal(eventData.Name, dcn[0].Name.String)
}
