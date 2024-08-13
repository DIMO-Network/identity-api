package account

import (
	"context"
	"fmt"
	"slices"
	"testing"
	"time"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/config"
	test "github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type AccountRepoTestSuite struct {
	suite.Suite
	ctx       context.Context
	pdb       db.Store
	container *postgres.PostgresContainer
	repo      *Repository
	settings  config.Settings
}

func (o *AccountRepoTestSuite) SetupSuite() {
	o.ctx = context.Background()
	o.pdb, o.container = test.StartContainerDatabase(o.ctx, o.T(), "../../../migrations")
	logger := zerolog.Nop()
	o.repo = &Repository{base.NewRepository(o.pdb, o.settings, &logger)}
}

// TearDownTest after each test truncate tables
func (s *AccountRepoTestSuite) TearDownTest() {
	s.Require().NoError(s.container.Restore(s.ctx))
}

// TearDownSuite cleanup at end by terminating container
func (o *AccountRepoTestSuite) TearDownSuite() {
	fmt.Printf("shutting down postgres at with session: %s \n", o.container.SessionID())

	if err := o.container.Terminate(o.ctx); err != nil {
		o.T().Fatal(err)
	}
}

// Test Runner
func TestAccountRepoTestSuite(t *testing.T) {
	suite.Run(t, new(AccountRepoTestSuite))
}

func (o *AccountRepoTestSuite) Test_GetAccountBySigner() {
	numKernels := 5
	createdAt := time.Now().UTC().Truncate(time.Second)
	signerAdded := time.Now().UTC().Truncate(time.Second)
	kernels := []common.Address{}
	_, signer, err := test.GenerateWallet()
	o.NoError(err)
	for range numKernels {
		_, kernel, err := test.GenerateWallet()
		o.NoError(err)
		kernels = append(kernels, *kernel)
		k := models.Account{
			Signer:      signer.Bytes(),
			Kernel:      kernel.Bytes(),
			CreatedAt:   createdAt,
			SignerAdded: signerAdded,
		}
		if err := k.Insert(o.ctx, o.pdb.DBS().Writer, boil.Infer()); err != nil {
			assert.NoError(o.T(), err)
		}
	}

	resp, err := o.repo.GetAccount(o.ctx, gmodel.AccountBy{
		Signer: signer,
	})
	o.NoError(err)

	o.Equal(resp.Signer, *signer)
	o.Equal(resp.Kernel[0].Signer.Address, *signer)
	o.Equal(len(resp.Kernel), numKernels)
	o.Equal(resp.Kernel[0].CreatedAt, createdAt)
	o.Equal(resp.Kernel[0].Signer.SignerAdded, signerAdded)
	for _, k := range resp.Kernel {
		slices.Contains(kernels, k.Address)
	}
}

func (o *AccountRepoTestSuite) Test_GetAccountByKernel() {
	numKernels := 5
	createdAt := time.Now().UTC().Truncate(time.Second)
	signerAdded := time.Now().UTC().Truncate(time.Second)
	kernels := []models.Account{}
	_, signer, err := test.GenerateWallet()
	o.NoError(err)
	for range numKernels {
		_, kernel, err := test.GenerateWallet()
		o.NoError(err)
		k := models.Account{
			Signer:      signer.Bytes(),
			Kernel:      kernel.Bytes(),
			CreatedAt:   createdAt,
			SignerAdded: signerAdded,
		}
		kernels = append(kernels, k)
		if err := k.Insert(o.ctx, o.pdb.DBS().Writer, boil.Infer()); err != nil {
			assert.NoError(o.T(), err)
		}
	}

	kernelAddr := common.BytesToAddress(kernels[0].Kernel)
	resp, err := o.repo.GetAccount(o.ctx, gmodel.AccountBy{
		Kernel: &kernelAddr,
	})
	o.NoError(err)

	o.Equal(resp.Signer, *signer)
	o.Equal(resp.Kernel[0].Signer.Address, *signer)
	o.Equal(len(resp.Kernel), 1)
	o.Equal(resp.Kernel[0].Address, kernelAddr)
	o.Equal(resp.Kernel[0].CreatedAt, createdAt)
	o.Equal(resp.Kernel[0].Signer.SignerAdded, signerAdded)
}
