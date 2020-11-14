//Lambda Function Go Code
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

const summonerInformation = "https://euw1.api.riotgames.com/lol/summoner/v4/summoners/by-name/"
const summonerLeagueInformation = "https://euw1.api.riotgames.com/lol/league/v4/entries/by-summoner/"

var riotToken = os.Getenv("riotToken")

const twitchAuth = "https://id.twitch.tv/oauth2/token"

var twitchClientID = os.Getenv("twitchClientID")
var twitchClientSecret = os.Getenv("twitchClientSecret")

var twitchOAUTH = os.Getenv("twitchOAUTH")

const twitchChannelURL = "https://api.twitch.tv/helix/search/channels"

type Participant struct {
	LolSummonerName   string
	TwitchChannelName string
}

type League struct {
	LeagueID     string `json:"leagueId"`
	QueueType    string `json:"queueType"`
	Tier         string `json:"tier"`
	Rank         string `json:"rank"`
	SummonerID   string `json:"summonerId"`
	SummonerName string `json:"summonerName"`
	LeaguePoints int    `json:"leaguePoints"`
	Wins         int    `json:"wins"`
	Losses       int    `json:"losses"`
	Veteran      bool   `json:"veteran"`
	Inactive     bool   `json:"inactive"`
	FreshBlood   bool   `json:"freshBlood"`
	HotStreak    bool   `json:"hotStreak"`
	MiniSeries   struct {
		Target   int    `json:"target"`
		Wins     int    `json:"wins"`
		Losses   int    `json:"losses"`
		Progress string `json:"progress"`
	} `json:"miniSeries,omitempty"`
}

// type ChannelInformation struct {
// 	BroadcasterLanguage string   `map:"broadcaster_language"`
// 	DisplayName         string   `map:"display_name"`
// 	GameID              string   `map:"game_id"`
// 	ID                  string   `map:"id"`
// 	IsLive              bool     `map:"is_live"`
// 	StartedAt           string   `map:"started_at"`
// 	TagIds              []string `map:"tag_ids"`
// 	ThumbnailURL        string   `map:"thumbnail_url"`
// 	Title               string   `map:"title"`
// }

type ChannelInformation struct {
	BroadcasterLanguage string   `json:"broadcaster_language"`
	DisplayName         string   `json:"display_name"`
	GameID              string   `json:"game_id"`
	ID                  string   `json:"id"`
	IsLive              bool     `json:"is_live"`
	StartedAt           string   `json:"started_at"`
	TagIds              []string `json:"tag_ids"`
	ThumbnailURL        string   `json:"thumbnail_url"`
	Title               string   `json:"title"`
}

type ChannelsGathered struct {
	Data []ChannelInformation `json:"data"`
}

type LambdaOutput struct {
	SummonerName string `json:"summonerName"`
	Tier         string `json:"tier"`
	Rank         string `json:"rank"`
	LeaguePoints int    `json:"leaguePoints"`
	Wins         int    `json:"wins"`
	Losses       int    `json:"losses"`
	MiniSeries   struct {
		Target   int    `json:"target"`
		Wins     int    `json:"wins"`
		Losses   int    `json:"losses"`
		Progress string `json:"progress"`
	} `json:"miniSeries,omitempty"`
	DisplayName  string `json:"display_name"`
	IsLive       bool   `json:"is_live"`
	StartedAt    string `json:"started_at"`
	ThumbnailURL string `json:"thumbnail_url"`
	Title        string `json:"title"`
}

type LambdaOutputs []LambdaOutput

var wg sync.WaitGroup

//ResponseBody for output to lambda
var ResponseBody string

func main() {

	// currentTime := time.Now()
	// if currentTime.Day() == 15 {
	// 	accessToken := getTwitchAuthToken()
	// 	fmt.Println(currentTime)
	// }

	// accessToken := getTwitchAuthToken()
	fmt.Println(riotToken, twitchClientID, twitchClientSecret, twitchOAUTH)
	participants := []Participant{
		Participant{"Choseris", "ChosenOneLol"},
		Participant{"KNOK1", "knok1zygis"},
	}
	summonerThreadChannel := make(chan League, len(participants))
	twitchThreadChannel := make(chan ChannelInformation, len(participants))
	outputThreadChannel := make(chan LambdaOutput, len(participants))
	for _, element := range participants {
		wg.Add(3)
		go getSummonerData(summonerThreadChannel, element.LolSummonerName)
		go getTwitchChannelInformation(twitchThreadChannel, element.TwitchChannelName, twitchOAUTH)
		out1 := <-summonerThreadChannel
		out2 := <-twitchThreadChannel
		go convertToOutputJSON(outputThreadChannel, out2, out1)
		// break
	}

	wg.Wait()
	close(summonerThreadChannel)
	close(twitchThreadChannel)
	close(outputThreadChannel)

	var OutputArray LambdaOutputs
	for element := range outputThreadChannel {
		OutputArray = append(OutputArray, element)
	}
	e, err := json.Marshal(OutputArray)
	if err != nil {
		fmt.Println(err)
		return
	}
	ResponseBody = string(e)
	fmt.Println(ResponseBody)
	lambda.Start(HandleRequest)
}

