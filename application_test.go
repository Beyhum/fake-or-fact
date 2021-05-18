package main

import (
	"encoding/json"
	"fake-or-fact/claim"
	"fake-or-fact/repo"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

var TIME_11_AM = time.Date(2020, 8, 6, 11, 0, 0, 0, time.UTC)
var TRUE_AT_11, _ = claim.NewClaim("A", "ABC", "http://abc.com", true, TIME_11_AM)
var TRUE_AT_12, _ = claim.NewClaim("B", "BCD", "http://bcd.com", true, TIME_11_AM.Add(time.Hour))
var FAKE_AT_10, _ = claim.NewClaim("C", "CXY", "http://cxy.com", false, TIME_11_AM.Add(-time.Hour))
var FAKE_AT_13, _ = claim.NewClaim("D", "DDD", "http://ddd.com", false, TIME_11_AM.Add(2 * time.Hour))

func Test_GetClaimsRoute(t *testing.T) {
	tests := []struct {
		name string
		requestUrl string
		expectedStatusCode int
		expectedClaims []claim.Claim

	}{
		{
			name: "Retrieves all claims before date",
			requestUrl: "/api/claims?before=" + TIME_11_AM.Add(time.Second).Format(time.RFC3339),
			expectedStatusCode: 200,
			expectedClaims: []claim.Claim {
				TRUE_AT_11,
				FAKE_AT_10,
			},
		},
		{
			name: "Orders claims from latest to oldest",
			requestUrl: "/api/claims?before=" + TIME_11_AM.Add(24 * time.Hour).Format(time.RFC3339),
			expectedStatusCode: 200,
			expectedClaims: []claim.Claim {
				FAKE_AT_13,
				TRUE_AT_12,
				TRUE_AT_11,
				FAKE_AT_10,
			},
		},
		{
			name: "'before' query param defaults to current time",
			requestUrl: "/api/claims",
			expectedStatusCode: 200,
			expectedClaims: []claim.Claim {
				FAKE_AT_13,
				TRUE_AT_12,
				TRUE_AT_11,
				FAKE_AT_10,
			},
		},
	}

	mock := &mockRepo{
		trueClaims: []claim.Claim{TRUE_AT_11, TRUE_AT_12},
		fakeClaims: []claim.Claim{FAKE_AT_10, FAKE_AT_13},
	}
	router := setupRouter(mock)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		response := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", tt.requestUrl, nil)
		router.ServeHTTP(response, req)
		
		actualStatusCode := response.Result().StatusCode
		actualClaims := []claim.Claim{}
		if actualStatusCode != 400 {
			json.Unmarshal(response.Body.Bytes(), &actualClaims)
		}
		if actualStatusCode != tt.expectedStatusCode {
			t.Errorf("HTTP Response Code = %#v, want %#v", actualStatusCode, tt.expectedStatusCode)
		}
		
		if tt.expectedStatusCode != 400 && !reflect.DeepEqual(actualClaims, tt.expectedClaims) {
			t.Errorf("Returned Claims = %#v, want %#v", actualClaims, tt.expectedClaims)
		}
		})
	}
}


func setupRouter(repo repo.ClaimRepo) *gin.Engine {
	router := gin.Default()
	router.GET(GET_CLAIMS_PATH, GetClaimsRoute(repo))
	return router
}

type mockRepo struct {
	trueClaims []claim.Claim
	fakeClaims []claim.Claim
}

func (mock *mockRepo) Save(claim claim.Claim) error {
	if claim.IsFact {
		mock.trueClaims = append(mock.trueClaims, claim) 
	} else {
		mock.fakeClaims = append(mock.fakeClaims, claim) 
	}
	return nil
}

func (mock *mockRepo) Get(isFact bool, reviewedBefore time.Time) ([]claim.Claim, error) {
	toReturn := []claim.Claim{}
	claimsToIterateOver := []claim.Claim{}
	if isFact {
		claimsToIterateOver = mock.trueClaims
	} else {
		claimsToIterateOver = mock.fakeClaims
	}

	for _, claim := range claimsToIterateOver {
		if claim.ReviewedAt.Before(reviewedBefore) {
			toReturn = append(toReturn, claim)
		}
	}
	return toReturn, nil
}