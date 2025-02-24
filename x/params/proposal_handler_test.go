package params_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/params/keeper"
	"github.com/cosmos/cosmos-sdk/x/params/testutil"
	"github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// StakingKeeper defines the expected staking keeper
type StakingKeeper interface {
	MaxValidators(ctx sdk.Context) (res uint32)
}

type HandlerTestSuite struct {
	suite.Suite

	ctx           sdk.Context
	govHandler    govv1beta1.Handler
	stakingKeeper StakingKeeper
}

func (suite *HandlerTestSuite) SetupTest() {
	var paramsKeeper keeper.Keeper
	app, err := simtestutil.Setup(
		testutil.AppConfig,
		&paramsKeeper,
		&suite.stakingKeeper,
	)
	suite.Require().NoError(err)

	suite.ctx = app.BaseApp.NewContext(false, tmproto.Header{})
	suite.govHandler = params.NewParamChangeProposalHandler(paramsKeeper)
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}

func testProposal(changes ...proposal.ParamChange) *proposal.ParameterChangeProposal {
	return proposal.NewParameterChangeProposal("title", "description", changes)
}

func (suite *HandlerTestSuite) TestProposalHandler() {
	testCases := []struct {
		name     string
		proposal *proposal.ParameterChangeProposal
		onHandle func()
		expErr   bool
	}{
		{
			"all fields",
			testProposal(proposal.NewParamChange(stakingtypes.ModuleName, string(stakingtypes.KeyMaxValidators), "1")),
			func() {
				maxVals := suite.stakingKeeper.MaxValidators(suite.ctx)
				suite.Require().Equal(uint32(1), maxVals)
			},
			false,
		},
		{
			"invalid type",
			testProposal(proposal.NewParamChange(stakingtypes.ModuleName, string(stakingtypes.KeyMaxValidators), "-")),
			func() {},
			true,
		},
		// {
		// 	"omit empty fields",
		// 	testProposal(proposal.ParamChange{
		// 		Subspace: govtypes.ModuleName,
		// 		Key:      string(govv1.ParamStoreKeyDepositParams),
		// 		Value:    `{"min_deposit": [{"denom": "uatom","amount": "64000000"}], "max_deposit_period": "172800000000000"}`,
		// 	}),
		// 	func() {
		// 		depositParams := suite.app.GovKeeper.GetDepositParams(suite.ctx)
		// 		defaultPeriod := govv1.DefaultPeriod
		// 		suite.Require().Equal(govv1.DepositParams{
		// 			MinDeposit:       sdk.NewCoins(sdk.NewCoin("uatom", sdk.NewInt(64000000))),
		// 			MaxDepositPeriod: &defaultPeriod,
		// 		}, depositParams)
		// 	},
		// 	false,
		// },
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			err := suite.govHandler(suite.ctx, tc.proposal)
			if tc.expErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				tc.onHandle()
			}
		})
	}
}
