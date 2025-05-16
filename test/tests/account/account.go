package accountTests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	errorHelpers "go-gin-test-job/src/common/error-helpers"
	"go-gin-test-job/src/config"
	"go-gin-test-job/src/database"
	"go-gin-test-job/src/database/entities"
	accountModuleDto "go-gin-test-job/src/modules/account/dto"
	numberUtil "go-gin-test-job/src/utils/number"
	orderUtil "go-gin-test-job/src/utils/order"
	"go-gin-test-job/test"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestAccountRoute(t *testing.T) {
	// GetAccounts
	validationGetAccountsTests(t)
	t.Run("TestGetAccountsRoute_SuccessNoParams", TestGetAccountsRoute_SuccessNoParams)
	t.Run("TestGetAccountsRoute_SuccessParamsOffsetAndCount", TestGetAccountsRoute_SuccessParamsOffsetAndCount)
	t.Run("TestGetAccountsRoute_SuccessParamsStatus", TestGetAccountsRoute_SuccessParamsStatus)
	t.Run("TestGetAccountsRoute_SuccessParamsOrderBy", TestGetAccountsRoute_SuccessParamsOrderBy)
	t.Run("TestGetAccountsRoute_SuccessParamsStatusAndOrderBy", TestGetAccountsRoute_SuccessParamsStatusAndOrderBy)
	t.Run("TestGetAccountsRoute_SuccessParamsOffsetAndCountAndStatusAndOrderBy", TestGetAccountsRoute_SuccessParamsOffsetAndCountAndStatusAndOrderBy)
	// CreateAccount
	validationCreateAccountTests(t)
	t.Run("TestCreateAccountRoute_FailAddressAlreadyExists", TestCreateAccountRoute_FailAddressAlreadyExists)
	t.Run("TestCreateAccountRoute_Success", TestCreateAccountRoute_Success)
}

