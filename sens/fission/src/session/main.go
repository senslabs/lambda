package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/senslabs/alpha/sens/logger"

	"github.com/senslabs/alpha/sens/errors"
	"github.com/senslabs/alpha/sens/httpclient"
	"github.com/senslabs/lambda/sens/fission/config"
	"github.com/senslabs/lambda/sens/fission/request"
	"github.com/senslabs/lambda/sens/fission/response"
)

type Session struct {
	Id        string `json:"SessionId"`
	UserId    string `json:"UserId"`
	Name      string `json:"SessionName"`
	Type      string `json:"SessionType"`
	StartedAt int64  `json:"StartedAt"`
	EndedAt   int64  `json:"EndedAt"`
}

type Sessions []Session

type SessionProperty struct {
	SessionId string `json:"SessionId"`
	Key       string `json:"Key"`
	Value     string `json:"Value"`
}

type SessionProperties []SessionProperty

type SessionRecord struct {
	UserId    string  `json:"UserId"`
	Key       string  `json:"Key"`
	Timestamp int64   `json:"Timestamp"`
	Value     float64 `json:"Value"`
}

type SessionRecords []SessionRecord

type SessionEvent struct {
	Id        string `json:"Id"`
	UserId    string `json:"UserId"`
	Key       string `json:"Key"`
	StartTime int64  `json:"StartTime"`
	EndTime   int64  `json:"EndTime"`
}

type SessionEvents []SessionEvent

type SessionSleep struct {
	Duration            int64
	RecoveryValue       int64
	RecommendedRecovery int64
	SleepTime           int64
	WakeupTime          int64
	Score               int64
	HeartRates          TimeSeriesData
	BreathRates         TimeSeriesData
	Recovery            TimeSeriesData
	Stress              TimeSeriesData
	Stages              TimeSeriesData
	AverageVitals       struct {
		HeartRate  int64
		BreathRate int64
		Stress     int64
	}
	Points struct {
		SleepQuality int64
		SleepRoutine int64
		Vitals       int64
		Restfulness  int64
	}
}

type SessionSleepData []SessionSleep

type SessionSnapshot struct {
	LastSync   int64
	Score      int64
	HeartRate  int64
	BreathRate int64
	Duration   int64
	Recovery   int64
	Id         string
	UserId     string
	StartTime  int64
	EndTime    int64
}

type SessionSnapshots []SessionSnapshot

type OperatorUser struct {
	OpId   string `json:"OpId"`
	UserId string `json:"UserId"`
}

type OperatorUsers []OperatorUser

type OrganizationUser struct {
	OrgId  string `json:"OrgId"`
	UserId string `json:"UserId"`
}

type OrganizationUsers []OrganizationUser

type TimeSeries struct {
	Time  int64
	Value interface{}
}

type TimeSeriesData []TimeSeries

type SessionsSummary struct {
	Sleeps      int64
	Meditations int64
	Alerts      int64
	Date        int64
}

func fetchSessionProperties(sessionId string, requiredSessionProperties map[string]int64) map[string]int64 {
	// Fetching session properties
	for key := range requiredSessionProperties {
		sessionPropertiesUrl := fmt.Sprintf("%v/api/session-properties/find?and=SessionId^%v&and=Key^%v&limit=1", config.GetDatastoreUrl(), sessionId, key)
		sessionPropertiesResponseData := getFromDataStore(sessionPropertiesUrl)

		var sessionPropertiesData SessionProperties
		err := json.Unmarshal(sessionPropertiesResponseData, &sessionPropertiesData)
		if err != nil {
			log.Println("Error unmarshalling response data to session properties")
		}

		if len(sessionPropertiesData) > 0 {
			sValue := sessionPropertiesData[0].Value
			var value int64
			if key == "HeartRate" || key == "BreathRate" || key == "Recovery" || key == "Score" {
				value, _ = strconv.ParseInt(sValue, 10, 64)
			} else if key == "WakeupTime" || key == "SleepTime" || key == "SunriseTime" {
				value, _ = strconv.ParseInt(sValue, 10, 64)
			} else {
				value, _ = strconv.ParseInt(sValue, 10, 64)
			}
			requiredSessionProperties[key] = value
		}
	}

	return requiredSessionProperties
}

