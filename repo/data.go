package repo

import (
	"fake-or-fact/claim"
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
)

type ClaimData struct {
	ID            uuid.UUID `gorm:"column:id;primary_key"`
	Title         string    `gorm:"column:title;type:varchar(500);not null"`
	PublisherName string    `gorm:"column:publisher_name;type:varchar(50);not null"`
	URL           string    `gorm:"column:url;type:varchar(500);unique;not null"`
	IsFact        bool      `gorm:"column:is_fact;not null;index:is_fact_and_reviewed_at_ix"`
	ReviewedAt    time.Time `gorm:"column:reviewed_at;not null;index:is_fact_and_reviewed_at_ix"`
}

const claimTableName string = "claim"

func (ClaimData) TableName() string {
	return claimTableName
}

type ClaimRepo interface {
	Save(claim claim.Claim) error
	Get(isFact bool, reviewedBefore time.Time) ([]claim.Claim, error)
}

type pgClaimRepo struct {
	db *gorm.DB
}

func NewClaimRepo(db *gorm.DB) ClaimRepo {
	return &pgClaimRepo{db}
}

// Save tries to persist a claim and returns an error if it fails.
// This might be due to a claim with the same URL already being stored.
func (repo *pgClaimRepo) Save(claim claim.Claim) error {
	claimData := asClaimData(claim)
	existingClaimWithSameURL := new(ClaimData)
	queryErr := repo.db.Where("url = ?", claimData.URL).First(existingClaimWithSameURL).Error
	if queryErr == nil {
		return claimExistsError{*existingClaimWithSameURL}
	} else {
		if gorm.IsRecordNotFoundError(queryErr) {
			return repo.db.Create(&claimData).Error
		} else {
			return queryErr

		}
	}
}

const pageLimit = 20

// Get returns a list of claims that are either real (isFact=true) or fake (isFact=false) that were reviewed at a time t < reviewedBefore.
// Claims are returned from latest to oldest, and are limited to 20 claims per request.
// An error is returned if an unexpected error is encountered while retrieving the claims.
func (repo *pgClaimRepo) Get(isFact bool, reviewedBefore time.Time) ([]claim.Claim, error) {
	foundClaimData := make([]ClaimData, 0, pageLimit)
	err := repo.db.Where("is_fact = ? AND reviewed_at < ?", isFact, reviewedBefore).Order("reviewed_at DESC").Limit(pageLimit).Find(&foundClaimData).Error
	if err != nil {
		return nil, err
	}
	mappedClaims := make([]claim.Claim, 0, pageLimit)
	for _, claimData := range foundClaimData {
		mappedClaims = append(mappedClaims, asClaim(claimData))
	}
	return mappedClaims, nil
}

// returns a new ClaimData based on a Claim.
func asClaimData(claim claim.Claim) ClaimData {
	return ClaimData{
		ID:            uuid.NewV4(),
		Title:         claim.Title,
		PublisherName: claim.PublisherName,
		URL:           claim.URL,
		IsFact:        claim.IsFact,
		ReviewedAt:    claim.ReviewedAt,
	}
}

// returns a new Claim based on a ClaimData.
func asClaim(claimData ClaimData) claim.Claim {
	return claim.Claim{
		Title:         claimData.Title,
		PublisherName: claimData.PublisherName,
		URL:           claimData.URL,
		IsFact:        claimData.IsFact,
		ReviewedAt:    claimData.ReviewedAt,
	}
}

// IsClaimExistsError returns true if the given error is of type claimExistsError
func IsClaimExistsError(err error) bool {
	_, isClaimExists := err.(claimExistsError)
	return isClaimExists
}

type claimExistsError struct {
	existingClaim ClaimData
}

func (e claimExistsError) Error() string {
	return fmt.Sprintf("A Claim already exists with the URL '%v': %#v", e.existingClaim.URL, e.existingClaim)
}

