// Logs the status of a user from Lotus Notes SameTime Instant Messager.
// Usage:
//
//    go run main.go -userid=USER_ID
//
// @requires Lotus Notes 8+, Google Go 1.2
// @project https://github.com/LarryBattle/SameTimeTrackStatus/
// @author Larry Battle
// @version 0.1.2
// @todo Add new flags : -users=id1,id2,id3 -verbose=bool -api_url=string -interval=#minutes
// @todo add check for Lotus Notes and setting.
// @todo refactor into objects; webapi, storage, cli, settings, user
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	VERSION             = "0.1.2"
	TIME_STAMP_FORMAT   = "01/02/2006 03:04:05pm"
	DEFAULT_OUTPUT_FILE = "output.txt"
)
const (
	RETURN_CODE_NO_PROBLEMS = iota
	RETURN_CODE_INVALID_ARGUMENT
	RETURN_CODE_MISSING_REQUIRED_ARGUMENT
	RETURN_CODE_WEB_API_DISABLED
)

var (
	sametime_getstatus_URL = `http://localhost:59449/stwebapi/getstatus?userId=`
	outputFile             string
	userId                 string
	showVersion            bool
	numOfMinutes           uint
)

// Used to only contain the essenential properties from the json response
type essential_ST_JSON struct {
	TimeStamp     string `json:"timestamp"`
	UnixTimeStamp int64  `json:"unixTimestamp"`
	DisplayName   string `json:"displayName"`
	Status        int    `json:"status"`
	StatusMessage string `json:"statusMessage"`
	UserName      string `json:"username"`
}

// Processes the flag information.
func processFlags() {
	flag.StringVar(&userId, "userid", "", "REQUIRED. Sametime User Id. Try your id if you don't know.")
	flag.StringVar(&outputFile, "output", DEFAULT_OUTPUT_FILE, "Output file to store logs.")
	flag.BoolVar(&showVersion, "version", false, "Shows version information.")
	flag.UintVar(&numOfMinutes, "interval", 5, "Interval to check status.")
	flag.Parse()
}

// Checks if all required flags are set.
func checkSettings() {
	if showVersion {
		fmt.Printf("Version %s\n", VERSION)
		os.Exit(RETURN_CODE_NO_PROBLEMS)
	}
	if userId == "" {
		fmt.Println("The argument `userid` is required.")
		os.Exit(RETURN_CODE_MISSING_REQUIRED_ARGUMENT)
	}
	if numOfMinutes < 1 || 200 < numOfMinutes {
		fmt.Println("The argument `interval` must be between 1 and 200.")
		os.Exit(RETURN_CODE_INVALID_ARGUMENT)
	}
	res, err := getUserInfo(userId)
	if err != nil || res == nil {
		fmt.Println("WebAPI is disabled. Please follow the documentation on how to enabled this.")
		os.Exit(RETURN_CODE_WEB_API_DISABLED)
	}
}

// Shows a message when the tool is called.
func printGreeting() {
	fmt.Println(`SameTime IM Status Tracking Tool by Larry Battle`)
}
func checkError(e error) {
	if e != nil {
		panic(e)
	}
}

// Returns a date time timestamp
func getTimeStamp() string {
	return time.Now().Format(TIME_STAMP_FORMAT)
}

// Returns the timestamp used by Javascript.
// Ex. new Date( timestamp )
func getJSTimeStamp() int64 {
	return time.Now().UnixNano() / 1e6
}

// Returns only the desired properties from the json response
func extractInfoFromJSON(json_string []byte) []byte {
	var obj essential_ST_JSON
	json.Unmarshal(json_string, &obj)
	obj.TimeStamp = getTimeStamp()
	obj.UnixTimeStamp = getJSTimeStamp()
	b, err := json.Marshal(obj)
	checkError(err)
	return b
}
func getUserInfo(userId string) (*http.Response, error) {
	return http.Get(sametime_getstatus_URL + userId)
}

// Returns the JSON response from a `getstatus` webapi call for a specific userId
func getSameTimeStatusOfUser(userId string) []byte {
	res, err := getUserInfo(userId)
	checkError(err)
	defer res.Body.Close()
	json_response, err := ioutil.ReadAll(res.Body)
	checkError(err)
	return json_response
}

// Appends a string with a new line to a file
func appendStringToFile(filename string, data []byte) {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
	defer f.Close()
	checkError(err)
	_, err = f.WriteString(string(data) + "\n")
	checkError(err)
}

// Sends status request then saves response to a file.
func logSameTimeStatus(userId string) {
	appendStringToFile(outputFile, extractInfoFromJSON(getSameTimeStatusOfUser(userId)))
}

// Calls a function every t times.
func startCounter(t time.Duration, fn func()) {
	i := 0
	var x = func() {
		i++
		log.Println("Logging status #", i)
		fn()
	}
	x()
	for _ = range time.Tick(t) {
		x()
	}
}
func main() {
	printGreeting()
	processFlags()
	checkSettings()
	log.Printf("Every %d minutes: Saving status for %s to %s\n", numOfMinutes, userId, outputFile)
	startCounter(time.Duration(numOfMinutes)*time.Minute, func() {
		logSameTimeStatus(userId)
	})
}