func fetchSessionRecords(sessionUserId string, sessionStartTime int64, sessionEndTime int64, requiredSessionRecords *map[string]TimeSeriesData) {
	// Fetch records
	for key := range *requiredSessionRecords {
		sessionRecordsUrl := fmt.Sprintf("%v/api/session-records/find?and=Key^%v&and=UserId^%v&span=Timestamp^%v^%v&limit=10000000", config.GetDatastoreUrl(), key, sessionUserId, sessionStartTime, sessionEndTime)
		sessionRecordsReponseData := getFromDataStore(sessionRecordsUrl)
		var sessionRecordsData SessionRecords
		json.Unmarshal(sessionRecordsReponseData, &sessionRecordsData)

		for _, value := range sessionRecordsData {
			timestamp := value.Timestamp
			var newEvent TimeSeries
			newEvent.Time = timestamp
			if key == "HeartRate" || key == "BreathRate" || key == "Stress" || key == "Recovery" || key == "StftRatio" {
				newEvent.Value = value.Value
			} else if key == "Stage" {
				newEvent.Value = int64(value.Value)
			}
			(*requiredSessionRecords)[key] = append((*requiredSessionRecords)[key], newEvent)
		}
	}
}

func getUserSessions(r *http.Request, sessionType string, limit int64, userId *string) Sessions {
	sFrom := request.GetQueryParam(r, "from")
	var from int64
	var to int64
	if len(sFrom) != 0 {
		from, _ = strconv.ParseInt(sFrom, 10, 64)
	}
	sTo := request.GetQueryParam(r, "to")
	if len(sTo) != 0 {
		to, _ = strconv.ParseInt(sTo, 10, 64)
	}

	var userIdList []string
	if userId == nil {
		userIdList = getUserList(r)
	} else {
		userIdList = append(userIdList, *userId)
	}

	userSessionsData := make(Sessions, 0)

	for _, currentUserId := range userIdList {
		var url string
		url = fmt.Sprintf("%v/api/sessions/find?and=UserId^%v&limit=%v&and=SessionType^%v&column=EndedAt", config.GetDatastoreUrl(), currentUserId, limit, sessionType)
		if from != 0 && to != 0 {
			url = fmt.Sprintf("%v&span=EndedAt^%v^%v", url, from, to)
		}

		userSessionResponseData := getFromDataStore(url)
		var currentSessionData Sessions
		json.Unmarshal(userSessionResponseData, &currentSessionData)
		userSessionsData = append(userSessionsData, currentSessionData...)
	}

	return userSessionsData
}

func getSessionSnapshot(sessionId string, sessionType string) SessionSnapshot {
	sessionData := getSessionData(sessionId)

	sessionUserId := sessionData.UserId
	sessionStartTime := sessionData.StartedAt
	sessionEndTime := sessionData.EndedAt

	requiredSessionProperties := map[string]int64{
		"Recovery": 0,
		"Score":    0,
	}

	if sessionType == "Sleep" {
		requiredSessionProperties["SleepTime"] = 0
		requiredSessionProperties["WakeupTime"] = 0
	} else {
		requiredSessionProperties["Duration"] = 0
	}

	requiredSessionProperties = fetchSessionProperties(sessionId, requiredSessionProperties)

	requiredSessionRecords := map[string]TimeSeriesData{
		"HeartRate":  make(TimeSeriesData, 0),
		"BreathRate": make(TimeSeriesData, 0),
		"Recovery":   make(TimeSeriesData, 0),
		"Stress":     make(TimeSeriesData, 0),
	}

	if sessionType == "Sleep" {
		requiredSessionRecords["Stage"] = make(TimeSeriesData, 0)
	}

	fetchSessionRecords(sessionUserId, sessionStartTime, sessionEndTime, &requiredSessionRecords)

	sessionSnapshot := createSessionSnapshotData(sessionData, requiredSessionProperties, requiredSessionRecords)

	return sessionSnapshot
}