func validationGetAccountsTests(t *testing.T) {
	validationTests := []struct {
		name         string
		params       accountModuleDto.GetAccountRequestDto
		expectedCode int
		expectedBody errorHelpers.ResponseBadRequestErrorHTTP
	}{
		{
			"FailInvalidOffsetMinValue",
			accountModuleDto.GetAccountRequestDto{Offset: -5},
			http.StatusBadRequest,
			errorHelpers.ResponseBadRequestErrorHTTP{Success: false, Message: "Offset must be greater than or equal 0"},
		},
		{
			"FailInvalidCountMinValue",
			accountModuleDto.GetAccountRequestDto{Count: -1},
			http.StatusBadRequest,
			errorHelpers.ResponseBadRequestErrorHTTP{Success: false, Message: "Count must be greater than or equal 1"},
		},
		{
			"FailInvalidCountMaxValue",
			accountModuleDto.GetAccountRequestDto{Count: 101},
			http.StatusBadRequest,
			errorHelpers.ResponseBadRequestErrorHTTP{Success: false, Message: "Count must be less than or equal 100"},
		},
		{
			"FailInvalidStatus",
			accountModuleDto.GetAccountRequestDto{Status: "invalid status"},
			http.StatusBadRequest,
			errorHelpers.ResponseBadRequestErrorHTTP{Success: false, Message: fmt.Sprintf("%s must be one of the next values: %s", "Status", strings.Join(entities.AccountStatusList, ","))},
		},
		{
			"FailInvalidOrderBy",
			accountModuleDto.GetAccountRequestDto{OrderBy: "invalid order by"},
			http.StatusBadRequest,
			errorHelpers.ResponseBadRequestErrorHTTP{Success: false, Message: "invalid order by parameter: invalid order by"},
		},
		{
			"FailInvalidOrderByMaxLength",
			accountModuleDto.GetAccountRequestDto{OrderBy: strings.Repeat("OrderBy", 255)},
			http.StatusBadRequest,
			errorHelpers.ResponseBadRequestErrorHTTP{Success: false, Message: "OrderBy must be shorter than or equal to 255 characters"},
		},
		{
			"FailInvalidSearchMaxLength",
			accountModuleDto.GetAccountRequestDto{Search: strings.Repeat("Search", 255)},
			http.StatusBadRequest,
			errorHelpers.ResponseBadRequestErrorHTTP{Success: false, Message: "Search must be shorter than or equal to 255 characters"},
		},
	}
	for _, validationTest := range validationTests {
		t.Run("TestGetAccountsRoute"+validationTest.name, func(t *testing.T) {
			type Params struct {
				Count   int                    `json:"count"`
				Offset  int                    `json:"offset"`
				Status  entities.AccountStatus `json:"status"`
				OrderBy string                 `json:"orderBy"`
				Search  string                 `json:"search"`
			}
			params := &Params{
				Count:   validationTest.params.Count,
				Offset:  validationTest.params.Offset,
				Status:  validationTest.params.Status,
				OrderBy: validationTest.params.OrderBy,
				Search:  validationTest.params.Search,
			}

			query := url.Values{}
			query.Add("count", numberUtil.IntToString(params.Count))
			query.Add("offset", numberUtil.IntToString(params.Offset))
			query.Add("status", string(params.Status))
			query.Add("orderBy", params.OrderBy)
			query.Add("search", params.Search)

			u := &url.URL{
				Path:     fmt.Sprintf("/account"),
				RawQuery: query.Encode(),
			}

			response := httptest.NewRecorder()
			request := httptest.NewRequest("GET", u.String(), nil)
			request.Header.Set("X-API-Key", config.AppConfig.AdminXApiKey)
			test.TestApp.ServeHTTP(response, request)
			assert.Equal(t, validationTest.expectedCode, response.Code)

			// Read the response body and parse JSON
			var responseDto errorHelpers.ResponseBadRequestErrorHTTP
			err := json.NewDecoder(response.Body).Decode(&responseDto)
			assert.Nil(t, err)

			assert.NotNil(t, responseDto.Success, "Success parameter should exist")
			assert.NotNil(t, responseDto.Message, "Message parameter should exist")

			assert.Equal(t, validationTest.expectedBody.Success, responseDto.Success)
			assert.Equal(t, validationTest.expectedBody.Message, responseDto.Message)
		})
	}
}

func TestGetAccountsRoute_SuccessNoParams(t *testing.T) {
	u := &url.URL{
		Path: fmt.Sprintf("/account"),
	}

	accounts, total := database.GetAccountsAndTotal("", make(map[string]string), accountModuleDto.DEFAULT_ACCOUNT_OFFSET, accountModuleDto.DEFAULT_ACCOUNT_COUNT, "")

	response := httptest.NewRecorder()
	request := httptest.NewRequest("GET", u.String(), nil)
	request.Header.Set("X-API-Key", config.AppConfig.AdminXApiKey)
	test.TestApp.ServeHTTP(response, request)
	assert.Equal(t, http.StatusOK, response.Code)

	// Read the response body and parse JSON
	var responseDto accountModuleDto.GetAccountResponseDto
	err := json.NewDecoder(response.Body).Decode(&responseDto)
	assert.Nil(t, err)

	assert.NotNil(t, responseDto.Offset, "Offset parameter should exist")
	assert.NotNil(t, responseDto.Count, "Count parameter should exist")
	assert.NotNil(t, responseDto.Total, "Total parameter should exist")
	assert.NotNil(t, responseDto.List, "List parameter should exist")

	assert.Equal(t, 0, responseDto.Offset)
	assert.Equal(t, accountModuleDto.DEFAULT_ACCOUNT_COUNT, responseDto.Count)
	assert.Equal(t, total, responseDto.Total)
	assert.Equal(t, len(accounts), len(responseDto.List))

	for _, accountDto := range responseDto.List {
		var matchingAccount *entities.Account
		for _, account := range accounts {
			if account.Id == accountDto.Id {
				matchingAccount = account
				break
			}
		}
		assert.NotNil(t, matchingAccount)
		test.CompareAccount(t, matchingAccount, accountDto)
	}
}