//HandleRequest serves data to lambda
func HandleRequest(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if request.HTTPMethod == "GET" {
		APIResponse := events.APIGatewayProxyResponse{Body: ResponseBody, StatusCode: 200}
		return APIResponse, nil
	} else {
		err := errors.New("method not allowed")
		APIResponse := events.APIGatewayProxyResponse{Body: "Method Not OK", StatusCode: 502}
		return APIResponse, err
	}
}

func getSummonerData(c chan League, summonerName string) {
	defer wg.Done()
	client := &http.Client{}
	req, err := http.NewRequest("GET", summonerInformation+summonerName, nil)
	if err != nil {
		// return nil
	}
	req.Header.Set("X-Riot-Token", riotToken)

	resp, err := client.Do(req)
	if err != nil {
		// return nil
	}

	data := make(map[string]interface{})
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &data)
	summonerID, _ := data["id"].(string)

	req, err = http.NewRequest("GET", summonerLeagueInformation+summonerID, nil)
	req.Header.Set("X-Riot-Token", riotToken)
	resp, err = client.Do(req)
	defer resp.Body.Close()
	bodyBytes, _ = ioutil.ReadAll(resp.Body)

	var allLeagues [2]League

	json.Unmarshal(bodyBytes, &allLeagues)
	for _, element := range allLeagues {
		if element.QueueType == "RANKED_SOLO_5x5" {
			rankedInformation := element
			c <- rankedInformation
		}
	}

}

// func getTwitchAuthToken() (token string) {
// 	client := &http.Client{}
// 	req, err := http.NewRequest("POST", twitchAuth, nil)
// 	if err != nil {
// 		// return nil
// 	}
// 	q := req.URL.Query()
// 	q.Add("client_id", twitchClientID)
// 	q.Add("client_secret", twitchClientSecret)
// 	q.Add("grant_type", "client_credentials")
// 	req.URL.RawQuery = q.Encode()
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		// return nil
// 	}

// 	data := make(map[string]interface{})
// 	bodyBytes, _ := ioutil.ReadAll(resp.Body)
// 	json.Unmarshal(bodyBytes, &data)
// 	accessToken, _ := data["access_token"].(string)
// 	return accessToken
// }

func getTwitchChannelInformation(c chan ChannelInformation, twitchChannelName string, accessToken string) {
	defer wg.Done()
	// ?query=Sponsorius
	client := &http.Client{}
	req, err := http.NewRequest("GET", twitchChannelURL, nil)
	if err != nil {
		// return nil
	}
	req.Header.Set("Client-ID", twitchClientID)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	q := req.URL.Query()
	q.Add("query", twitchChannelName)
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		// return nil
	}
	bodyBytes, _ := ioutil.ReadAll(resp.Body)

	// Convert response body to Summoner struct
	var allChannelGathered ChannelsGathered

	json.Unmarshal(bodyBytes, &allChannelGathered)
	for _, element := range allChannelGathered.Data {
		if strings.EqualFold(element.DisplayName, twitchChannelName) {
			c <- element
		}
	}
}

func convertToOutputJSON(c chan LambdaOutput, channelInformation ChannelInformation, summonerLeague League) {
	defer wg.Done()
	lambdaOutput := LambdaOutput{
		SummonerName: summonerLeague.SummonerName,
		Tier:         summonerLeague.Tier,
		Rank:         summonerLeague.Rank,
		LeaguePoints: summonerLeague.LeaguePoints,
		Wins:         summonerLeague.Wins,
		Losses:       summonerLeague.Losses,
		MiniSeries:   summonerLeague.MiniSeries,
		DisplayName:  channelInformation.DisplayName,
		IsLive:       channelInformation.IsLive,
		StartedAt:    channelInformation.StartedAt,
		ThumbnailURL: channelInformation.ThumbnailURL,
		Title:        channelInformation.Title,
	}
	c <- lambdaOutput
}
