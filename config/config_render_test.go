package config

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	s = []string{"A= {{B}}\n", "B= {{C}}\nC={{A}}\n"}
)

type testCaseData struct {
	name                 string
	contents             []string
	envVars              map[string]string
	expectedMerged       string
	expectedRenderConfig string
	expectedError        error
}

func TestConfigRenderMerge(t *testing.T) {
	var tests = []testCaseData{
		{
			name:                 "Merge 2 elements",
			contents:             []string{"A=1\n", "B=2\n"},
			expectedRenderConfig: "A = 1\nB = 2\n",
		},
		{
			name:                 "Merge 3 elements",
			contents:             []string{"A=1\n", "B=2\n", "C=3\n"},
			expectedRenderConfig: "A = 1\nB = 2\nC = 3\n",
		},
		{
			name:                 "Merge 3 elements, overlapped",
			contents:             []string{"A=1\n", "A=2\nB=2\n", "A=3\nC=3\n"},
			expectedRenderConfig: "A = 3\nB = 2\nC = 3\n",
		},
		{
			name:                 "Merge 3 elements, overlapped final var",
			contents:             []string{"A=1\n", "A=2\nB=2\n", "A={{VAR}}\nC=3\n"},
			expectedRenderConfig: "A = {{VAR}}\nB = 2\nC = 3\n",
			expectedError:        ErrMissingVars,
		},
	}
	executeCases(t, tests)
}

func TestConfigRenderDetectCycle(t *testing.T) {
	var tests = []testCaseData{
		{
			name:                 "Cycle 3 elements",
			contents:             []string{"A= {{B}}\n", "B= {{C}}\nC={{A}}\n"},
			expectedMerged:       "A = {{B}}\nB = {{C}}\nC = {{A}}\n",
			expectedRenderConfig: "A = {{B}}\nB = {{C}}\nC = {{A}}\n",
			expectedError:        ErrCycleVars,
		},
		{
			name:                 "Cycle 2 elements",
			contents:             []string{"A= {{B}}\n", "B= {{A}}\n"},
			expectedRenderConfig: "A = {{B}}\nB = {{A}}\n",
			expectedError:        ErrCycleVars,
		},
		{
			name:                 "Cycle 1 elements",
			contents:             []string{"A= {{A}}\n", ""},
			expectedRenderConfig: "A = {{A}}\n",
			expectedError:        ErrCycleVars,
		},
	}
	executeCases(t, tests)
}

func TestConfigRenderTypes(t *testing.T) {
	var tests = []testCaseData{
		{
			name: "Cycle 3 elements B, break var",
			contents: []string{"INT_VALUE={{MY_INT}}\n STR_VALUE= \"{{MY_STR}}\"\n MYBOOL={{MY_BOOL}}\n",
				"MY_STR=\"a string\"\nMY_INT=4\nMY_BOOL=true\nNO_RESOLVED={{NOT_DEFINED_VAR}}\n"},
			envVars:              map[string]string{"UTCR_B": "4"},
			expectedError:        ErrMissingVars,
			expectedRenderConfig: "INT_VALUE = 4\nMYBOOL = true\nMY_BOOL = true\nMY_INT = 4\nMY_STR = \"a string\"\nNO_RESOLVED = {{NOT_DEFINED_VAR}}\nSTR_VALUE = \"a string\"\n",
		},
	}
	executeCases(t, tests)
}

func TestConfigRenderComposedValues(t *testing.T) {
	var tests = []testCaseData{
		{
			name:                 "Composed var",
			contents:             []string{"A=\"path\"\n", "B= \"{{A}}to\"\n"},
			expectedRenderConfig: "A = \"path\"\nB = \"pathto\"\n",
		},
	}
	executeCases(t, tests)
}

func TestConfigRenderCycleBrokenByEnvVar(t *testing.T) {
	var tests = []testCaseData{
		{
			name:                 "Cycle 3 elements B, break var",
			contents:             []string{"A= {{B}}\n", "B= {{C}}\nC={{A}}\n"},
			envVars:              map[string]string{"UTCR_B": "4"},
			expectedRenderConfig: "A = 4\nB = 4\nC = 4\n",
		},
		{
			name:                 "Cycle 3 elements A, break var",
			contents:             []string{"A= {{B}}\n", "B= {{C}}\nC={{A}}\n"},
			envVars:              map[string]string{"UTCR_A": "4"},
			expectedRenderConfig: "A = 4\nB = 4\nC = 4\n",
		},
		{
			name:                 "Cycle 3 elements C, break var",
			contents:             []string{"A= {{B}}\n", "B= {{C}}\nC={{A}}\n"},
			envVars:              map[string]string{"UTCR_C": "4"},
			expectedRenderConfig: "A = 4\nB = 4\nC = 4\n",
		},
	}
	executeCases(t, tests)
}

func TestConfigRenderOverrideByEnvVars(t *testing.T) {
	var tests = []testCaseData{
		{
			name:                 "Variable is not set in config file but override  as number",
			contents:             []string{"A={{C}}\n"},
			envVars:              map[string]string{"UTCR_C": "4"},
			expectedRenderConfig: "A = 4\n",
		},
		// Notice that the exported variable have the quotes
		{
			name:                 "Variable is not set in config file but override  as string",
			contents:             []string{"A={{C}}\n"},
			envVars:              map[string]string{"UTCR_C": "\"4\""},
			expectedRenderConfig: "A = \"4\"\n",
		},
	}
	executeCases(t, tests)
}