func TestGetAccountsRoute_SuccessParamsOffsetAndCount(t *testing.T) {
	type Params struct {
		Count  int `json:"count"`
		Offset int `json:"offset"`
	}
	params := &Params{
		Count:  2,
		Offset: 1,
	}

	query := url.Values{}
	query.Add("count", numberUtil.IntToString(params.Count))
	query.Add("offset", numberUtil.IntToString(params.Offset))

	u := &url.URL{
		Path:     fmt.Sprintf("/account"),
		RawQuery: query.Encode(),
	}

	accounts, total := database.GetAccountsAndTotal("", make(map[string]string), params.Offset, params.Count, "")

	response := httptest.NewRecorder()
	request := httptest.NewRequest("GET", u.String(), nil)
	request.Header.Set("X-API-Key", config.AppConfig.AdminXApiKey)
	test.TestApp.ServeHTTP(response, request)
	assert.Equal(t, http.StatusOK, response.Code)

	// Read the response body and parse JSON
	var responseDto accountModuleDto.GetAccountResponseDto
	err := json.NewDecoder(response.Body).Decode(&responseDto)
	assert.Nil(t, err)

	assert.NotNil(t, responseDto.Offset, "Offset parameter should exist")
	assert.NotNil(t, responseDto.Count, "Count parameter should exist")
	assert.NotNil(t, responseDto.Total, "Total parameter should exist")
	assert.NotNil(t, responseDto.List, "List parameter should exist")

	assert.Equal(t, params.Offset, responseDto.Offset)
	assert.Equal(t, params.Count, responseDto.Count)
	assert.Equal(t, total, responseDto.Total)
	assert.Equal(t, len(accounts), len(responseDto.List))

	for _, accountDto := range responseDto.List {
		var matchingAccount *entities.Account
		for _, account := range accounts {
			if account.Id == accountDto.Id {
				matchingAccount = account
				break
			}
		}
		assert.NotNil(t, matchingAccount)
		test.CompareAccount(t, matchingAccount, accountDto)
	}
}

func TestGetAccountsRoute_SuccessParamsStatus(t *testing.T) {
	type Params struct {
		Status entities.AccountStatus `json:"status"`
	}
	params := &Params{
		Status: entities.AccountStatusOn,
	}

	query := url.Values{}
	query.Add("status", string(params.Status))

	u := &url.URL{
		Path:     fmt.Sprintf("/account"),
		RawQuery: query.Encode(),
	}

	accounts, total := database.GetAccountsAndTotal(params.Status, make(map[string]string), accountModuleDto.DEFAULT_ACCOUNT_OFFSET, accountModuleDto.DEFAULT_ACCOUNT_COUNT, "")

	response := httptest.NewRecorder()
	request := httptest.NewRequest("GET", u.String(), nil)
	request.Header.Set("X-API-Key", config.AppConfig.AdminXApiKey)
	test.TestApp.ServeHTTP(response, request)
	assert.Equal(t, http.StatusOK, response.Code)

	// Read the response body and parse JSON
	var responseDto accountModuleDto.GetAccountResponseDto
	err := json.NewDecoder(response.Body).Decode(&responseDto)
	assert.Nil(t, err)

	assert.NotNil(t, responseDto.Offset, "Offset parameter should exist")
	assert.NotNil(t, responseDto.Count, "Count parameter should exist")
	assert.NotNil(t, responseDto.Total, "Total parameter should exist")
	assert.NotNil(t, responseDto.List, "List parameter should exist")

	assert.Equal(t, 0, responseDto.Offset)
	assert.Equal(t, accountModuleDto.DEFAULT_ACCOUNT_COUNT, responseDto.Count)
	assert.Equal(t, total, responseDto.Total)
	assert.Equal(t, len(accounts), len(responseDto.List))

	for _, accountDto := range responseDto.List {
		var matchingAccount *entities.Account
		for _, account := range accounts {
			if account.Id == accountDto.Id {
				matchingAccount = account
				break
			}
		}
		assert.NotNil(t, matchingAccount)
		test.CompareAccount(t, matchingAccount, accountDto)
	}
}

