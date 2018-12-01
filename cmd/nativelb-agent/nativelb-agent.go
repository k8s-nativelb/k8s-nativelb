/*
Copyright 2018 Sebastian Sch.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"github.com/k8s-nativelb/pkg/nativelb-agent"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	clusterName, isExist := os.LookupEnv("CLUSTER_NAME")
	if !isExist {
		panic(fmt.Errorf("CLUSTER_NAME environment variable doesn't exist"))
	}

	controlIP, isExist := os.LookupEnv("CONTROL_IP")
	if !isExist {
		panic(fmt.Errorf("CONTROL_INTERFACE environment variable doesn't exist"))
	}

	controlPort, isExist := os.LookupEnv("CONTROL_PORT")
	if !isExist {
		panic(fmt.Errorf("CONTROL_PORT environment variable doesn't exist"))
	}

	dataInterface, isExist := os.LookupEnv("DATA_INTERFACE")
	if !isExist || dataInterface == "" {
		dataInterface = ""
	}

	syncInterface, isExist := os.LookupEnv("SYNC_INTERFACE")
	if !isExist {
		syncInterface = ""
	}

	agent, err := nativelb_agent.NewNativeAgent(clusterName, controlIP, controlPort, dataInterface, syncInterface)
	if err != nil {
		panic(err)
	}

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)
	go agent.StopAgent(stopChan)

	err = agent.StartAgent()
	if err != nil {
		panic(err)
	}
}
