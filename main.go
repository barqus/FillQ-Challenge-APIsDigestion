//Lambda Function Go Code
// http://ddragon.leagueoflegends.com/cdn/6.3.1/img/profileicon/profileIconId.png
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
	ParticipantName   string
	LolSummonerName   string
	TwitchChannelName string
}

type League struct {
	LeagueID     string `json:"leagueId"`
	SummonerIcon float64
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
	Participant  string  `json:"Participant"`
	SummonerName string  `json:"summonerName"`
	SummonerIcon float64 `json:"summonerIcon"`
	Tier         string  `json:"tier"`
	Rank         string  `json:"rank"`
	LeaguePoints int     `json:"leaguePoints"`
	Wins         int     `json:"wins"`
	Losses       int     `json:"losses"`
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

var wg sync.WaitGroup

//ResponseBody for output to lambda
var ResponseBody string

func main() {
	var LambdaOutputs []LambdaOutput

	participants := []Participant{
		Participant{"Optimas", "OptimasLinijas", "optimaslol"},
		Participant{"ChosenOne", "ChosenOne Filler", "chosenonelol"},
		Participant{"Real", "wo mei yali", "realzyyy_"},
		Participant{"KNOK1", "YouCantKillMe", "knok1zygis"},
		Participant{"Saulius", "MarkRank1", "sauliuslol"},
		Participant{"Sponsorius", "S11 was So Fun", "sponsorius"},
		Participant{"Dethron", "sergu", "dethronlol"},
		Participant{"Yashiro", "DonOuei38", "yaashiro"},
	}

	for _, element := range participants {
		out1 := getTwitchChannelInformation(element.TwitchChannelName, twitchOAUTH)
		out2 := getSummonerData(element.LolSummonerName)

		out3 := convertToOutputJSON(out1, out2, element.ParticipantName)
		LambdaOutputs = append(LambdaOutputs, out3)
	}
	e, _ := json.Marshal(LambdaOutputs)
	ResponseBody = string(e)
	fmt.Println(ResponseBody)
	// lambda.Start(HandleRequest)
}

//HandleRequest serves data to lambda
func HandleRequest(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	resp := events.APIGatewayProxyResponse{Headers: make(map[string]string)}
	resp.Headers["Access-Control-Allow-Origin"] = "*"

	if request.HTTPMethod == "GET" {
		APIResponse := events.APIGatewayProxyResponse{Body: ResponseBody, StatusCode: 200, Headers: resp.Headers}
		return APIResponse, nil
	} else {
		err := errors.New("method not allowed")
		APIResponse := events.APIGatewayProxyResponse{Body: "Method Not OK", StatusCode: 502}
		return APIResponse, err
	}
}

func getSummonerData(summonerName string) League {
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
	summonerIcon, _ := data["profileIconId"].(float64)

	req, err = http.NewRequest("GET", summonerLeagueInformation+summonerID, nil)
	req.Header.Set("X-Riot-Token", riotToken)
	resp, err = client.Do(req)
	defer resp.Body.Close()
	bodyBytes, _ = ioutil.ReadAll(resp.Body)
	var allLeagues [2]League
	var toReturn League
	json.Unmarshal(bodyBytes, &allLeagues)
	s := string(bodyBytes)
	if s == "[]" {
		toReturn.SummonerIcon = summonerIcon
		toReturn.SummonerName = summonerName
	}

	for _, element := range allLeagues {
		if element.QueueType == "RANKED_SOLO_5x5" {
			element.SummonerIcon = summonerIcon
			toReturn = element
			break
		}
	}
	return toReturn
}

func getTwitchChannelInformation(twitchChannelName string, accessToken string) ChannelInformation {
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
	var toReturn ChannelInformation
	json.Unmarshal(bodyBytes, &allChannelGathered)
	for _, element := range allChannelGathered.Data {
		if strings.EqualFold(element.DisplayName, twitchChannelName) {
			toReturn = element
		}
	}

	return toReturn
}

func convertToOutputJSON(channelInformation ChannelInformation, summonerLeague League, ParticipantName string) LambdaOutput {
	lambdaOutput := LambdaOutput{
		Participant:  ParticipantName,
		SummonerName: summonerLeague.SummonerName,
		SummonerIcon: summonerLeague.SummonerIcon,
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
	return lambdaOutput
}