func TestGetAccountsRoute_SuccessParamsOrderBy(t *testing.T) {
	type Params struct {
		OrderBy string `json:"orderBy"`
	}
	params := &Params{
		OrderBy: "id ASC",
	}

	query := url.Values{}
	query.Add("orderBy", params.OrderBy)

	u := &url.URL{
		Path:     fmt.Sprintf("/account"),
		RawQuery: query.Encode(),
	}

	orderParams, err := orderUtil.GetOrderByParamsSecure(nil, params.OrderBy, ",", accountModuleDto.GetAvailableAccountSortFieldList)
	accounts, total := database.GetAccountsAndTotal("", orderParams, accountModuleDto.DEFAULT_ACCOUNT_OFFSET, accountModuleDto.DEFAULT_ACCOUNT_COUNT, "")

	response := httptest.NewRecorder()
	request := httptest.NewRequest("GET", u.String(), nil)
	request.Header.Set("X-API-Key", config.AppConfig.AdminXApiKey)
	test.TestApp.ServeHTTP(response, request)
	assert.Equal(t, http.StatusOK, response.Code)

	// Read the response body and parse JSON
	var responseDto accountModuleDto.GetAccountResponseDto
	err = json.NewDecoder(response.Body).Decode(&responseDto)
	assert.Nil(t, err)

	assert.NotNil(t, responseDto.Offset, "Offset parameter should exist")
	assert.NotNil(t, responseDto.Count, "Count parameter should exist")
	assert.NotNil(t, responseDto.Total, "Total parameter should exist")
	assert.NotNil(t, responseDto.List, "List parameter should exist")

	assert.Equal(t, 0, responseDto.Offset)
	assert.Equal(t, accountModuleDto.DEFAULT_ACCOUNT_COUNT, responseDto.Count)
	assert.Equal(t, total, responseDto.Total)
	assert.Equal(t, len(accounts), len(responseDto.List))

	for _, accountDto := range responseDto.List {
		var matchingAccount *entities.Account
		for _, account := range accounts {
			if account.Id == accountDto.Id {
				matchingAccount = account
				break
			}
		}
		assert.NotNil(t, matchingAccount)
		test.CompareAccount(t, matchingAccount, accountDto)
	}
}

func TestGetAccountsRoute_SuccessParamsStatusAndOrderBy(t *testing.T) {
	type Params struct {
		Status  entities.AccountStatus `json:"status"`
		OrderBy string                 `json:"orderBy"`
	}
	params := &Params{
		Status:  entities.AccountStatusOn,
		OrderBy: "id ASC",
	}

	query := url.Values{}
	query.Add("status", string(params.Status))
	query.Add("orderBy", params.OrderBy)

	u := &url.URL{
		Path:     fmt.Sprintf("/account"),
		RawQuery: query.Encode(),
	}

	orderParams, err := orderUtil.GetOrderByParamsSecure(nil, params.OrderBy, ",", accountModuleDto.GetAvailableAccountSortFieldList)
	accounts, total := database.GetAccountsAndTotal(params.Status, orderParams, accountModuleDto.DEFAULT_ACCOUNT_OFFSET, accountModuleDto.DEFAULT_ACCOUNT_COUNT, "")

	response := httptest.NewRecorder()
	request := httptest.NewRequest("GET", u.String(), nil)
	request.Header.Set("X-API-Key", config.AppConfig.AdminXApiKey)
	test.TestApp.ServeHTTP(response, request)
	assert.Equal(t, http.StatusOK, response.Code)

	// Read the response body and parse JSON
	var responseDto accountModuleDto.GetAccountResponseDto
	err = json.NewDecoder(response.Body).Decode(&responseDto)
	assert.Nil(t, err)

	assert.NotNil(t, responseDto.Offset, "Offset parameter should exist")
	assert.NotNil(t, responseDto.Count, "Count parameter should exist")
	assert.NotNil(t, responseDto.Total, "Total parameter should exist")
	assert.NotNil(t, responseDto.List, "List parameter should exist")

	assert.Equal(t, 0, responseDto.Offset)
	assert.Equal(t, accountModuleDto.DEFAULT_ACCOUNT_COUNT, responseDto.Count)
	assert.Equal(t, total, responseDto.Total)
	assert.Equal(t, len(accounts), len(responseDto.List))

	for _, accountDto := range responseDto.List {
		var matchingAccount *entities.Account
		for _, account := range accounts {
			if account.Id == accountDto.Id {
				matchingAccount = account
				break
			}
		}
		assert.NotNil(t, matchingAccount)
		test.CompareAccount(t, matchingAccount, accountDto)
	}
}