func createSessionSnapshotData(sessionData Session, requiredSessionProperties map[string]int64, requiredSessionRecords map[string]TimeSeriesData) SessionSnapshot {
	sessionSleepTime := sessionData.StartedAt
	sessionWakeupTime := sessionData.EndedAt
	sessionType := sessionData.Type

	if sessionType == "Sleep" {
		sessionSleepTime = requiredSessionProperties["SleepTime"]
		sessionWakeupTime = requiredSessionProperties["WakeupTime"]
	}
	sessionScore := requiredSessionProperties["Score"]
	sessionRecovery := requiredSessionProperties["Recovery"]
	sessionHeartRateAverage := getVitalsAverage(requiredSessionRecords["HeartRate"], sessionSleepTime, sessionWakeupTime)
	sessionBreathRateAverage := getVitalsAverage(requiredSessionRecords["BreathRate"], sessionSleepTime, sessionWakeupTime)
	var sessionSleepDuration int64
	if sessionType == "Sleep" {
		sessionSleepDuration = getTotalDurationFromStages(requiredSessionRecords["Stage"])
	} else {
		sessionSleepDuration = requiredSessionProperties["Duration"]
	}

	var sessionSnapshot SessionSnapshot
	sessionSnapshot.Duration = sessionSleepDuration
	sessionSnapshot.Score = sessionScore
	sessionSnapshot.BreathRate = sessionBreathRateAverage
	sessionSnapshot.HeartRate = sessionHeartRateAverage
	sessionSnapshot.LastSync = sessionData.EndedAt
	sessionSnapshot.Recovery = sessionRecovery
	sessionSnapshot.Id = sessionData.Id
	sessionSnapshot.UserId = sessionData.UserId
	if sessionType == "Sleep" {
		sessionSnapshot.StartTime = sessionSleepTime
		sessionSnapshot.EndTime = sessionWakeupTime
	} else {
		sessionSnapshot.StartTime = sessionData.StartedAt
		sessionSnapshot.EndTime = sessionData.EndedAt
	}

	return sessionSnapshot
}

func getVitalsAverage(data TimeSeriesData, sleepTime int64, wakeupTime int64) int64 {
	var vitalsBetweenSleepTime float64
	var vitalsBetweenSleepTimeCounter float64
	for _, vital := range data {
		if vital.Time >= sleepTime && vital.Time <= wakeupTime {
			vitalValue := vital.Value.(float64)
			vitalsBetweenSleepTime += vitalValue
			vitalsBetweenSleepTimeCounter++
		}
	}
	var averageVital int64
	if vitalsBetweenSleepTimeCounter > 0 {
		averageVital = int64(vitalsBetweenSleepTime / vitalsBetweenSleepTimeCounter)
	} else {
		averageVital = 0
	}
	return averageVital
}

func getTotalDurationFromStages(stages TimeSeriesData) int64 {
	var sessionSleepDuration int64
	var eventTimeDiff int64
	var previousTime int64
	var sleepStageCounter int64
	for _, stage := range stages {
		if eventTimeDiff == 0 && previousTime == 0 {
			previousTime = stage.Time
		} else if eventTimeDiff == 0 && previousTime != 0 {
			eventTimeDiff = stage.Time - previousTime
		}
		currentStage := stage.Value.(int64)
		if currentStage != 4 {
			sleepStageCounter += 1
		}
	}
	sessionSleepDuration = sleepStageCounter * (eventTimeDiff / 1000)
	return sessionSleepDuration
}

func getUserList(r *http.Request) []string {
	var userIdList []string

	orgId := r.Header.Get("X-Sens-Org-Id")
	opId := r.Header.Get("X-Sens-Op-Id")
	userId := r.Header.Get("X-Sens-User-Id")

	if len(orgId) != 0 {
		// fetch users under this organization id
		orgUsers := getOrganizationUsers(orgId)
		userIdList = append(userIdList, orgUsers...)
	} else if len(opId) != 0 {
		// fetch users under this operator id
		opUsers := getOperatorUsers(opId)
		userIdList = append(userIdList, opUsers...)
	} else if len(userId) != 0 {
		// add userId to the userIdList
		userIdList = append(userIdList, userId)
	}

	return userIdList
}

func getSessionData(sessionId string) Session {
	sessionUrl := fmt.Sprintf("%v/api/sessions/%v/get", config.GetDatastoreUrl(), sessionId)

	sessionResponseData := getFromDataStore(sessionUrl)

	var sessionData Session
	err := json.Unmarshal(sessionResponseData, &sessionData)

	if err != nil {
		log.Printf("Error unmarshalling response data to sleep data : %v", err)
	}

	return sessionData
}

