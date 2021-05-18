package claim

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"
)

// Claim represents a fact or a fake claim according to a certain publisher
type Claim struct {
	Title string
	PublisherName string
	// the url to the article evaluating the claim
	URL string
	// true if the claim is a fact, false if it is fake
	IsFact bool
	// the time at which this claim appeared 
	ReviewedAt time.Time
}

const invalidParamMsgFormat string = "Cannot create claim with invalid parameter %v: %v"

// NewClaim attempts to construct a claim based on the passed parameters. Returns an error on failure.
// The title's first letter is also capitalized if possible.
func NewClaim(title string, publisherName string, url string, isFact bool, reviewedAt time.Time) (Claim, error) {
	if title == "" {
		return Claim{}, fmt.Errorf(invalidParamMsgFormat, "title", title) 
	}
	if publisherName == "" {
		return Claim{}, fmt.Errorf(invalidParamMsgFormat, "publisherName", publisherName) 
	}
	if url == "" {
		return Claim{}, fmt.Errorf(invalidParamMsgFormat, "url", url) 
	}
	if reviewedAt.IsZero() {
		return Claim{}, fmt.Errorf(invalidParamMsgFormat, "reviewedAt", reviewedAt) 
	}
	
	title = capitalizeFirstLetter(title)
	constructedClaim := Claim{
		Title:          title,
		URL:           url,
		PublisherName: publisherName,
		IsFact:        isFact,
		ReviewedAt:    reviewedAt,
	}

	return constructedClaim, nil
}

// ReferencesVisuals returns true if the Claim's title contains a reference to some form of visual media
func (c Claim) ReferencesVisuals() bool {
	matches, _ := regexp.MatchString(`(?i)photo|video|image|picture`, c.Title)
	return matches
}

// Source enables the retrieval of Claims for a given publisher. The publisher could be a URL or domain depending on the implementation
type Source interface {
	GetClaims(publisher string) []Claim
}

func capitalizeFirstLetter(oldTitle string) string {
	var firstLetter rune
	for _, char := range []rune(oldTitle) {
		if unicode.IsLetter(char) {
			firstLetter = char
			break
		}
	}

	return strings.Replace(oldTitle, string(firstLetter), string(unicode.ToTitle(firstLetter)), 1)
}

// Sorter sorts claims from most recently reviewed to least recently reviewed
type Sorter struct {
	Claims []Claim
}

func (s Sorter) Len() int           { return len(s.Claims) }
func (s Sorter) Swap(i, j int)      { s.Claims[i], s.Claims[j] = s.Claims[j], s.Claims[i] }
func (s Sorter) Less(i, j int) bool { return s.Claims[i].ReviewedAt.After(s.Claims[j].ReviewedAt) }
