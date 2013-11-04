package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"unicode/utf8"
)

const ALL_WK_KANJI = "wkkanji.txt"
const INCORRECT_API_KEY = "Incorrect"

const WK_API_URL = "http://www.wanikani.com/api/user/"
const WK_API_REQUEST_VOCAB = "/vocabulary/"
const WK_API_REQUEST_KANJI = "/kanji/"
const WK_API_REQUEST_USER_INFO = "/user-information"

const REGEXP_CONTAIN_KANJI = "[\u4E00-\u9FAF].*"

var InputFiles []string
var WaniKaniApiKey string
var Levels = ""

var KanjiFromWK string
var AllWKKanji string
var KanjiInText string
var UniqKanjiInTexts string
var TextLength int
var KanjiPercentageInTexts float32

var UnknownKanji string
var KanjiNotInWK string
var UnknownKanjiInWK string
var KnownKanji string

var UnknownKanjiPercentage float32
var KanjiNotInWKPercentage float32
var UnknownKanjiInWKPercentage float32
var KnownKanjiPercentage float32

func main() {

	for _, arg := range os.Args {
		pair := strings.Split(arg, "=")
		if len(pair) > 2 {
			continue
		}

		switch pair[0] {
		case "--apik":
			WaniKaniApiKey = pair[1]
			if len(WaniKaniApiKey) != 32 {
				WaniKaniApiKey = INCORRECT_API_KEY
				fmt.Printf("Incorrect API key. Please check your input and try again.\n")
			}
		case "--levels":
			Levels = pair[1]
			fmt.Printf("Request kanji only for levels: %s\n", Levels)
		default:
			if len(pair) == 1 && pair[0] != os.Args[0] {
				InputFiles = append(InputFiles, pair[0])
				fmt.Printf("Input file:%s.\n", pair[0])
			}
		}
	}

	var slice []string
	slice = append(slice, ALL_WK_KANJI)

	AllWKKanji, _ := readInputFiles(slice)

	if WaniKaniApiKey != INCORRECT_API_KEY {
		fmt.Print("API key is valid.\n")
		KanjiFromWK = loadWaniKaniData(WaniKaniApiKey)
		fmt.Printf("Your kanji list loaded. You already know %d/%d of WK kanji.\n", utf8.RuneCountInString(KanjiFromWK), utf8.RuneCountInString(AllWKKanji))
	} else {
		return
	}

	KanjiInText, TextLength = readInputFiles(InputFiles)
	UniqKanjiInTexts = uniqueKanjiInString(KanjiInText)
	KanjiPercentageInTexts = float32(utf8.RuneCountInString(KanjiInText)) / float32(TextLength) * 100

	UnknownKanji = kanjiDifference(UniqKanjiInTexts, KanjiFromWK)
	KanjiNotInWK = kanjiDifference(UnknownKanji, AllWKKanji)
	UnknownKanjiInWK = kanjiDifference(UnknownKanji, KanjiNotInWK)
	KnownKanji = kanjiDifference(UniqKanjiInTexts, UnknownKanji)

	UnknownKanjiPercentage = kanjiPercent(UnknownKanji, KanjiInText)
	KanjiNotInWKPercentage = kanjiPercent(KanjiNotInWK, KanjiInText)
	UnknownKanjiInWKPercentage = kanjiPercent(UnknownKanjiInWK, KanjiInText)
	KnownKanjiPercentage = kanjiPercent(KnownKanji, KanjiInText)

	fmt.Printf("All:%d	100.0%%\nKnown: %d	%3.1f%%\nUnknown:%d	%3.1f%%\n   WK: %d	%3.1f%%\n   not WK:%d	%3.1f%%\n", utf8.RuneCountInString(UniqKanjiInTexts), utf8.RuneCountInString(KnownKanji), KnownKanjiPercentage, utf8.RuneCountInString(UnknownKanji), UnknownKanjiPercentage, utf8.RuneCountInString(UnknownKanjiInWK), UnknownKanjiInWKPercentage, utf8.RuneCountInString(KanjiNotInWK), KanjiNotInWKPercentage)

	fmt.Printf("Kanji percent in texts: %3.1f%%\n", KanjiPercentageInTexts)

}

