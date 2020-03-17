package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Session struct {
	Id        string `json:"Id"`
	UserId    string `json:"UserId"`
	Name      string `json:"Name"`
	Type      string `json:"Type"`
	StartTime int64  `json:"StartTime"`
	EndTime   int64  `json:"EndTime"`
}

type Sessions []Session

type SessionProperty struct {
	SessionId string `json:"SessionId"`
	Name      string `json:"name"`
	Value     string `json:"value"`
}

type SessionProperties []SessionProperty

type SessionRecord struct {
	UserId    string `json:"UserId"`
	Name      string `json:"Name"`
	Timestamp int64  `json:"Timestamp"`
	Value     string `json:"Value"`
}

type SessionRecords []SessionRecord

type SessionEvent struct {
	Id        string `json:"Id"`
	UserId    string `json:"UserId"`
	Name      string `json:"Name"`
	StartTime int64  `json:"StartTime"`
	EndTime   int64  `json:"EndTime"`
}

type SessionEvents []SessionEvent

type SessionSleep struct {
	Duration      int64
	RecoveryValue int64
	SleepTime     int64
	WakeupTime    int64
	Score         int64
	HeartRates    TimeSeriesData
	BreathRates   TimeSeriesData
	Recovery      TimeSeriesData
	Stress        TimeSeriesData
	Stages        TimeSeriesData
	AverageVitals struct {
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
}

func fetchSessionProperties(sessionId string, requiredSessionProperties map[string]int64) map[string]int64 {
	// Fetching session properties
	for key := range requiredSessionProperties {
		sessionPropertiesUrl := fmt.Sprintf("http://35.225.36.244:9804/api/sessions-properties/find/?and=sessionId:%v&and=name:%v&limit=1", sessionId, key)
		sessionPropertiesResponseData := getFromDataStore(sessionPropertiesUrl)

		var sessionPropertiesData SessionProperties
		err := json.Unmarshal(sessionPropertiesResponseData, &sessionPropertiesData)
		if err != nil {
			log.Println("Error unmarshalling response data to session properties")
		}

		sValue := sessionPropertiesData[0].Value
		value, _ := strconv.ParseInt(sValue, 10, 64)
		requiredSessionProperties[key] = value
	}

	return requiredSessionProperties
}

func fetchSessionRecords(sessionStartTime int64, sessionEndTime int64, requiredSessionRecords map[string]TimeSeriesData) map[string]TimeSeriesData {
	// Fetch records
	for key, _ := range requiredSessionRecords {
		sessionRecordsUrl := fmt.Sprintf("http://35.225.36.244:9804/api/sessions-records/find/?and=name:%v&range:%v:%v", key, sessionStartTime, sessionEndTime)
		sessionRecordsReponseData := getFromDataStore(sessionRecordsUrl)
		var sessionRecordsData SessionRecords
		json.Unmarshal(sessionRecordsReponseData, &sessionRecordsData)

		for _, value := range sessionRecordsData {
			timestamp := value.Timestamp
			var newEvent TimeSeries
			newEvent.Time = timestamp
			if key == "HeartRate" || key == "BreathRate" {
				value, _ := strconv.ParseFloat(value.Value, 10)
				newEvent.Value = value
			} else if key == "Stage" {
				value, _ := strconv.ParseInt(value.Value, 10, 64)
				newEvent.Value = value
			}
			requiredSessionRecords[key] = append(requiredSessionRecords[key], newEvent)
		}
	}

	return requiredSessionRecords
}

func getUserSessions(r *http.Request) Sessions {
	urlQueryParams := r.URL.Query()
	//sessionId := urlQueryParams.Get("id")
	sessionType := urlQueryParams.Get("type")
	sLimit := urlQueryParams.Get("limit")
	var limit int64
	if len(sLimit) != 0 {
		limit, _ = strconv.ParseInt(sLimit, 10, 64)
	} else {
		limit = 1
	}

	//sFrom := urlQueryParams.Get("from")
	//from, _ := strconv.ParseInt(sFrom, 10, 64)
	//sTo := urlQueryParams.Get("to")
	//to, _ := strconv.ParseInt(sTo, 10, 64)

	userIdList := getUserList(r)

	var userSessionsData Sessions

	for _, currentUserId := range userIdList {
		url := fmt.Sprintf("http://35.225.36.244:9804/api/sessions/find/?and=userId:%v&limit=%v&type=%v", currentUserId, limit, sessionType)
		userSessionResponseData := getFromDataStore(url)
		json.Unmarshal(userSessionResponseData, &userSessionsData)
	}

	return userSessionsData
}

func getSessionSnapshot(sessionId string, sessionType string) SessionSnapshot {
	sessionUrl := fmt.Sprintf("http://35.225.36.244:9804/api/sessions/get/%v", sessionId)

	sessionResponseData := getFromDataStore(sessionUrl)

	var sessionData Session
	err := json.Unmarshal(sessionResponseData, &sessionData)

	if err != nil {
		log.Println("Error unmarshalling response data to sleep data")
	}

	sessionStartTime := sessionData.StartTime
	sessionEndTime := sessionData.EndTime

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

	requiredSessionRecords = fetchSessionRecords(sessionStartTime, sessionEndTime, requiredSessionRecords)

	sessionSnapshot := createSessionSnapshotData(sessionStartTime, sessionEndTime, sessionType, requiredSessionProperties, requiredSessionRecords)

	return sessionSnapshot
}

func createSessionSnapshotData(sessionStartTime int64, sessionEndTime int64, sessionType string, requiredSessionProperties map[string]int64, requiredSessionRecords map[string]TimeSeriesData) SessionSnapshot {
	sessionSleepTime := sessionStartTime
	sessionWakeupTime := sessionEndTime

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
		sessionSleepDuration = getTotalDurationFromStages(requiredSessionRecords["Stages"])
	} else {
		sessionSleepDuration = requiredSessionProperties["Duration"]
	}

	var sessionSnapshot SessionSnapshot
	sessionSnapshot.Duration = sessionSleepDuration
	sessionSnapshot.Score = sessionScore
	sessionSnapshot.BreathRate = sessionBreathRateAverage
	sessionSnapshot.HeartRate = sessionHeartRateAverage
	sessionSnapshot.LastSync = sessionEndTime
	sessionSnapshot.Recovery = sessionRecovery

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
	averageVital := int64(vitalsBetweenSleepTime / vitalsBetweenSleepTimeCounter)
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
		if stage.Value != 4 {
			sleepStageCounter += 1
		}
	}
	sessionSleepDuration = sleepStageCounter * eventTimeDiff
	return sessionSleepDuration
}