func getOrganizationUsers(orgId string) []string {
	var orgUsers []string
	url := fmt.Sprintf("%v/api/org-users/find?and=OrgId^%v&limit=10000", config.GetDatastoreUrl(), orgId)
	organizationUsersResponseData := getFromDataStore(url)
	var organizationUserData OrganizationUsers

	err := json.Unmarshal(organizationUsersResponseData, &organizationUserData)
	if err != nil {
		log.Println("Error fetching org users")
	}

	for _, orgUser := range organizationUserData {
		orgUsers = append(orgUsers, orgUser.UserId)
	}

	return orgUsers
}

func getOperatorUsers(orgId string) []string {
	var opUsers []string
	url := fmt.Sprintf("%v/api/op-users/find?and=OrgId^%v&limit=10000", config.GetDatastoreUrl(), orgId)
	operatorUsersResponseData := getFromDataStore(url)
	var operatorUserData OperatorUsers
	err := json.Unmarshal(operatorUsersResponseData, &operatorUserData)
	if err != nil {
		log.Println("Error fetching operator users")
		return opUsers
	}

	for _, opUser := range operatorUserData {
		opUsers = append(opUsers, opUser.UserId)
	}

	return opUsers
}

func Bod(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

func getDayStart(timestamp int64) int64 {
	timestampTime := time.Unix(timestamp, 0)
	startOfDay := Bod(timestampTime)
	startOfDayUnix := startOfDay.Unix()
	return startOfDayUnix
}

func getFromDataStore(URL string) []byte {
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		log.Panicf("Error creating a new request for fetching %v", URL)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Panic("Error performing a request")
	}
	responseBody, _ := ioutil.ReadAll(resp.Body)
	err = resp.Body.Close()
	if err != nil {
		log.Panic("Error closing the response body")
	}
	return responseBody
}

func GetSession(w http.ResponseWriter, r *http.Request) {
	os.Setenv("FILE_STORE", "fluentd")
	os.Setenv("LOG_LEVEL", "DEBUG")
	logger.InitLogger("sens.lambda.GetSession")
	// Get the sessionId from the request body
	// Fetch session details using the datastore link and type as sleep
	// Using the Start Time and End Time, fetch the following
	// 1. Fetch session properties using the sessionId
	// 2. Fetch records using the UserId, Timestamp should be between Session Start Time and Session End Time
	// 3. Fetch events using the UserId, Event Start Time should be between Session Start Time and Session End Time
	// Create Sleep Map Data
	// Return Sleep Map Data through Response
	sessionId := request.GetPathParam(r, "id")
	sessionData := getSessionData(sessionId)
	sessionUserId := sessionData.UserId
	sessionStartTime := sessionData.StartedAt
	sessionEndTime := sessionData.EndedAt

	requiredSessionProperties := map[string]int64{
		"Recovery":            0,
		"Score":               0,
		"SleepTime":           0,
		"WakeupTime":          0,
		"Duration":            0,
		"SleepQualityPoints":  0,
		"SleepRoutinePoints":  0,
		"VitalsPoints":        0,
		"RestfulnessPoints":   0,
		"RecommendedRecovery": 0,
	}

	requiredSessionProperties = fetchSessionProperties(sessionId, requiredSessionProperties)

	requiredSessionRecords := map[string]TimeSeriesData{
		"HeartRate":  make(TimeSeriesData, 0),
		"BreathRate": make(TimeSeriesData, 0),
		"Recovery":   make(TimeSeriesData, 0),
		"Stress":     make(TimeSeriesData, 0),
		"Stage":      make(TimeSeriesData, 0),
		"Snoring":    make(TimeSeriesData, 0),
	}

	fetchSessionRecords(sessionUserId, sessionStartTime, sessionEndTime, &requiredSessionRecords)

	sessionSleepTime := requiredSessionProperties["SleepTime"]
	sessionWakeupTime := requiredSessionProperties["WakeupTime"]
	sessionScore := requiredSessionProperties["Score"]
	sessionRecovery := requiredSessionProperties["Recovery"]
	sessionRecommendedRecovery := requiredSessionProperties["RecommendedRecovery"]
	sessionHeartRateAverage := getVitalsAverage(requiredSessionRecords["HeartRate"], sessionSleepTime, sessionWakeupTime)
	sessionBreathRateAverage := getVitalsAverage(requiredSessionRecords["BreathRate"], sessionSleepTime, sessionWakeupTime)
	sessionStressAverage := getVitalsAverage(requiredSessionRecords["Stress"], sessionSleepTime, sessionWakeupTime)
	sessionSleepDuration := getTotalDurationFromStages(requiredSessionRecords["Stage"])

	var sessionSessionData SessionSleep
	sessionSessionData.RecoveryValue = sessionRecovery
	sessionSessionData.RecommendedRecovery = sessionRecommendedRecovery
	sessionSessionData.Score = sessionScore
	sessionSessionData.SleepTime = sessionSleepTime
	sessionSessionData.WakeupTime = sessionWakeupTime
	sessionSessionData.Duration = sessionSleepDuration

	sessionSessionData.AverageVitals.HeartRate = sessionHeartRateAverage
	sessionSessionData.AverageVitals.BreathRate = sessionBreathRateAverage
	sessionSessionData.AverageVitals.Stress = sessionStressAverage

	sessionSessionData.Points.Vitals = requiredSessionProperties["VitalsPoints"]
	sessionSessionData.Points.SleepQuality = requiredSessionProperties["SleepQualityPoints"]
	sessionSessionData.Points.Restfulness = requiredSessionProperties["RestfulnessPoints"]
	sessionSessionData.Points.SleepRoutine = requiredSessionProperties["SleepRoutinePoints"]

	sessionSessionData.HeartRates = requiredSessionRecords["HeartRate"]
	sessionSessionData.BreathRates = requiredSessionRecords["BreathRate"]
	sessionSessionData.Recovery = requiredSessionRecords["Recovery"]
	sessionSessionData.Stages = requiredSessionRecords["Stage"]
	sessionSessionData.Stress = requiredSessionRecords["Stress"]

	json.NewEncoder(w).Encode(sessionSessionData)
}