func TestGetAccountsRoute_SuccessParamsOffsetAndCountAndStatusAndOrderBy(t *testing.T) {
	type Params struct {
		Count   int                    `json:"count"`
		Offset  int                    `json:"offset"`
		Status  entities.AccountStatus `json:"status"`
		OrderBy string                 `json:"orderBy"`
	}
	params := &Params{
		Count:   2,
		Offset:  1,
		Status:  entities.AccountStatusOn,
		OrderBy: "id ASC",
	}

	query := url.Values{}
	query.Add("count", numberUtil.IntToString(params.Count))
	query.Add("offset", numberUtil.IntToString(params.Offset))
	query.Add("status", string(params.Status))
	query.Add("orderBy", params.OrderBy)

	u := &url.URL{
		Path:     fmt.Sprintf("/account"),
		RawQuery: query.Encode(),
	}

	orderParams, err := orderUtil.GetOrderByParamsSecure(nil, params.OrderBy, ",", accountModuleDto.GetAvailableAccountSortFieldList)
	accounts, total := database.GetAccountsAndTotal(params.Status, orderParams, params.Offset, params.Count, "")

	response := httptest.NewRecorder()
	request := httptest.NewRequest("GET", u.String(), nil)
	request.Header.Set("X-API-Key", config.AppConfig.AdminXApiKey)
	test.TestApp.ServeHTTP(response, request)
	assert.Equal(t, http.StatusOK, response.Code)

	// Read the response body and parse JSON
	var responseDto accountModuleDto.GetAccountResponseDto
	err = json.NewDecoder(response.Body).Decode(&responseDto)
	assert.Nil(t, err)

	assert.NotNil(t, responseDto.Offset, "Offset parameter should exist")
	assert.NotNil(t, responseDto.Count, "Count parameter should exist")
	assert.NotNil(t, responseDto.Total, "Total parameter should exist")
	assert.NotNil(t, responseDto.List, "List parameter should exist")

	assert.Equal(t, params.Offset, responseDto.Offset)
	assert.Equal(t, params.Count, responseDto.Count)
	assert.Equal(t, total, responseDto.Total)
	assert.Equal(t, len(accounts), len(responseDto.List))

	for _, accountDto := range responseDto.List {
		var matchingAccount *entities.Account
		for _, account := range accounts {
			if account.Id == accountDto.Id {
				matchingAccount = account
				break
			}
		}
		assert.NotNil(t, matchingAccount)
		test.CompareAccount(t, matchingAccount, accountDto)
	}
}

