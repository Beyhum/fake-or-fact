package claim

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"time"
)

// GoogleSource is a claim source based on the google fact-check API
type GoogleSource struct {
	api claimAPI
}

// NewGoogleSource creates a claim source based on the google fact-check API
func NewGoogleSource(apiKey string) *GoogleSource {
	return &GoogleSource{newClaimAPI(apiKey)}
}

// GetClaims returns a slice of Claims that could be parsed for the given publisher
func (googleSource *GoogleSource) GetClaims(publisher string) []Claim {
	claims := make([]Claim, 0)
	claimResponse, apiErr := googleSource.api.getClaims(publisher)
	if apiErr != nil {
		log.Printf("Encountered error while collecting claims for publisher %s: %s", publisher, apiErr.Error())
	} else {
		for _, claimDto := range claimResponse.Claims {
			if len(claimDto.ClaimReview) > 0 {
				claimReview := claimDto.ClaimReview[0]
				isFact, err := parseTextualRating(claimReview.TextualRating)
				if err == nil {

					claim, creationErr := NewClaim(claimDto.Text, claimReview.Publisher.Name, claimReview.URL, isFact, claimReview.ReviewDate)
					if creationErr == nil {
						claims = append(claims, claim)
					} else {
						log.Printf("%v: %v", claimReview.Publisher.Name, creationErr.Error())
					}
				} else {
					log.Printf("%v: %v", claimReview.Publisher.Name, err.Error())
				}
			}
		}
	}
	return claims
}

const matchTrueRegex string = `(?i)^([^n]|n[^o]|no[^t])*(true|real)`
const trueRating bool = true
const matchFalseRegex string = `(?i)false|fake|not.(true|real)`
const falseRating bool = false

func parseTextualRating(textualRating string) (bool, error) {
	matchesTrue, matchTrueRegexErr := regexp.MatchString(matchTrueRegex, textualRating)
	if matchTrueRegexErr != nil {
		return false, matchTrueRegexErr
	}
	if matchesTrue {
		return trueRating, nil
	}
	matchesFalse, matchFalseRegexErr := regexp.MatchString(matchFalseRegex, textualRating)
	if matchFalseRegexErr != nil {
		return false, matchFalseRegexErr
	}
	if matchesFalse {
		return falseRating, nil
	}
	return false, fmt.Errorf("Could not find any match for textual rating %s", textualRating)
}

// claimAPI allows us to retrieve to retrieve claims based on an http endpoint
type claimAPI interface {
	getClaims(publisher string) (claimResponseDto, error)
}

const apiURL string = "https://factchecktools.googleapis.com/v1alpha1/claims:search?languageCode=en-US&pageSize=100&maxAgeDays=20"

var claimResponseOnError claimResponseDto = claimResponseDto{}

// googleClaimAPI is the real implementation of the claimAPI
type googleClaimAPI struct {
	apiKey     string
	httpClient http.Client
}

func newClaimAPI(apiKey string) claimAPI {
	return &googleClaimAPI{apiKey: apiKey, httpClient: http.Client{}}
}

// GetClaims returns a slice of ClaimResponseDto for a given news publisher as per the google claim API
func (api *googleClaimAPI) getClaims(publisher string) (claimResponseDto, error) {
	resp, httpError := api.httpClient.Get(apiURL + "&reviewPublisherSiteFilter=" + publisher + "&key=" + api.apiKey)
	if httpError != nil {
		return claimResponseOnError, httpError
	}

	if isFailure(resp) {
		return claimResponseOnError, fmt.Errorf("Encountered invalid response code while querying claims API: %v", resp.StatusCode)
	}

	defer resp.Body.Close()
	responseBody, parseError := ioutil.ReadAll(resp.Body)
	if parseError != nil {
		return claimResponseOnError, parseError
	}
	claimResponse := new(claimResponseDto)
	serializationErr := json.Unmarshal(responseBody, claimResponse)
	if serializationErr != nil {
		return claimResponseOnError, serializationErr
	}
	return *claimResponse, nil
}

// DTO returned by the google claim API
type claimResponseDto struct {
	Claims []claimDto
}
type claimDto struct {
	Text        string
	ClaimReview []claimReviewDto
}
type claimReviewDto struct {
	Publisher     publisherDto
	URL           string
	TextualRating string
	ReviewDate    time.Time
}

type publisherDto struct {
	Name string
	Site string
}

func isFailure(httpResponse *http.Response) bool {
	return httpResponse.StatusCode >= 400 && httpResponse.StatusCode < 600
}