func ListSessions(w http.ResponseWriter, r *http.Request) {
	os.Setenv("FILE_STORE", "fluentd")
	os.Setenv("LOG_LEVEL", "DEBUG")
	logger.InitLogger("sens.lambda.ListSessions")
	// Get the sessionId from the request body
	// Fetch session details using the datastore link and type as sleep
	// A function which fetches the following for the session
	// 1. Last Synced At
	// 2. Score
	// 3. Heart Rate
	// 4. Breath Rate
	// 5. Session Duration
	// 6. Recovery Value
	//sessionId := urlQueryParams.Get("id")
	logger.Debug(r)
	sessionType := request.GetQueryParam(r, "type")
	if len(sessionType) == 0 {
		httpclient.WriteError(w, http.StatusBadRequest, errors.New(http.StatusBadRequest, "Type not passed with request"))
		return
	} else {
		var limit int64
		sLimit := request.GetQueryParam(r, "limit")
		if len(sLimit) != 0 {
			limit, _ = strconv.ParseInt(sLimit, 10, 64)
		} else {
			limit = 1
		}

		userSessionsData := getUserSessions(r, sessionType, limit, nil)

		sessionsSnapshots := make(map[string]SessionSnapshots, 0)

		for _, currentSession := range userSessionsData {
			currentSessionId := currentSession.Id
			currentUserId := currentSession.UserId
			currentSessionType := currentSession.Type
			sessionSnapshotData := getSessionSnapshot(currentSessionId, currentSessionType)
			sessionsSnapshots[currentUserId] = append(sessionsSnapshots[currentUserId], sessionSnapshotData)
		}

		w.Header().Add("Content-Type", "application/json")

		json.NewEncoder(w).Encode(sessionsSnapshots)
	}
}