func loadWaniKaniData(apik string) string {

	fmt.Printf("Loading your WaniKani data.\n")
	res, err := http.Get(WK_API_URL + apik + WK_API_REQUEST_KANJI + Levels)
	if err != nil {
		log.Fatal(err)
	}

	type Item_stats struct {
		Srs                    string  `json:"srs"`
		Unlocked_date          float64 `json:"unlocked_date"`
		Available_date         float64 `json:"available_date"`
		Burned                 bool    `json:"burned"`
		Burned_date            float64 `json:"burned_date"`
		Meaning_correct        float64 `json:"meaning_correct"`
		Meaning_incorrect      float64 `json:"meaning_incorrect"`
		Meaning_max_streak     float64 `json:"meaning_max_streak"`
		Meaning_current_streak float64 `json:"meaning_current_streak"`
		Reading_correct        float64 `json:"reading_correct"`
		Reading_incorrect      float64 `json:"reading_incorrect"`
		Reading_max_streak     float64 `json:"reading_max_streak"`
		Reading_current_streak float64 `json:"reading_current_streak"`
	}

	type Kanji struct {
		Character        string  `json:"character"`
		Meaning          string  `json:"meaning"`
		Onyomi           string  `json:"onyomi"`
		Kunyomi          string  `json:"kunyomi"`
		ImportantReading string  `json:"important_reading"`
		Level            float64 `json:"level"`
		Stats            struct {
			Srs                    string  `json:"srs"`
			Unlocked_date          float64 `json:"unlocked_date"`
			Available_date         float64 `json:"available_date"`
			Burned                 bool    `json:"burned"`
			Burned_date            float64 `json:"burned_date"`
			Meaning_correct        float64 `json:"meaning_correct"`
			Meaning_incorrect      float64 `json:"meaning_incorrect"`
			Meaning_max_streak     float64 `json:"meaning_max_streak"`
			Meaning_current_streak float64 `json:"meaning_current_streak"`
			Reading_correct        float64 `json:"reading_correct"`
			Reading_incorrect      float64 `json:"reading_incorrect"`
			Reading_max_streak     float64 `json:"reading_max_streak"`
			Reading_current_streak float64 `json:"reading_current_streak"`
		} `json:"stats"`
	}

	type Requested_info struct {
		Items []Kanji `json:"general"`
	}

	type User_info struct {
		Username      string  `json:"username"`
		Gravatar      string  `json:"gravatar"`
		Level         float64 `json:"level"`
		Title         string  `json:"title"`
		About         string  `json:"about"`
		Website       string  `json:"website"`
		Twitter       string  `json:"twitter"`
		Topics_count  float64 `json:"topics_count"`
		Posts_count   float64 `json:"posts_count"`
		Creation_date float64 `json:"creation_date"`
	}

	type WKResponse struct {
		UserInfo      User_info      `json:"user_information"`
		RequestedInfo Requested_info `json:"requested_information"`
	}

	type WKResponseLimited struct {
		UserInfo      User_info `json:"user_information"`
		RequestedInfo []Kanji   `json:"requested_information"`
	}

	//	var inp = new(WKResponse)
	var inpLimited = new(WKResponseLimited)
	var encode_err error

	jsonResp, resp_err := ioutil.ReadAll(res.Body)

	if Levels != "" {
		encode_err = json.Unmarshal(jsonResp, &inpLimited)
	} else {
		encode_err = json.Unmarshal(jsonResp, &inpLimited)
	}

	if resp_err != nil {
		log.Fatal(resp_err)
		fmt.Printf("resperr")
	}
	if encode_err != nil {
		log.Fatal(encode_err)
		fmt.Printf("encerr")
	}

	res.Body.Close()

	var json = new(WKResponse)

	if Levels != "" {
		json.RequestedInfo.Items = inpLimited.RequestedInfo
		json.UserInfo = inpLimited.UserInfo
	} else {
		json.RequestedInfo.Items = inpLimited.RequestedInfo
		json.UserInfo = inpLimited.UserInfo
	}

	var result = ""

	for _, kanji := range json.RequestedInfo.Items {
		result = result + kanji.Character
	}

	fmt.Printf("Hello, %s of sect %s! ^_^\n", json.UserInfo.Username, json.UserInfo.Title)

	return result

}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func writeLines(lines []string, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}

func containKanji(arg string) bool {
	matched, merr := regexp.MatchString(REGEXP_CONTAIN_KANJI, arg)
	if merr != nil {
		log.Fatal(merr)
	}
	return matched
}

func readInputFiles(paths []string) (result string, textlen int) {

	var runeres []rune

	for _, file := range paths {
		fileData, err := readLines(file)
		if err != nil {
			log.Fatal(err)
			continue
		}
		for _, str := range fileData {
			textlen = textlen + utf8.RuneCountInString(str)
			for _, r := range str {
				if r > 0x4e00 && r < 0x9faf {
					runeres = append(runeres, r)
				}
			}
		}
	}

	result = string(runeres)

	return

}

func uniqueKanjiInString(arg string) (result string) {

	var runeres []rune

	for _, r := range arg {
		if !strings.ContainsRune(string(runeres), r) {
			runeres = append(runeres, r)
		}
	}

	result = string(runeres)

	return

}

func kanjiDifference(s1 string, s2 string) (result string) {

	var runeres []rune

	for _, r := range s1 {
		if !strings.ContainsRune(s2, r) {
			runeres = append(runeres, r)
		}
	}

	result = string(runeres)

	return

}

func kanjiPercent(s1 string, s2 string) (result float32) {

	var all = float32(utf8.RuneCountInString(s2))
	var res = 0

	for _, r := range s1 {
		if strings.ContainsRune(s2, r) {
			res = res + strings.Count(s2, string(r))
		}
	}

	result = float32(res) / all * 100

	return

}
