package operations

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"time"

	"github.com/umbracle/ethgo/jsonrpc"
	"github.com/umbracle/ethgo/wallet"
)

const (
	L2RPCURL                = "http://localhost:8123"
	L1RPCURL                = "http://localhost:8545"
	L2ChainID               = 1001
	PrivateKeyWithFundsOnL2 = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	DefaultCDKStack         = "cdk-validium-node-1"

	NodeENV           = "CDK_NODE"
	SequenceSenderENV = "CDK_SEQUENCE_SENDER"
	RPCENV            = "CDK_RPC"
	L1ENV             = "CDK_L1"
	ProverENV         = "CDK_PROVER"
	AggLayerENV       = "CDK_AGGLAYER"
	DBENV             = "CDK_DB"
	AggregatorENV     = "CDK_AGGREGATOR"
	CDKStackENV       = "CDK_STACk"
)

type CDKStack struct {
	Node           string
	SequenceSender string
	RPC            string
	L1             string
	Prover         string
	AggLayer       string
	DB             string
	Aggregator     string
}

type CDKComponent struct {
	CDKStackKey      string
	envVar           string
	supportedOptions []string
}

var (
	preconfiguredCDKStacks = map[string]CDKStack{
		DefaultCDKStack: {
			Node:           "cdk-validium-node-1",
			SequenceSender: "cdk-validium-cdk-sequence-sender-1",
			RPC:            "cdk-validium-hermez-rpc-1",
			L1:             "cdk-l1",
			Prover:         "cdk-prover-1",
			AggLayer:       "",
			DB:             "cdk-postgres",
			Aggregator:     "cdk-validium-hermez-aggregator-direct-1",
		},
	}
	cdkComponents = []CDKComponent{
		{
			CDKStackKey:      "Node",
			envVar:           NodeENV,
			supportedOptions: []string{"cdk-validium-node-1"},
		},
		{
			CDKStackKey: "SequenceSender",
			envVar:      SequenceSenderENV,
			supportedOptions: []string{
				"cdk-validium-hermez-sequence-sender-1",
				"cdk-validium-cdk-sequence-sender-1",
			},
		},
		{
			CDKStackKey:      "RPC",
			envVar:           RPCENV,
			supportedOptions: []string{"cdk-validium-hermez-rpc-1"},
		},
		{
			CDKStackKey:      "L1",
			envVar:           RPCENV,
			supportedOptions: []string{"cdk-l1"},
		},
		{
			CDKStackKey:      "Prover",
			envVar:           ProverENV,
			supportedOptions: []string{"cdk-prover-1"},
		},
		{
			CDKStackKey:      "AggLayer",
			envVar:           AggLayerENV,
			supportedOptions: []string{""},
		},
		{
			CDKStackKey:      "DB",
			envVar:           DBENV,
			supportedOptions: []string{"cdk-postgres"},
		},
		{
			CDKStackKey: "Aggregator",
			envVar:      AggregatorENV,
			supportedOptions: []string{
				"cdk-validium-hermez-aggregator-direct-1",
				"cdk-validium-hermez-aggregator-agglayer-1",
			},
		},
	}
)

func RunCDKStack() error {
	stack, err := getStack()
	if err != nil {
		return err
	}

	jsonStack, err := json.MarshalIndent(stack, "", "    ")
	if err != nil {
		return err
	}
	cleanStackStr := strings.Replace(string(jsonStack), "\"", "", -1)
	cleanStackStr = strings.Replace(cleanStackStr, ": ,", ": -,", -1)
	cleanStackStr = strings.Replace(cleanStackStr, ": ,", ": -,", -1)
	cleanStackStr = strings.Trim(cleanStackStr, "{}")
	fmt.Printf("starting docker with the following components:\n%s\n", cleanStackStr)
	if out, err := exec.Command(
		"bash", "-l", "-c", fmt.Sprintf(
			"docker compose up -d %s %s %s %s %s %s %s %s",
			stack.Node, stack.SequenceSender, stack.RPC, stack.L1,
			stack.Prover, stack.AggLayer, stack.DB, stack.Aggregator,
		),
	).CombinedOutput(); err != nil {
		fmt.Println(string(out))
		return err
	}
	fmt.Println("docker is up, waiting for L2 blocks to appear")
	client, err := jsonrpc.NewClient(L2RPCURL)
	if err != nil {
		return err
	}
	var l2Started bool
	for i := 0; i < 120; i++ {
		bn, _ := client.Eth().BlockNumber()
		if bn > 3 {
			l2Started = true
			break
		}
		time.Sleep(time.Second)
	}
	if !l2Started {
		return errors.New("CDK failed to start, can't get new blocks form RPC")
	}
	return nil
}

func StopCDKStack() error {
	fmt.Println("stopping containers")
	return exec.Command("docker", "compose", "down").Run()
}

func getStack() (*CDKStack, error) {
	baseStackName := DefaultCDKStack
	if preconfigStackName := os.Getenv(CDKStackENV); preconfigStackName != "" {
		if _, ok := preconfiguredCDKStacks[preconfigStackName]; !ok {
			return nil, fmt.Errorf(
				"there is no pre configured stack named %s (trying to load from the ENV var %s)",
				preconfigStackName, CDKStackENV,
			)
		} else {
			baseStackName = preconfigStackName
		}
	}
	stack := preconfiguredCDKStacks[baseStackName]
	for _, component := range cdkComponents {
		if componentFromEnv := os.Getenv(component.envVar); componentFromEnv != "" {
			var supported bool
			for _, supportedOption := range component.supportedOptions {
				if supportedOption == componentFromEnv {
					supported = true
					break
				}
			}
			if !supported {
				return nil, fmt.Errorf(
					"%s is not a supported option for %s component. Available options are: %+v",
					componentFromEnv, component.CDKStackKey, component.supportedOptions,
				)
			}
			reflect.ValueOf(&stack).Elem().FieldByName(component.CDKStackKey).SetString(componentFromEnv)
		}
	}

	return &stack, nil
}

func GetL2Wallet() (*wallet.Key, error) {
	prvKeyBytes, err := hex.DecodeString(PrivateKeyWithFundsOnL2)
	if err != nil {
		return nil, err
	}
	return wallet.NewWalletFromPrivKey(prvKeyBytes)
}