func GetGeneralSummary(w http.ResponseWriter, r *http.Request) {
	os.Setenv("FILE_STORE", "fluentd")
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("sens.lambda.ListSessions")

	// get days from the url query
	// take current date and then subtract the number of days to get the started_at date
	sDays := request.GetQueryParam(r, "days")
	if len(sDays) == 0 {
		http.Error(w, "Days not passed along with request", http.StatusBadRequest)
		return
	}
	days, _ := strconv.ParseInt(sDays, 10, 64)
	sStartDate := request.GetQueryParam(r, "start")
	var endDate int64
	if len(sStartDate) != 0 {
		endDate, _ = strconv.ParseInt(sStartDate, 10, 64)
	} else {
		endDate = time.Now().Unix()
	}
	startDate := endDate - days*3600*24

	userIdList := getUserList(r)

	generatedSummary := make(map[int64]SessionsSummary, 0)

	for _, currentUserId := range userIdList {
		url := fmt.Sprintf("%v/api/sessions/find?and=UserId^%v&span=EndedAt^%v^%v&limit=100000000", config.GetDatastoreUrl(), currentUserId, startDate, endDate)
		userSessionResponseData := getFromDataStore(url)
		var userSessionsData Sessions
		err := json.Unmarshal(userSessionResponseData, &userSessionsData)
		if err != nil {
			log.Println("Error unmarshalling session data")
		}

		for _, session := range userSessionsData {
			currentSessionType := session.Type
			sessionEndTime := session.EndedAt
			sessionKey := getDayStart(sessionEndTime)
			var currentDateSessionSummary SessionsSummary
			if sessionSummary, ok := generatedSummary[sessionKey]; ok {
				currentDateSessionSummary = sessionSummary
			} else {
				currentDateSessionSummary = SessionsSummary{}
			}
			if currentSessionType == "Sleep" {
				currentDateSessionSummary.Sleeps++
			} else if currentSessionType == "Meditation" {
				currentDateSessionSummary.Meditations++
			}
			generatedSummary[sessionKey] = currentDateSessionSummary
		}
	}

	generatedSummaryList := make([]SessionsSummary, 0)
	for key := range generatedSummary {
		currentSummary := generatedSummary[key]
		currentSummary.Date = key
		generatedSummaryList = append(generatedSummaryList, currentSummary)
	}

	w.Header().Add("Content-Type", "application/json")

	json.NewEncoder(w).Encode(generatedSummaryList)
}

func validateAndFetchQueryParameters(w http.ResponseWriter, r *http.Request) (string, *int64, *int64, *string, error) {
	sFrom := request.GetQueryParam(r, "from")
	sTo := request.GetQueryParam(r, "to")
	sessionId := request.GetPathParam(r, "id")
	property := request.GetQueryParam(r, "property")

	if len(sFrom) == 0 {
		return sessionId, nil, nil, nil, errors.New(http.StatusBadRequest, "from is missing in request")
	}
	if len(sTo) == 0 {
		return sessionId, nil, nil, nil, errors.New(http.StatusBadRequest, "to is missing in request")
	}
	if len(sessionId) == 0 {
		return sessionId, nil, nil, nil, errors.New(http.StatusBadRequest, "sessionId is missing in request")
	}
	if len(property) == 0 {
		return sessionId, nil, nil, nil, errors.New(http.StatusBadRequest, "property is missing in request")
	}

	from, _ := strconv.ParseInt(sFrom, 10, 64)
	to, _ := strconv.ParseInt(sTo, 10, 64)

	return sessionId, &from, &to, &property, nil
}

func GetSessionPropertyFunc(w http.ResponseWriter, r *http.Request) {
	os.Setenv("FILE_STORE", "fluentd")
	os.Setenv("LOG_LEVEL", "DEBUG")
	logger.InitLogger("sens.lambda.GetSessionPropertyFunc")

	sessionId, from, to, property, err := validateAndFetchQueryParameters(w, r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	timeData := map[string]TimeSeriesData{
		*property: make(TimeSeriesData, 0),
	}

	sessionData := getSessionData(sessionId)

	fetchSessionRecords(sessionData.UserId, *from, *to, &timeData)

	w.Header().Add("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(timeData)

	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, err)
	}
}

func GetParameterWiseAdvancedSessionData(w http.ResponseWriter, r *http.Request) {
	os.Setenv("FILE_STORE", "fluentd")
	os.Setenv("LOG_LEVEL", "DEBUG")
	logger.InitLogger("sens.lambda.GetParameterWiseAdvancedSessionData")

	sessionId := request.GetQueryParam(r, "id")

	sessionData := getSessionData(sessionId)

	sessionUserId := sessionData.UserId
	sessionStartTime := sessionData.StartedAt
	sessionEndTime := sessionData.EndedAt

	requestedKey := r.URL.Query().Get("dataKey")

	requestedData := map[string]TimeSeriesData{
		requestedKey: make(TimeSeriesData, 0),
	}

	fetchSessionRecords(sessionUserId, sessionStartTime, sessionEndTime, &requestedData)

	json.NewEncoder(w).Encode(requestedData)
}

