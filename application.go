package main

import (
	"encoding/json"
	"fake-or-fact/claim"
	. "fake-or-fact/collector"
	"fake-or-fact/repo"
	"io/ioutil"
	"log"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

const GET_CLAIMS_PATH = "/api/claims"

func main() {

	dbConfig := loadConfig().Database
	db, _ := gorm.Open(dbConfig.Dialect, dbConfig.ConnectionString)
	defer db.Close()
	db.AutoMigrate(&repo.ClaimData{})
	repo := repo.NewClaimRepo(db)

	initializeCollector(repo)

	r := gin.Default()
	r.GET(GET_CLAIMS_PATH, GetClaimsRoute(repo))

	r.StaticFile("/", "./public/index.html")
	r.StaticFile("/index.html", "./public/index.html")
	r.StaticFile("/favicon.ico", "./public/favicon.ico")
	r.StaticFile("/manifest.json", "./public/manifest.json")
	r.StaticFile("/serviceworker.js", "./public/serviceworker.js")
	r.Static("/css", "./public/css")
	r.Static("/gif", "./public/gif")
	r.Static("/img", "./public/img")
	r.Run()
}

func GetClaimsRoute(repo repo.ClaimRepo) func(*gin.Context) {
	return func(c *gin.Context) {
		reviewedAt := BeforeQuery{}
		e := c.MustBindWith(&reviewedAt, binding.Query)
		if e == nil {
			if reviewedAt.Before.IsZero() {
				reviewedAt.Before = time.Now()
			}
			facts, _ := repo.Get(true, reviewedAt.Before)
			fakes, _ := repo.Get(false, reviewedAt.Before)

			claims := append(facts, fakes...)
			sort.Sort(claim.Sorter{Claims: claims})

			c.JSON(200, claims)
		}
	}
}

type BeforeQuery struct {
	Before time.Time `form:"before"`
}

func initializeCollector(r repo.ClaimRepo) {
	config := loadConfig()
	collector := NewClaimCollector(r, &config)
	collectorTicker := time.NewTicker(time.Duration(15) * time.Hour)
	go func() {
		for ; true; <-collectorTicker.C {
			collector.CollectAndPersist()
		}
	}()
}

func loadConfig() ClaimConfig {
	configBytes, err := ioutil.ReadFile("claim_config.json")
	if err != nil {
		log.Panicf("Failed to load claim_config.json: %v", err)
	}
	config := new(ClaimConfig)
	json.Unmarshal(configBytes, config)
	return *config
}
