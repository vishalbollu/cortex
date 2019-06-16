/*
Copyright 2019 Cortex Labs, Inc.

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

package cmd

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/cortexlabs/cortex/pkg/consts"
	"github.com/cortexlabs/cortex/pkg/lib/errors"
	"github.com/cortexlabs/cortex/pkg/lib/json"
	"github.com/cortexlabs/cortex/pkg/operator/api/schema"
)

var httpTransport = &http.Transport{
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}

var httpClient = &http.Client{
	Timeout:   time.Second * 20,
	Transport: httpTransport,
}

func makeRequest(request *http.Request) ([]byte, error) {
	request.Header.Set("Authorization", authHeader())
	request.Header.Set("CortexAPIVersion", consts.CortexVersion)

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		bodyBytes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, errors.Wrap(err, errStrRead)
		}

		var output schema.ErrorResponse
		err = json.Unmarshal(bodyBytes, &output)
		if err != nil || output.Error == "" {
			return nil, errors.New(strings.TrimSpace(string(bodyBytes)))
		}

		return nil, errors.New(output.Error)
	}

	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Wrap(err, errStrRead)
	}
	return bodyBytes, nil
}

func authHeader() string {
	cliConfig := getValidCLIConfig()
	return fmt.Sprintf("CortexAWS %s|%s", cliConfig.AWSAccessKeyID, cliConfig.AWSSecretAccessKey)
}