func TestConfigRenderPropagateType(t *testing.T) {
	var tests = []testCaseData{
		{
			name:                 "propagateType: set directly",
			contents:             []string{"A=\"hello\"\n", "B= \"{{A}}\"\n"},
			expectedRenderConfig: "A = \"hello\"\nB = \"hello\"\n",
		},
		{
			name:                 "propagateType: set directly",
			contents:             []string{"A=\"hello\"\n", "B=\"{{A}}\"\n"},
			envVars:              map[string]string{"UTCR_A": "you"},
			expectedRenderConfig: "A = \"hello\"\nB = \"you\"\n",
		},
	}
	executeCases(t, tests)
}

func TestConfigRenderComplexStruct(t *testing.T) {
	defaultValues := `
		[Etherman]
	URL="http://generic_url"
	ForkIDChunkSize=100
	[Etherman.EthermanConfig]
		URL="http://localhost:8545"
`
	confiFile := `
		[Etherman.EthermanConfig]
		URL="{{Etherman.URL}}"
	`
	var tests = []testCaseData{
		{
			name:                 "Complex struct merge",
			contents:             []string{defaultValues, confiFile},
			expectedRenderConfig: "\n[Etherman]\n  ForkIDChunkSize = 100\n  URL = \"http://generic_url\"\n\n  [Etherman.EthermanConfig]\n    URL = \"http://generic_url\"\n",
		},
		// This test Etherman.URL doesnt change because is not a var, it will change value on viper stage
		{
			name:                 "Complex struct merge override env-var, but we must propagate the string type",
			contents:             []string{defaultValues, confiFile},
			envVars:              map[string]string{"UTCR_Etherman_URL": "env"},
			expectedRenderConfig: "\n[Etherman]\n  ForkIDChunkSize = 100\n  URL = \"http://generic_url\"\n\n  [Etherman.EthermanConfig]\n    URL = \"env\"\n",
		},
	}
	executeCases(t, tests)
}

func TestConfigRenderConvertFileToToml(t *testing.T) {
	jsonFile := `{
	"rollupCreationBlockNumber": 63,
  "rollupManagerCreationBlockNumber": 57,
  "genesisBlockNumber": 63,
  "L1Config": {
    "chainId": 271828,
    "polygonZkEVMGlobalExitRootAddress": "0x1f7ad7caA53e35b4f0D138dC5CBF91aC108a2674",
    "polygonRollupManagerAddress": "0x2F50ef6b8e8Ee4E579B17619A92dE3E2ffbD8AD2",
    "polTokenAddress": "0xEdE9cf798E0fE25D35469493f43E88FeA4a5da0E",
    "polygonZkEVMAddress": "0x1Fe038B54aeBf558638CA51C91bC8cCa06609e91"
  }
}
`
	data, err := convertFileToToml(jsonFile, "json")
	require.NoError(t, err)
	require.Equal(t, "genesisBlockNumber = 63.0\nrollupCreationBlockNumber = 63.0\nrollupManagerCreationBlockNumber = 57.0\n\n[L1Config]\n  chainId = 271828.0\n  polTokenAddress = \"0xEdE9cf798E0fE25D35469493f43E88FeA4a5da0E\"\n  polygonRollupManagerAddress = \"0x2F50ef6b8e8Ee4E579B17619A92dE3E2ffbD8AD2\"\n  polygonZkEVMAddress = \"0x1Fe038B54aeBf558638CA51C91bC8cCa06609e91\"\n  polygonZkEVMGlobalExitRootAddress = \"0x1f7ad7caA53e35b4f0D138dC5CBF91aC108a2674\"\n", data)
}

/*
TODO: This test generate this, is the same?
[PrivateKey]
    Password = "testonly"
    Path = "./test/sequencer.keystore"

func TestConfigRenderValueIsAObject(t *testing.T) {
	var tests = []testCaseData{
		{
			name:                 "Complex struct object inside var",
			contents:             []string{"PrivateKey = {Path = \"./test/sequencer.keystore\", Password = \"testonly\"}"},
			expectedRenderConfig: "PrivateKey = {Path = \"./test/sequencer.keystore\", Password = \"testonly\"}",
		},
	}
	executeCases(t, tests)
}
*/

type configRenderTestData struct {
	Sut     *ConfigRender
	EnvMock *osLookupEnvMock
}

func newConfigRenderTestData(data []string) configRenderTestData {
	envMock := &osLookupEnvMock{
		Env: map[string]string{},
	}
	filesData := make([]FileData, len(data))
	for i, d := range data {
		filesData[i] = FileData{Name: fmt.Sprintf("file%d", i), Content: d}
	}
	return configRenderTestData{
		EnvMock: envMock,
		Sut: &ConfigRender{
			FilesData:         filesData,
			LookupEnvFunc:     envMock.LookupEnv,
			EnvinormentPrefix: "UTCR",
		},
	}
}

type osLookupEnvMock struct {
	Env map[string]string
}

func (m *osLookupEnvMock) LookupEnv(key string) (string, bool) {
	val, exists := m.Env[key]
	return val, exists
}

func executeCases(t *testing.T, tests []testCaseData) {
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testData := newConfigRenderTestData(tt.contents)
			if tt.envVars != nil {
				testData.EnvMock.Env = tt.envVars
			}
			if tt.expectedMerged != "" {
				merged, err := testData.Sut.Merge()
				require.NoError(t, err)
				require.Equal(t, tt.expectedMerged, merged)
			}
			res, err := testData.Sut.Render()
			if tt.expectedError != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.expectedError)
			} else {
				require.NoError(t, err)
			}
			if len(tt.expectedRenderConfig) > 0 {
				require.Equal(t, tt.expectedRenderConfig, res)
			}
		})
	}
}
