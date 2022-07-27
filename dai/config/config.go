// Copyright (c) 2021 PaddlePaddle Authors. All Rights Reserved.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"strings"

	"github.com/spf13/viper"

	"github.com/PaddlePaddle/PaddleDTX/dai/util/file"
)

var (
	logConf      *Log
	executorConf *ExecutorConf
	cliConf      *ExecutorBlockchainConf
)

// ExecutorConf defines the configuration info required for excutor node startup,
// and convert it to a struct by parsing 'conf/config.toml'.
type ExecutorConf struct {
	Name            string // executor node name
	ListenAddress   string // the port on which the executor node is listening
	PublicAddress   string // local grpc host
	PrivateKey      string // private key
	PaddleFLAddress string
	PaddleFLRole    int
	KeyPath         string            // key path, include private key and public key
	HttpServer      *HttpServerConf   // include executor node's httpserver configuration
	Mode            *ExecutorModeConf // the task execution type
	Mpc             *ExecutorMpcConf
	Storage         *ExecutorStorageConf // model storage and prediction results storage
	Blockchain      *ExecutorBlockchainConf
}

// HttpServerConf defines the configuration required to start the executor node's httpserver
// 'AllowCros' decides whether to allow cross-domain requests, the default is false
type HttpServerConf struct {
	Switch      string
	HttpAddress string
	HttpPort    string
	AllowCros   bool
}

// ExecutorModeConf defines the task execution type, such as proxy-execution or self-execution.
// "Self" is suitable for the executor node and the dataOwner node are the same organization and execute by themselves,
// and the executor node can download sample files from the dataOwner node without permission application.
type ExecutorModeConf struct {
	Type string
	Self *XuperDBConf
}

// ExecutorMpcConf defines the features of the mpc process
type ExecutorMpcConf struct {
	TrainTaskLimit   int
	PredictTaskLimit int
	RpcTimeout       int // rpc request timeout between executor nodes
	TaskLimitTime    int
}

// ExecutorStorageConf defines the storage used by the executor,
// include model storage and prediction results storage, and evaluation storage and live evaluation storage.
// the prediction results storage support 'XuperDB' and 'Local' two storage mode.
type ExecutorStorageConf struct {
	Type                       string
	LocalModelStoragePath      string
	LocalEvaluationStoragePath string
	LiveEvaluationStoragePath  string // live evaluation results storage path
	XuperDB                    *XuperDBConf
	Local                      *PredictLocalConf
}

// XuperDBConf defines the XuperDB's endpoint, used to upload or download files
type XuperDBConf struct {
	PrivateKey string
	Host       string
	KeyPath    string
	NameSpace  string
	ExpireTime int64
}

// PredictLocalConf defines the local path of prediction results storage
type PredictLocalConf struct {
	LocalPredictStoragePath string
}

// ExecutorBlockchainConf defines the configuration required to invoke blockchain contracts
type ExecutorBlockchainConf struct {
	Type   string
	Xchain *XchainConf // only 'xchain' is supported
}

type XchainConf struct {
	Mnemonic        string
	ContractName    string
	ContractAccount string
	ChainAddress    string
	ChainName       string
}

// Log defines the storage path of the logs generated by the executor node at runtime
type Log struct {
	Level string
	Path  string
}

// InitConfig parses configuration file
func InitConfig(configPath string) error {
	v := viper.New()
	v.SetConfigFile(configPath)
	if err := v.ReadInConfig(); err != nil {
		return err
	}
	logConf = new(Log)
	err := v.Sub("log").Unmarshal(logConf)
	if err != nil {
		return err
	}
	executorConf = new(ExecutorConf)
	err = v.Sub("executor").Unmarshal(executorConf)
	if err != nil {
		return err
	}
	// get the private key , if the private key does not exist, read it from 'keyPath'
	if executorConf.PrivateKey == "" {
		privateKeyBytes, err := file.ReadFile(executorConf.KeyPath, file.PrivateKeyFileName)
		if err == nil && len(privateKeyBytes) != 0 {
			executorConf.PrivateKey = strings.TrimSpace(string(privateKeyBytes))
		} else {
			return err
		}
	}
	return nil
}

// InitCliConfig parses client configuration file. if cli's configuration file is not existed, use executor's configuration file.
func InitCliConfig(configPath string) error {
	v := viper.New()
	v.SetConfigFile(configPath)
	if err := v.ReadInConfig(); err != nil {
		return err
	}
	innerV := v.Sub("blockchain")
	if innerV != nil {
		// If "blockchain" was existed, cli would use the configuration of cli.
		cliConf = new(ExecutorBlockchainConf)
		err := innerV.Unmarshal(cliConf)
		if err != nil {
			return err
		}
		return nil
	} else {
		// If "blockchain" wasn't existed, use the configuration of the executor.
		err := InitConfig(configPath)
		if err == nil {
			cliConf = executorConf.Blockchain
		}
		return err
	}
}

// GetExecutorConf returns all configuration of the executor
func GetExecutorConf() *ExecutorConf {
	return executorConf
}

// GetLogConf returns log configuration of the executor
func GetLogConf() *Log {
	return logConf
}

// GetCliConf returns blockchain configuration of the executor
func GetCliConf() *ExecutorBlockchainConf {
	return cliConf
}

