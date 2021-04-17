package integration_test

import (
	"encoding/json"
	"fmt"
	"kraicklist/config"
	"kraicklist/domain/model"
	"kraicklist/helper/jsons"
	"kraicklist/helper/logging"
	"math/rand"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type test struct {
	input    string
	expected string
}

type IntegrationTestSuite struct {
	suite.Suite
	host string

	positiveTestData map[string]test

	// TODO: add various test data
	negativeTestData map[string]test
	errorTestData    map[string]test
}

var positiveTestData = map[string]test{
	"search contains 1": {
		input:    "iphone",
		expected: "iphone",
	}, "search contains 2": {
		input:    "android",
		expected: "android",
	},
}

func (suite *IntegrationTestSuite) SetupTest() {
	conf := config.Get()
	logging.Init(conf.LogLevel)

	suite.host = "http://localhost:" + conf.Port
	suite.positiveTestData = positiveTestData
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (suite *IntegrationTestSuite) TestSeedAndSearch() {
	// data seeding by hitting index endpoint
	data := model.Advertisements{
		model.Advertisement{
			ID:      1,
			Title:   randomizeString(10),
			Content: randomizeString(100),
		},
		model.Advertisement{
			ID:      2,
			Title:   randomizeString(10),
			Content: randomizeString(100),
		},
	}
	result, err := suite.hitIndexDocs(data)
	assert.NoError(suite.T(), err, "should not error out")
	assert.Equal(suite.T(), result, "success")

	time.Sleep(3 * time.Second)

	// search query by hitting search endpoint
	for _, adData := range data {
		adsResult, err := suite.hitSearch(adData.Title)
		assert.NoError(suite.T(), err, "should not error out")
		assert.NotEmpty(suite.T(), adsResult, "result should not be empty")
		assert.Equal(suite.T(), adsResult[0].Title, adData.Title)
	}
}

func (suite *IntegrationTestSuite) TestPositiveCases() {
	for k, data := range suite.positiveTestData {
		suite.T().Log("Testing", k)
		adsResult, err := suite.hitSearch(data.input)
		assert.NoError(suite.T(), err, "should not error out")
		assert.NotEmpty(suite.T(), adsResult, "result should not be empty")
		assert.Contains(suite.T(), strings.ToUpper(adsResult[0].Title), strings.ToUpper(data.input))
	}
}

func (suite *IntegrationTestSuite) hitSearch(q string) (adsResult model.Advertisements, err error) {
	url := suite.host + "/api/advertisement/search"
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	queryParam := req.URL.Query()
	queryParam.Set("q", q)

	req.URL.RawQuery = queryParam.Encode()
	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()

	result := map[string]model.Advertisements{}
	if err = json.NewDecoder(res.Body).Decode(&result); err != nil {
		return
	}
	adsResult = result["data"]
	return
}

func (suite *IntegrationTestSuite) hitIndexDocs(adsData model.Advertisements) (data string, err error) {
	url := suite.host + "/api/advertisement/index"

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, strings.NewReader(jsons.ToStringJsonNoError(adsData)))
	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()
	result := map[string]string{}
	if err = json.NewDecoder(res.Body).Decode(&result); err != nil {
		return
	}
	data = result["data"]
	return
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randomizeString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
