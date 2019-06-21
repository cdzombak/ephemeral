package main

import (
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ChimeraCoder/anaconda"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/crewjam/errset"
	"github.com/pkg/errors"
)

var (
	consumerKey       = getenv("TWITTER_CONSUMER_KEY")
	consumerSecret    = getenv("TWITTER_CONSUMER_SECRET")
	accessToken       = getenv("TWITTER_ACCESS_TOKEN")
	accessTokenSecret = getenv("TWITTER_ACCESS_TOKEN_SECRET")
	maxTweetAge       = getenv("MAX_TWEET_AGE")
	whitelist         = getWhitelist()
)

// MyResponse for AWS SAM
type MyResponse struct {
	StatusCode string `json:"StatusCode"`
	Message    string `json:"Body"`
}

// getenv returns the value of a required environment variable, or panics if the variable is empty
func getenv(name string) string {
	v := os.Getenv(name)
	if v == "" {
		panic("missing required environment variable " + name)
	}
	return v
}

// getWhitelist returns the contents of the WHITELIST environment variable, split by colons
func getWhitelist() []string {
	v := os.Getenv("WHITELIST")
	if v == "" {
		return nil
	}
	return strings.Split(v, ":")
}

// getTimeline returns the latest 200 tweets from the timeline, or an error
func getTimeline(api *anaconda.TwitterApi) ([]anaconda.Tweet, error) {
	args := url.Values{}
	args.Add("count", "200")        // Twitter only returns most recent 20 tweets by default, so override
	args.Add("include_rts", "true") // When using count argument, RTs are excluded, so include them as recommended
	timeline, err := api.GetUserTimeline(args)
	if err != nil {
		return nil, err
	}
	return timeline, nil
}

// isWhitelisted checks the global whitelist (from the WHITELIST env var) for the given tweet id
func isWhitelisted(id int64) bool {
	tweetID := strconv.FormatInt(id, 10)
	for _, w := range whitelist {
		if w == tweetID {
			return true
		}
	}
	return false
}

// deleteFromTimeline deletes tweets older than the given ageLimit from the user's timeling,
// using the given TwitterApi.
// Returns the number of tweets deleted and any error that occurred.
func deleteFromTimeline(api *anaconda.TwitterApi, ageLimit time.Duration) (int, error) {
	timeline, err := getTimeline(api)
	if err != nil {
		return 0, errors.Wrap(err, "could not get timeline")
	}

	count := 0
	errs := errset.ErrSet{}
	for _, t := range timeline {
		idStr := strconv.FormatInt(t.Id, 10)
		createdTime, err := t.CreatedAtTime()
		if err != nil {
			errs = append(errs, errors.Wrap(err, "could not parse time for "+idStr))
			continue
		}

		if time.Since(createdTime) > ageLimit && !isWhitelisted(t.Id) {
			_, err := api.DeleteTweet(t.Id, true)
			if err != nil {
				errs = append(errs, errors.Wrap(err, "failed to delete "+idStr))
				continue
			}
			count++
			log.Print("DELETED ID ", idStr, " CREATED AT ", createdTime, "; TWEET \"", t.Text, "\"")
		}
	}

	return count, errs.ReturnValue()
}

func ephemeral() (MyResponse, error) {
	anaconda.SetConsumerKey(consumerKey)
	anaconda.SetConsumerSecret(consumerSecret)
	api := anaconda.NewTwitterApi(accessToken, accessTokenSecret)
	api.SetLogger(anaconda.BasicLogger)

	h, err := time.ParseDuration(maxTweetAge)
	if err != nil {
		panic("cannot parse max tweet age " + maxTweetAge)
	}

	count, err := deleteFromTimeline(api, h)
	resultMessage := "Deleted " + strconv.Itoa(count) + " tweets."
	log.Print(resultMessage)
	if err != nil {
		errMessage := " Errors occurred: " + err.Error()
		resultMessage += " " + errMessage
		log.Print(errMessage)
	}

	return MyResponse{
		Message:    resultMessage,
		StatusCode: "200",
	}, nil
}

func main() {
	lambda.Start(ephemeral)
}
