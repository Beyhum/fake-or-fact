package claim

import (
	"reflect"
	"testing"
	"time"
)

func Test_NewClaim(t *testing.T) {
	type args struct {
		title          string
		publisherName string
		url           string
		isFact        bool
		reviewedAt    time.Time
	}
	tests := []struct {
		name    string
		args    args
		want    Claim
		wantErr bool
	}{
		{
			name: "Empty title returns error",
			args: args{ title: "", publisherName: "p", url: "u", isFact: true, reviewedAt: time.Now() },
			wantErr: true,
		},
		{
			name: "Empty publisherName returns error",
			args: args{ publisherName: "", title: "t", url: "u", isFact: true, reviewedAt: time.Now() },
			wantErr: true,
		},
		{
			name: "Empty url returns error",
			args: args{ url: "", title: "t", publisherName: "p", isFact: true, reviewedAt: time.Now() },
			wantErr: true,
		},
		{
			name: "Zero time returns error",
			args: args{reviewedAt: time.Time{}, title: "t", publisherName: "p", url: "u", isFact: true },
			wantErr: true,
		},
		{
			name: "Capitalizes title and maps fields properly",
			args: args{
				title: "'first letter of title is capitalized.'",
				publisherName: "Publisher Name",
				url: "http://url.com",
				isFact: true,
				reviewedAt: time.Date(2020, 8, 6, 23, 20, 42, 0, time.UTC),
			},
			wantErr: false,
			want: Claim {
				Title: "'First letter of title is capitalized.'",
				PublisherName: "Publisher Name",
				URL: "http://url.com",
				IsFact: true,
				ReviewedAt: time.Date(2020, 8, 6, 23, 20, 42, 0, time.UTC),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewClaim(tt.args.title, tt.args.publisherName, tt.args.url, tt.args.isFact, tt.args.reviewedAt)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClaim() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewClaim() = %v, want %v", got, tt.want)
			}
		})
	}
}
