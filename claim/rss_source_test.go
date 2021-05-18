package claim

import (
	"reflect"
	"testing"
	"time"

	"github.com/mmcdole/gofeed"
)


func TestRssSource_GetClaims(t *testing.T) {
	
	tests := []struct {
		name      string
		rssSource *RssSource
		want      []Claim
	}{
		{
			name: "Correctly maps rss feed tags",
			rssSource: rssSourceWithFeed(`
			<?xml version="1.0" encoding="UTF-8"?>
			<feed xmlns="http://www.w3.org/2005/Atom" xmlns:media="http://search.yahoo.com/mrss/">
				<category term="publisher_name"/>
				<link href="http://publisher_site.com" />
				<entry>
					<link href="http://article_url.com" />
					<updated>2020-08-06T23:20:42+00:00</updated>
					<title>article title</title>
				</entry>
				<entry>
					<link href="http://second_article_url.com" />
					<updated>2021-10-03T05:00:15+00:00</updated>
					<title>second article title</title>
				</entry>
			</feed>
			`),
			want: []Claim{
				claim(
					"article title",
					"publisher_name",
					"http://article_url.com",
					true,
					time.Date(2020, 8, 6, 23, 20, 42, 0, time.UTC),
				),
				claim(
					"second article title",
					"publisher_name",
					"http://second_article_url.com",
					true,
					time.Date(2021, 10, 3, 5, 0, 15, 0, time.UTC),
				),
				
			},
		},
		{
			name: "Uses 'pubDate' for review date if 'updated' tag not available",
			rssSource: rssSourceWithFeed(`
			<rss xmlns:atom="http://www.w3.org/2005/Atom" xmlns:media="http://search.yahoo.com/mrss/" version="2.0">
				<channel>
					<link>http://publisher_site.com</link>
					<category>publisher_name</category>
					<item>
						<title>article title</title>
						<link>http://article_url.com</link>
						<pubDate>Sun, 02 Aug 2020 15:13:00 +0000</pubDate>
					</item>
				</channel>
			</rss>
			`),
			want: []Claim{
				claim(
					"article title",
					"publisher_name",
					"http://article_url.com",
					true,
					time.Date(2020, 8, 2, 15, 13, 0, 0, time.UTC),
				),
			},
		},
		{
			name: "Excludes items missing title, link, or review date tags",
			rssSource: rssSourceWithFeed(`
			<rss xmlns:atom="http://www.w3.org/2005/Atom" xmlns:media="http://search.yahoo.com/mrss/" version="2.0">
				<channel>
					<link>http://publisher_site.com</link>
					<category>publisher_name</category>
					<item>
						<link>http://article_missing_title.com</link>
						<pubDate>Sun, 02 Aug 2020 15:13:00 +0000</pubDate>
					</item>
					<item>
						<title>article missing link</title>
						<pubDate>Sun, 03 Aug 2020 15:13:00 +0000</pubDate>
					</item>
					<item>
						<title>article missing review date</title>
						<link>http://article_missing_review_date.com</link>
					</item>
				</channel>
			</rss>
			`),
			want: []Claim{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.rssSource.GetClaims("any publisher url"); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RssSource.GetClaims() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

// returns an RssSource which will read from a hardcoded feed as a string regardless of which publisher url is passed 
func rssSourceWithFeed(feed string) *RssSource {
	parseFromTextFunc := func(string) (*gofeed.Feed, error) {
		return gofeed.NewParser().ParseString(feed)
	}
	return &RssSource{true, parseFromTextFunc}
}