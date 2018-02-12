package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"github.com/ChimeraCoder/anaconda"
	_ "github.com/k0kubun/pp"
	"net/url"
	"os"
)

func tweetUrl(tweet *anaconda.Tweet) string {
	return `https://twitter.com/` + tweet.User.ScreenName + `/status/` + tweet.IdStr + `/`
}

func lastId() (*string, error) {
	file, err := os.Open("./last_id.txt")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var lastId string
	for scanner.Scan() {
		lastId = scanner.Text()
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "File scan error: %v\n", err)
		return nil, err
	}

	return &lastId, nil
}

func writeLastId(lastId string) {
	file, err := os.Create("./last_id.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	_, err = file.WriteString(lastId)
	if err != nil {
		panic(err)
	}
}

func writeCsv(records [][]string) {
	buf := new(bytes.Buffer)
	w := csv.NewWriter(buf)
	if err := w.WriteAll(records); err != nil {
		panic(err)
	}

	file, err := os.OpenFile("./tweets.csv", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	_, err = file.WriteString(buf.String())
	if err != nil {
		panic(err)
	}
}

func main() {
	anaconda.SetConsumerKey(os.Getenv(`TWITHINGS_CONSUMER_KEY`))
	anaconda.SetConsumerSecret(os.Getenv(`TWITHINGS_CONSUMER_SECRET`))
	api := anaconda.NewTwitterApi(os.Getenv(`TWITHINGS_ACCESS_TOKEN`), os.Getenv(`TWITHINGS_ACCESS_TOKEN_SECRET`))

	v := url.Values{}
	v.Set("lang", "ja")
	v.Set("count", "100")

	lastId, err := lastId()
	if err == nil {
		v.Set("since_id", *lastId)
	}

	// https://developer.twitter.com/en/docs/tweets/search/api-reference/get-search-tweets
	// https://developer.twitter.com/en/docs/tweets/search/guides/standard-operators
	searchString := os.Getenv(`TWITHINGS_TWEET_SEARCH_STRING`)
	if searchString == "" {
		panic("require TWITHINGS_TWEET_SEARCH_STRING env\nhttps://developer.twitter.com/en/docs/tweets/search/api-reference/get-search-tweets")
	}
	searchResult, _ := api.GetSearch(searchString, v)
	max_index := len(searchResult.Statuses) - 1

	var records [][]string
	for i := range searchResult.Statuses {
		tweet := searchResult.Statuses[max_index-i]
		//pp.Print(tweet)
		record := []string{tweet.CreatedAt, tweet.User.ScreenName, tweet.FullText, tweet.User.Description, tweetUrl(&tweet)}
		records = append(records, record)

		if i == max_index {
			writeLastId(tweet.IdStr)
		}
	}

	writeCsv(records)
}