func validationCreateAccountTests(t *testing.T) {
	validationTests := []struct {
		name         string
		jsonParams   string
		expectedCode int
		expectedBody errorHelpers.ResponseBadRequestErrorHTTP
	}{
		{
			"FailInvalidPayload",
			`{ "invalid_field": "wrong value" }`,
			http.StatusBadRequest,
			errorHelpers.ResponseBadRequestErrorHTTP{Success: false, Message: "Address format is wrong"},
		},
		{
			"FailInvalidAccountAddress",
			`{"address": "wrong address", "name": "Test Account", "rank": 50, "status": "On"}`,
			http.StatusBadRequest,
			errorHelpers.ResponseBadRequestErrorHTTP{Success: false, Message: "Address format is wrong"},
		},
		{
			"FailInvalidAccountStatus",
			`{"address": "1JzfdUygUFk2M6KS3ngFMGRsy5vsH4N37a", "name": "Test Account", "rank": 50, "status": "invalid status"}`,
			http.StatusBadRequest,
			errorHelpers.ResponseBadRequestErrorHTTP{Success: false, Message: fmt.Sprintf("%s must be one of the next values: %s", "Status", strings.Join(entities.AccountStatusList, ","))},
		},
		{
			"FailMissingName",
			`{"address": "1JzfdUygUFk2M6KS3ngFMGRsy5vsH4N37a", "rank": 50, "status": "On"}`,
			http.StatusBadRequest,
			errorHelpers.ResponseBadRequestErrorHTTP{Success: false, Message: "Name is required"},
		},
		{
			"FailMissingRank",
			`{"address": "1JzfdUygUFk2M6KS3ngFMGRsy5vsH4N37a", "name": "Test Account", "status": "On"}`,
			http.StatusBadRequest,
			errorHelpers.ResponseBadRequestErrorHTTP{Success: false, Message: "Rank is required"},
		},
		{
			"FailInvalidRank",
			`{"address": "1JzfdUygUFk2M6KS3ngFMGRsy5vsH4N37a", "name": "Test Account", "rank": 150, "status": "On"}`,
			http.StatusBadRequest,
			errorHelpers.ResponseBadRequestErrorHTTP{Success: false, Message: "Invalid request query"},
		},
	}
	for _, validationTest := range validationTests {
		t.Run("TestCreateAccountRoute"+validationTest.name, func(t *testing.T) {
			type Params struct {
				Address string                 `json:"address"`
				Name    string                 `json:"name"`
				Rank    int8                   `json:"rank"`
				Memo    *string                `json:"memo"`
				Status  entities.AccountStatus `json:"status"`
			}

			response := httptest.NewRecorder()
			request := httptest.NewRequest("POST", "/account", bytes.NewBufferString(validationTest.jsonParams))
			request.Header.Set("Content-Type", "application/json")
			request.Header.Set("X-API-Key", config.AppConfig.AdminXApiKey)
			test.TestApp.ServeHTTP(response, request)
			assert.Equal(t, validationTest.expectedCode, response.Code)

			// Read the response body and parse JSON
			var responseDto errorHelpers.ResponseBadRequestErrorHTTP
			err := json.NewDecoder(response.Body).Decode(&responseDto)
			assert.Nil(t, err)

			assert.NotNil(t, responseDto.Success, "Success parameter should exist")
			assert.NotNil(t, responseDto.Message, "Message parameter should exist")

			assert.Equal(t, validationTest.expectedBody.Success, responseDto.Success)
			assert.Equal(t, validationTest.expectedBody.Message, responseDto.Message)
		})
	}
}

