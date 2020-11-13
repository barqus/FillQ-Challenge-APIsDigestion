//Lambda Function Go Code
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const summonerInformation = "https://euw1.api.riotgames.com/lol/summoner/v4/summoners/by-name/"
const summonerLeagueInformation = "https://euw1.api.riotgames.com/lol/league/v4/entries/by-summoner/"
const riotToken = "RGAPI-14e029e4-0a8d-4b22-81e3-47210056b453"

const twitchAuth = "https://id.twitch.tv/oauth2/token"
const twitchClientID = "by7zl6rwazu7ks1z6sby63bnwq1267"
const twitchClientSecret = "97idvjwxeb2clcc5j4nwr3g5ebr12m"

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

type Channel struct {
	LolSummonerName   string
	TwitchChannelName string
}

// 'X-Riot-Token': 'RGAPI-6b4aa854-1d3c-447b-8432-c94eda4462e8'
// https://euw1.api.riotgames.com/lol/summoner/v4/summoners/by-name/Choseris
func main() {

	participants := [2]Participant{
		Participant{"Choseris", "ChosenOneLol"},
		Participant{"KNOK1", "KNOK1ZYGIS"},
	}
	fmt.Println(participants[0].LolSummonerName)
	_ = getSummonerData(participants[0].LolSummonerName)
	_ = getTwitchData(participants[0].TwitchChannelName)
	// lambda.Start(HandleRequest)
}

// func HandleRequest(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
// 	if request.HTTPMethod == "GET" {
// 		var stringResponse string = "Yay a successful Response!!"
// 		ApiResponse := events.APIGatewayProxyResponse{Body: stringResponse, StatusCode: 200}
// 		return ApiResponse, nil
// 	} else {
// 		err := errors.New("Method Not Allowed!")
// 		ApiResponse := events.APIGatewayProxyResponse{Body: "Method Not OK", StatusCode: 502}
// 		return ApiResponse, err
// 	}
// }

// https://euw1.api.riotgames.com/lol/league/v4/entries/by-summoner/
func getSummonerData(summonerName string) (summonerLeague *League) {
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

	// Convert response body to Summoner struct
	var allLeagues [2]League

	json.Unmarshal(bodyBytes, &allLeagues)
	for _, element := range allLeagues {
		if element.QueueType == "RANKED_SOLO_5x5" {
			rankedInformation := &element
			return rankedInformation
		}
		// element is the element from someSlice for where we are
	}
	return nil
}

func getTwitchData(twitchChannelName string) (twitchInfo *Channel) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", twitchAuth, nil)
	if err != nil {
		// return nil
	}
	q := req.URL.Query()
	q.Add("client_id", twitchClientID)
	q.Add("client_secret", twitchClientSecret)
	q.Add("grant_type", "client_credentials")
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		// return nil
	}
	fmt.Println(resp)
	data := make(map[string]interface{})
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &data)
	summonerID, _ := data["access_token"].(string)
	fmt.Println(summonerID)
	return nil
}

/*
TWITCH API:

For token = https://id.twitch.tv/oauth2/token?client_id=<>&client_secret=<>&grant_type=client_credentials

https://api.twitch.tv/helix/search/channels?query=Sponsorius
Headers:
Client-ID:
Authentication: Beare <token>

*/
