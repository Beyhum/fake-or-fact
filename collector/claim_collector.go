package collector

import (
	"fake-or-fact/claim"
	"fake-or-fact/repo"
	"log"
	"sync"
)

type ClaimConfig struct {
	Database struct {
		Dialect          string
		ConnectionString string
	}
	GoogleFactCheckAPIKey     string
	GoogleFactCheckPublishers []string
	RealRssFeeds              []string
	FakeRssFeeds              []string
}

type ClaimCollector struct {
	r      repo.ClaimRepo
	config *ClaimConfig
}

func NewClaimCollector(r repo.ClaimRepo, config *ClaimConfig) ClaimCollector {
	return ClaimCollector{r, config}
}

func (collector ClaimCollector) CollectAndPersist() {
	config := collector.config

	googleSource := claim.NewGoogleSource(config.GoogleFactCheckAPIKey)
	googleClaimSink := goThroughClaims(googleSource, config.GoogleFactCheckPublishers)

	fakeRssSource := claim.NewRssSource(false)
	fakeRssClaimSink := goThroughClaims(fakeRssSource, config.FakeRssFeeds)

	realRssSource := claim.NewRssSource(true)
	realRssClaimSink := goThroughClaims(realRssSource, config.RealRssFeeds)

	aggregatedClaimChannel := aggregateClaimChannels(googleClaimSink, fakeRssClaimSink, realRssClaimSink)
	for claim := range aggregatedClaimChannel {

		e := collector.r.Save(claim)
		if e != nil && !repo.IsClaimExistsError(e) {
			log.Println(e)
		}
	}

}

func aggregateClaimChannels(chans ...<-chan claim.Claim) <-chan claim.Claim {
	var aggregateWaitGroup sync.WaitGroup
	aggregateChannel := make(chan claim.Claim)
	for _, c := range chans {
		aggregateWaitGroup.Add(1)
		go func(c <-chan claim.Claim) {
			for claim := range c {
				aggregateChannel <- claim
			}
			aggregateWaitGroup.Done()
		}(c)
	}
	go func() {
		aggregateWaitGroup.Wait()
		close(aggregateChannel)
	}()
	return aggregateChannel
}

// Asynchronously retrieves valid claims for the specified publishers and pushes them into the returned channel (sink).
// The channel WILL BE CLOSED once all claims have been collected.
// Only claims that could be correctly parsed and which do not reference any visuals are pushed into the channel.
func goThroughClaims(source claim.Source, publishers []string) <-chan claim.Claim {
	sink := make(chan claim.Claim)
	go func() {
		for _, publisher := range publishers {
			validClaims := source.GetClaims(publisher)
			facts := 0
			fakes := 0
			for _, c := range validClaims {
				if c.IsFact {
					facts++
				} else {
					fakes++
				}
				if !c.ReferencesVisuals() {
					sink <- c
				}
			}
			log.Printf("Publisher %v: %v facts and %v fakes", publisher, facts, fakes)
		}
		close(sink)
	}()
	return sink
}
