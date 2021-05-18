package claim

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

func Test_parseTextualRating(t *testing.T) {
	type args struct {
		textualRating string
	}
	tests := []struct {
		name          string
		textualRating string
		want          bool
		wantErr       bool
	}{
		{
			name:          "parsing 'true' returns true",
			textualRating: "true",
			want:          true,
			wantErr:       false,
		},
		{
			name:          "parsing 'real' returns true",
			textualRating: "real",
			want:          true,
			wantErr:       false,
		},

		{
			name:          "parsing 'false' returns false",
			textualRating: "false",
			want:          false,
			wantErr:       false,
		},
		{
			name:          "parsing 'fake' returns false",
			textualRating: "fake",
			want:          false,
			wantErr:       false,
		},
		{
			name:          "parsing 'not true' returns false",
			textualRating: "not true",
			want:          false,
			wantErr:       false,
		},
		{
			name:          "parsing 'not real' returns false",
			textualRating: "not real",
			want:          false,
			wantErr:       false,
		},
		{
			name:          "parsing without a match returns an error",
			textualRating: "STRING_THAT_SHOULD_NEVER_MATCH",
			want:          false,
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTextualRating(tt.textualRating)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTextualRating() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseTextualRating() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGoogleSource_GetClaims(t *testing.T) {
	tests := []struct {
		name      string
		mockedAPI mockedClaimAPI
		want      []Claim
	}{
		{
			name: "Correctly maps claim dto fields",
			mockedAPI: mockedClaimAPI{
				claimDtos: []claimDto{
					{Text: "article text", ClaimReview: []claimReviewDto{
						{
							Publisher:     publisherDto{Name: "publisher_name", Site: "http://publisher_site.com"},
							URL:           "http://article_url.com",
							TextualRating: "True",
							ReviewDate:    time.Unix(100, 0),
						},
					}},
				},
				err: nil,
			},
			want: []Claim{
				claim("article text", "publisher_name", "http://article_url.com", true, time.Unix(100, 0)),
			},
		},
		{
			name: "Excludes claim dtos if api error was encountered",
			mockedAPI: mockedClaimAPI{
				claimDtos: []claimDto{{Text: "Do not include this claim"}},
				err:       errors.New("Encountered error with claim API"),
			},
			want:      []Claim{},
		},
		{
			name: "Excludes claim dtos without any claim reviews",
			mockedAPI: mockedClaimAPI{
				claimDtos: []claimDto{{Text: "Do not include this claim", ClaimReview: []claimReviewDto{}}},
				err:       nil,
			},
			want:      []Claim{},
		},
		{
			name: "Excludes claim dtos with invalid textual rating",
			mockedAPI: mockedClaimAPI{
				claimDtos: []claimDto{
					{Text: "Do not include this claim",
						ClaimReview: []claimReviewDto{
							{TextualRating: "STRING_THAT_SHOULD_NEVER_MATCH_TEXTUAL_RATING"},
						},
					},
				},
				err: nil,
			},
			want:      []Claim{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			googleSource := GoogleSource{api: tt.mockedAPI}
			if got := googleSource.GetClaims("anypublisher.com"); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GoogleSource.GetClaims() = %v, want %v", got, tt.want)
			}
		})
	}
}

type mockedClaimAPI struct {
	// the claim dtos returned by the mock on request
	claimDtos []claimDto
	// the error returned by the mock on request
	err error
}

func (mock mockedClaimAPI) getClaims(publisher string) (claimResponseDto, error) {
	return claimResponseDto{mock.claimDtos}, mock.err
}

func claim(title string, publisherName string, url string, isFact bool, reviewedAt time.Time) Claim {
	c, err := NewClaim(title, publisherName, url, isFact, reviewedAt)
	if err != nil {
		panic("Claim constructed for test assertion is not valid: " + err.Error())
	}
	return c
}