package datacommittee

import (
	"errors"
	"math/big"
	"testing"

	smcparis "github.com/0xPolygon/cdk-contracts-tooling/contracts/banana-paris/polygondatacommittee"
	"github.com/0xPolygon/cdk-contracts-tooling/contracts/banana/polygondatacommittee"
	"github.com/0xPolygon/cdk/log"
	erc1967proxy "github.com/0xPolygon/cdk/test/contracts/erc1967proxy"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateDataCommitteeEvent(t *testing.T) {
	// Set up testing environment
	dac, ethBackend, auth, da := newTestingEnv(t)

	// Update the committee
	requiredAmountOfSignatures := big.NewInt(2)
	URLs := []string{"1", "2", "3"}
	addrs := []common.Address{
		common.HexToAddress("0x1"),
		common.HexToAddress("0x2"),
		common.HexToAddress("0x3"),
	}
	addrsBytes := []byte{}
	for _, addr := range addrs {
		addrsBytes = append(addrsBytes, addr.Bytes()...)
	}
	_, err := da.SetupCommittee(auth, requiredAmountOfSignatures, URLs, addrsBytes)
	require.NoError(t, err)
	ethBackend.Commit()

	// Assert the committee update
	actualSetup, err := dac.getCurrentDataCommittee()
	require.NoError(t, err)
	expectedMembers := []DataCommitteeMember{}
	expectedSetup := DataCommittee{
		RequiredSignatures: uint64(len(URLs) - 1),
		AddressesHash:      crypto.Keccak256Hash(addrsBytes),
	}
	for i, url := range URLs {
		expectedMembers = append(expectedMembers, DataCommitteeMember{
			URL:  url,
			Addr: addrs[i],
		})
	}
	expectedSetup.Members = expectedMembers
	assert.Equal(t, expectedSetup, *actualSetup)
}

func init() {
	log.Init(log.Config{
		Level:   "debug",
		Outputs: []string{"stderr"},
	})
}

// This function prepare the blockchain, the wallet with funds and deploy the smc
func newTestingEnv(t *testing.T) (
	dac *Backend,
	ethBackend *simulated.Backend,
	auth *bind.TransactOpts,
	da *polygondatacommittee.Polygondatacommittee,
) {
	t.Helper()
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}
	auth, err = bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(1337))
	if err != nil {
		log.Fatal(err)
	}
	dac, ethBackend, da, err = newSimulatedDacman(t, auth)
	if err != nil {
		log.Fatal(err)
	}

	return dac, ethBackend, auth, da
}

// NewSimulatedEtherman creates an etherman that uses a simulated blockchain. It's important to notice that the ChainID of the auth
// must be 1337. The address that holds the auth will have an initial balance of 10 ETH
func newSimulatedDacman(t *testing.T, auth *bind.TransactOpts) (
	dacman *Backend,
	ethBackend *simulated.Backend,
	da *polygondatacommittee.Polygondatacommittee,
	err error,
) {
	t.Helper()
	if auth == nil {
		// read only client
		return &Backend{}, nil, nil, nil
	}
	// 10000000 ETH in wei
	balance, _ := new(big.Int).SetString("10000000000000000000000000", 10)
	address := auth.From
	genesisAlloc := map[common.Address]types.Account{
		address: {
			Balance: balance,
		},
	}
	blockGasLimit := uint64(999999999999999999)
	client := simulated.NewBackend(genesisAlloc, simulated.WithBlockGasLimit(blockGasLimit))

	// DAC Setup
	addr, _, _, err := smcparis.DeployPolygondatacommittee(auth, client.Client())
	if err != nil {
		return &Backend{}, nil, nil, err
	}
	client.Commit()
	proxyAddr, err := deployDACProxy(auth, client.Client(), addr)
	if err != nil {
		return &Backend{}, nil, nil, err
	}

	client.Commit()
	da, err = polygondatacommittee.NewPolygondatacommittee(proxyAddr, client.Client())
	if err != nil {
		return &Backend{}, nil, nil, err
	}

	_, err = da.SetupCommittee(auth, big.NewInt(0), []string{}, []byte{})
	if err != nil {
		return &Backend{}, nil, nil, err
	}
	client.Commit()

	c := &Backend{
		dataCommitteeContract: da,
	}

	return c, client, da, nil
}

func deployDACProxy(auth *bind.TransactOpts, client bind.ContractBackend, dacImpl common.Address) (common.Address, error) {
	// Deploy proxy
	dacABI, err := polygondatacommittee.PolygondatacommitteeMetaData.GetAbi()
	if err != nil {
		return common.Address{}, err
	}
	if dacABI == nil {
		return common.Address{}, errors.New("GetABI returned nil")
	}
	initializeCallData, err := dacABI.Pack("initialize")
	if err != nil {
		return common.Address{}, err
	}
	proxyAddr, err := deployProxy(
		auth,
		client,
		dacImpl,
		initializeCallData,
	)
	if err != nil {
		return common.Address{}, err
	}
	log.Debugf("DAC proxy deployed at", proxyAddr)

	return proxyAddr, nil
}

func deployProxy(auth *bind.TransactOpts,
	client bind.ContractBackend,
	implementationAddr common.Address,
	initializeParams []byte) (common.Address, error) {
	addr, _, _, err := erc1967proxy.DeployErc1967proxy(
		auth,
		client,
		implementationAddr,
		initializeParams,
	)

	return addr, err
}