func TestCreateAccountRoute_FailAddressAlreadyExists(t *testing.T) {
	// Clean up any existing test accounts
	database.DbConn.Where("address = ?", "1JzfdUygUFk2M6KS3ngFMGRsy5vsH4N37a").Delete(&entities.Account{})
	
	// Create an initial account
	account := entities.Account{
		Address: "1JzfdUygUFk2M6KS3ngFMGRsy5vsH4N37a",
		Name:    "Test Account",
		Rank:    50,
		Status:  entities.AccountStatusOn,
	}
	
	// Insert the account into the database
	result := database.DbConn.Create(&account)
	assert.Nil(t, result.Error)
	
	// Try to create an account with the same address
	type Params struct {
		Address string                 `json:"address"`
		Name    string                 `json:"name"`
		Rank    int8                   `json:"rank"`
		Memo    *string                `json:"memo"`
		Status  entities.AccountStatus `json:"status"`
	}
	memo := "Test memo"
	params := &Params{
		Address: "1JzfdUygUFk2M6KS3ngFMGRsy5vsH4N37a",
		Name:    "New Account",
		Rank:    75,
		Memo:    &memo,
		Status:  entities.AccountStatusOn,
	}

	body, _ := json.Marshal(params)
	response := httptest.NewRecorder()
	request := httptest.NewRequest("POST", "/account", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-API-Key", config.AppConfig.AdminXApiKey)
	test.TestApp.ServeHTTP(response, request)
	assert.Equal(t, http.StatusConflict, response.Code)

	// Read the response body and parse JSON
	var responseDto errorHelpers.ResponseConflictErrorHTTP
	err := json.NewDecoder(response.Body).Decode(&responseDto)
	assert.Nil(t, err)

	assert.NotNil(t, responseDto.Success, "Success parameter should exist")
	assert.NotNil(t, responseDto.Message, "Message parameter should exist")

	assert.Equal(t, false, responseDto.Success)
	assert.Equal(t, "Address already exists", responseDto.Message)
	
	// Clean up
	database.DbConn.Delete(&account)
}

func TestCreateAccountRoute_Success(t *testing.T) {
	// Clean up any existing test accounts
	database.DbConn.Where("address = ?", "1JzfdUygUFk2M6KS3ngFMGRsy5vsH4N37a").Delete(&entities.Account{})
	
	type Params struct {
		Address string                 `json:"address"`
		Name    string                 `json:"name"`
		Rank    int8                   `json:"rank"`
		Memo    *string                `json:"memo"`
		Status  entities.AccountStatus `json:"status"`
	}
	memo := "Important account for transactions"
	params := &Params{
		Address: "1JzfdUygUFk2M6KS3ngFMGRsy5vsH4N37a",
		Name:    "Main Account",
		Rank:    50,
		Memo:    &memo,
		Status:  entities.AccountStatusOn,
	}

	body, _ := json.Marshal(params)
	response := httptest.NewRecorder()
	request := httptest.NewRequest("POST", "/account", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-API-Key", config.AppConfig.AdminXApiKey)
	test.TestApp.ServeHTTP(response, request)
	assert.Equal(t, http.StatusOK, response.Code)

	// Read the response body and parse JSON
	var responseDto accountModuleDto.AccountDto
	err := json.NewDecoder(response.Body).Decode(&responseDto)
	assert.Nil(t, err)

	assert.NotNil(t, responseDto.Id, "Id parameter should exist")
	assert.NotNil(t, responseDto.Address, "Address parameter should exist")
	assert.NotNil(t, responseDto.Name, "Name parameter should exist")
	assert.NotNil(t, responseDto.Rank, "Rank parameter should exist")
	assert.NotNil(t, responseDto.Memo, "Memo parameter should exist")
	assert.NotNil(t, responseDto.Balance, "Balance parameter should exist")
	assert.NotNil(t, responseDto.Status, "Status parameter should exist")
	assert.NotNil(t, responseDto.CreatedAt, "CreatedAt parameter should exist")
	assert.NotNil(t, responseDto.UpdatedAt, "UpdatedAt parameter should exist")

	assert.Greater(t, responseDto.Id, int64(0))
	assert.Equal(t, params.Address, responseDto.Address)
	assert.Equal(t, params.Name, responseDto.Name)
	assert.Equal(t, params.Rank, responseDto.Rank)
	assert.Equal(t, *params.Memo, *responseDto.Memo)
	assert.Equal(t, "0", responseDto.Balance)
	assert.Equal(t, string(params.Status), responseDto.Status)
	
	// Clean up
	database.DbConn.Where("address = ?", params.Address).Delete(&entities.Account{})
}