func getUserList(r *http.Request) []string {
	var userIdList []string

	orgId := r.Header.Get("x-sens-org-id")
	opId := r.Header.Get("x-sens-op-id")
	userId := r.Header.Get("x-sens-op-id")

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

//func GetParameterWiseAdvancedSessionData(w http.ResponseWriter, r *http.Request) {
//	urlQueryParams := r.URL.Query()
//	sessionId := urlQueryParams.Get("id")
//
//	sessionData := getSessionData(sessionId)
//
//	sessionStartTime := sessionData.StartTime
//	sessionEndTime := sessionData.EndTime
//
//
//}

//func GetCategoryWiseAdvancedSessionData() {
//	categoriesData := map[string]map[string]TimeSeriesData{
//		"Stress" : map[string]TimeSeriesData {
//			"Vlf":   make(TimeSeriesData, 0),
//			"Hf":    make(TimeSeriesData, 0),
//			"Rmssd": make(TimeSeriesData, 0),
//			"Pnn50": make(TimeSeriesData, 0),
//		},
//		"OriginalStress" : map[string]TimeseriesData {
//			"Stress": make(TimeSeriesData, 0)
//		},
//		"SlepeData": map[string]TimeseriesData {
//			""
//		}
//	}
//	requiredStressData := map[string]interface{
//		"Vlf":   make(TimeSeriesData, 0),
//		"Hf":    make(TimeSeriesData, 0),
//		"Rmssd": make(TimeSeriesData, 0),
//		"Pnn50": make(TimeSeriesData, 0),
//	}
//}
//
//func fetchAdvancedSessionData() {
//
//}

func getSessionData(sessionId string) Session {
	sessionUrl := fmt.Sprintf("http://35.225.36.244:9804/api/sessions/get/%v", sessionId)

	sessionResponseData := getFromDataStore(sessionUrl)

	var sessionData Session
	err := json.Unmarshal(sessionResponseData, &sessionData)

	if err != nil {
		log.Println("Error unmarshalling response data to sleep data")
	}

	return sessionData
}

func getOrganizationUsers(orgId string) []string {
	url := fmt.Sprintf("http://35.225.36.244:9804/api/org-users/find/?and=orgId:%v&limit=10000", orgId)
	organizationUsersResponseData := getFromDataStore(url)
	var organizationUserData OrganizationUsers
	json.Unmarshal(organizationUsersResponseData, &organizationUserData)

	var orgUsers []string

	for _, orgUser := range organizationUserData {
		orgUsers = append(orgUsers, orgUser.UserId)
	}

	return orgUsers
}

func getOperatorUsers(orgId string) []string {
	url := fmt.Sprintf("http://35.225.36.244:9804/api/op-users/find/?and=orgId:%v&limit=10000", orgId)
	operatorUsersResponseData := getFromDataStore(url)
	var operatorUserData OperatorUsers
	json.Unmarshal(operatorUsersResponseData, &operatorUserData)

	var opUsers []string

	for _, opUser := range operatorUserData {
		opUsers = append(opUsers, opUser.UserId)
	}

	return opUsers
}

func Bod(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

func getDayStart(timestamp int64, timezone string) int64 {
	timezoneObj, _ := time.LoadLocation(timezone)
	timestampTime := time.Unix(timestamp, 0).In(timezoneObj)
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
	resp.Body.Close()
	return responseBody
}

func GetUserSleep(w http.ResponseWriter, r *http.Request) {
	// Get the sessionId from the request body
	// Fetch session details using the datastore link and type as sleep
	// Using the Start Time and End Time, fetch the following
	// 1. Fetch session properties using the sessionId
	// 2. Fetch records using the UserId, Timestamp should be between Session Start Time and Session End Time
	// 3. Fetch events using the UserId, Event Start Time should be between Session Start Time and Session End Time
	// Create Sleep Map Data
	// Return Sleep Map Data through Response
	urlQueryParams := r.URL.Query()
	sessionId := urlQueryParams.Get("id")

	sessionUrl := fmt.Sprintf("http://35.225.36.244:9804/api/sessions/get/%v", sessionId)

	sessionResponseData := getFromDataStore(sessionUrl)

	var sessionData Session
	err := json.Unmarshal(sessionResponseData, &sessionData)

	if err != nil {
		log.Println("Error unmarshalling response data to sleep data")
	}

	sessionStartTime := sessionData.StartTime
	sessionEndTime := sessionData.EndTime

	requiredSessionProperties := map[string]int64{
		"Recovery":           0,
		"Score":              0,
		"SleepTime":          0,
		"WakeupTime":         0,
		"Duration":           0,
		"SleepQualityPoints": 0,
		"SleepRoutinePoints": 0,
		"VitalsPoints":       0,
		"RestfulnessPoints":  0,
	}

	requiredSessionProperties = fetchSessionProperties(sessionId, requiredSessionProperties)

	requiredSessionRecords := map[string]TimeSeriesData{
		"HeartRate":  make(TimeSeriesData, 0),
		"BreathRate": make(TimeSeriesData, 0),
		"Recovery":   make(TimeSeriesData, 0),
		"Stress":     make(TimeSeriesData, 0),
		"Stages":     make(TimeSeriesData, 0),
		"Snoring":    make(TimeSeriesData, 0),
	}

	requiredSessionRecords = fetchSessionRecords(sessionStartTime, sessionEndTime, requiredSessionRecords)

	sessionSleepTime := requiredSessionProperties["SleepTime"]
	sessionWakeupTime := requiredSessionProperties["WakeupTime"]
	sessionScore := requiredSessionProperties["Score"]
	sessionRecovery := requiredSessionProperties["Recovery"]
	sessionHeartRateAverage := getVitalsAverage(requiredSessionRecords["HeartRate"], sessionSleepTime, sessionWakeupTime)
	sessionBreathRateAverage := getVitalsAverage(requiredSessionRecords["BreathRate"], sessionSleepTime, sessionWakeupTime)
	sessionStressAverage := getVitalsAverage(requiredSessionRecords["Stress"], sessionSleepTime, sessionWakeupTime)
	sessionSleepDuration := getTotalDurationFromStages(requiredSessionRecords["Stages"])

	var sessionSessionData SessionSleep
	sessionSessionData.RecoveryValue = sessionRecovery
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
	sessionSessionData.Stages = requiredSessionRecords["Stages"]
	sessionSessionData.Stress = requiredSessionRecords["Stress"]

	responseData := map[string]interface{}{
		"data": sessionSessionData,
	}

	json.NewEncoder(w).Encode(responseData)
}

func GetSessions(w http.ResponseWriter, r *http.Request) {
	// Get the sessionId from the request body
	// Fetch session details using the datastore link and type as sleep
	// A function which fetches the following for the session
	// 1. Last Synced At
	// 2. Score
	// 3. Heart Rate
	// 4. Breath Rate
	// 5. Session Duration
	// 6. Recovery Value

	userSessionsData := getUserSessions(r)

	sessionsSnapshots := make(map[string]SessionSnapshots, 0)

	for _, currentSession := range userSessionsData {
		currentSessionId := currentSession.Id
		currentUserId := currentSession.UserId
		currentSessionType := currentSession.Type
		sessionSnapshotData := getSessionSnapshot(currentSessionId, currentSessionType)
		sessionsSnapshots[currentUserId] = append(sessionsSnapshots[currentUserId], sessionSnapshotData)
	}

	responseData := map[string]interface{}{
		"data": sessionsSnapshots,
	}

	json.NewEncoder(w).Encode(responseData)
}

func GetSessionsSummary(w http.ResponseWriter, r *http.Request) {
	// get days from the url query
	// take current date and then subtract the number of days to get the started_at date
	urlQueryParams := r.URL.Query()
	sDays := urlQueryParams.Get("days")
	days, _ := strconv.ParseInt(sDays, 10, 64)
	endDate := time.Now().Unix()
	startDate := endDate - days*3600*24

	userIdList := getUserList(r)

	generatedSummary := make(map[int64]SessionsSummary, 0)

	for _, currentUserId := range userIdList {
		url := fmt.Sprintf("http://35.225.36.244:9804/api/sessions/find/?and=userId:%v&range=timestamp:%v:%v", currentUserId, startDate, endDate)
		userSessionResponseData := getFromDataStore(url)
		var userSessionsData Sessions
		json.Unmarshal(userSessionResponseData, &userSessionsData)

		for _, session := range userSessionsData {
			currentSessionType := session.Type
			sessionStartTime := session.StartTime
			sessionStartDate := getDayStart(sessionStartTime, "Asia/Kolkata")
			var currentDateSessionSummary SessionsSummary
			if sessionSummary, ok := generatedSummary[sessionStartDate]; ok {
				currentDateSessionSummary = sessionSummary
			} else {
				currentDateSessionSummary = SessionsSummary{}
			}
			if currentSessionType == "Sleep" {
				currentDateSessionSummary.Sleeps++
			} else if currentSessionType == "Meditation" {
				currentDateSessionSummary.Meditations++
			}
		}
	}

	responseData := map[string]interface{}{
		"data": generatedSummary,
	}

	json.NewEncoder(w).Encode(responseData)
}