func GetCategoryWiseAdvancedSessionData(w http.ResponseWriter, r *http.Request) {
	os.Setenv("FILE_STORE", "fluentd")
	os.Setenv("LOG_LEVEL", "DEBUG")
	logger.InitLogger("sens.lambda.GetCategoryWiseAdvancedSessionData")

	sessionId := request.GetPathParam(r, "id")

	session := getSessionData(sessionId)
	sessionUserId := session.UserId
	sessionStartTime := session.StartedAt
	sessionEndTime := session.EndedAt

	categoriesData := map[string]interface{}{
		"Stress": map[string]TimeSeriesData{
			"Vlf":   make(TimeSeriesData, 0),
			"Hf":    make(TimeSeriesData, 0),
			"Rmssd": make(TimeSeriesData, 0),
			"Pnn50": make(TimeSeriesData, 0),
		},
		"OriginalStress": map[string]TimeSeriesData{
			"Stress": make(TimeSeriesData, 0),
		},
		"Sleep": map[string]TimeSeriesData{
			"Stage": make(TimeSeriesData, 0),
		},
		"Heart": map[string]TimeSeriesData{
			"JjPeaks":   make(TimeSeriesData, 0),
			"HeartRate": make(TimeSeriesData, 0),
			"Sdnn":      make(TimeSeriesData, 0),
			"Rmssd":     make(TimeSeriesData, 0),
			"Pnn50":     make(TimeSeriesData, 0),
			"Vlf":       make(TimeSeriesData, 0),
			"Hf":        make(TimeSeriesData, 0),
		},
		"HRV Pack": map[string]TimeSeriesData{
			"Sdnn":  make(TimeSeriesData, 0),
			"Rmssd": make(TimeSeriesData, 0),
			"Pnn50": make(TimeSeriesData, 0),
			"Vlf":   make(TimeSeriesData, 0),
			"Hf":    make(TimeSeriesData, 0),
		},
		"Respiration": map[string]TimeSeriesData{
			"ZeroCrossing": make(TimeSeriesData, 0),
			"Snoring":      make(TimeSeriesData, 0),
			"Apnea":        make(TimeSeriesData, 0),
		},
	}

	requestedKey := r.URL.Query().Get("dataKey")
	w.Header().Add("Content-Type", "application/json")
	if value, ok := categoriesData[requestedKey]; ok {
		typeOfValue := reflect.TypeOf(value)
		if typeOfValue == reflect.TypeOf(map[string]TimeSeriesData{}) {
			typeCastedValue := value.(map[string]TimeSeriesData)
			fetchSessionRecords(sessionUserId, sessionStartTime, sessionEndTime, &typeCastedValue)
		}

		json.NewEncoder(w).Encode(value)
	} else {
		log.Println("No category by that name found")
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Category not found!",
		})
	}
}

func ListUserSessions(w http.ResponseWriter, r *http.Request) {
	os.Setenv("FILE_STORE", "fluentd")
	os.Setenv("LOG_LEVEL", "DEBUG")
	logger.InitLogger("sens.lambda.ListUserSessions")

	userId := request.GetPathParam(r, "id")
	sessionType := request.GetQueryParam(r, "type")
	if len(sessionType) == 0 {
		httpclient.WriteError(w, http.StatusBadRequest, errors.New(http.StatusBadRequest, "Type not passed with request"))
	} else {
		var limit int64
		sLimit := request.GetQueryParam(r, "limit")
		if len(sLimit) != 0 {
			limit, _ = strconv.ParseInt(sLimit, 10, 64)
		} else {
			limit = 1
		}
		userSessionsData := getUserSessions(r, sessionType, limit, &userId)

		userSessionsSnapshots := make(SessionSnapshots, 0)

		for _, currentSession := range userSessionsData {
			currentSessionId := currentSession.Id
			currentSessionType := currentSession.Type
			sessionSnapshotData := getSessionSnapshot(currentSessionId, currentSessionType)
			userSessionsSnapshots = append(userSessionsSnapshots, sessionSnapshotData)
		}

		w.Header().Add("Content-Type", "application/json")

		json.NewEncoder(w).Encode(userSessionsSnapshots)
	}
}
