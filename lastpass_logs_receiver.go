package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/avast/retry-go"
	// "github.com/simpleforce/simpleforce"
)

type Log struct {
	Time       string
	Username   string
	IP_Address string
	Action     string
	Data       string
}
type LogToSend struct {
	Name       string
	Time       string
	Username   string
	IP_Address string
	Action     string
	Data       string
}
type JsonData struct {
	Status string         `json:"status"`
	Next   string         `json:"next"`
	Data   map[string]Log `json:"data"`
}

type handlerConfig struct {
	token      string
	url        string
	customerId int
	hashToken  string
}

type logzioHandler struct {
	config     handlerConfig
	httpClient *http.Client
	dataBuffer bytes.Buffer
}

func lastpass_logs_receiver() []LogToSend {
	var data struct {
		Status string         `json:"status"`
		Next   string         `json:"next"`
		Data   map[string]Log `json:"data"`
	}

	lastTimeEvent := ""
	lastPassApiKey := os.Getenv("LASTPASS_KEY")
	customerId := os.Getenv("CUSTOMER_ID")

	arrtoSend := fmt.Sprintf(`{
		"cid": %s,
		"provhash": "%s",
		"cmd": "reporting",
		"data": {
			"from": "%s",
			"to": "%s"
		}
		}`, customerId, lastPassApiKey, lastTimeEvent, time.Now())
	jsonStr := []byte(arrtoSend)

	req, err := http.NewRequest(http.MethodPost, enterpriseUrl, bytes.NewReader(jsonStr))
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	errr := json.Unmarshal(body, &data)
	if errr != nil {
		panic(errr)
	}
	// Marshal back to json (as original)
	// out, _ := json.Marshal(&data)
	// fmt.Println(out))
	json.NewDecoder(resp.Body).Decode(&data)

	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	dataToSend := []LogToSend{}

	for key, value := range data.Data {
		fmt.Println(key)

		var logToSend LogToSend = parseLog(key, value)
		dataToSend = append(dataToSend, logToSend)
		fmt.Println(key)

	}
	return dataToSend

}

func parseLog(logName string, fields Log) LogToSend {
	var logToSend LogToSend
	logToSend.Name = logName
	logToSend.Action = fields.Action
	logToSend.Time = fields.Time
	logToSend.IP_Address = fields.IP_Address
	logToSend.Data = fields.Data
	logToSend.Username = fields.Username

	return logToSend

}

const (
	EventLogFileSObjectName = "eventlogfile"
	defaultApiVersion       = "55.0"
)

type LastpassLogsReceiver struct {
	securityToken           string
	client                  http.Client
	currentTimeMinusOneHour string
}

type SObjectToCollect struct {
	SObjectType     string
	LatestTimestamp string
}

func NewLastpassLogsReceiver(
	url string,
	securityToken string,
) (*LastpassLogsReceiver, error) {
	if securityToken == "" {
		return nil, fmt.Errorf("security token must have a value")
	}

	currentTimeMinusOneHour := time.Now().Format("2006-01-02 15:04:05")

	client := http.Client{}

	return &LastpassLogsReceiver{
		securityToken:           securityToken,
		client:                  client,
		currentTimeMinusOneHour: currentTimeMinusOneHour,
	}, nil
}

// func (slr *LastpassLogsReceiver) GetSObjectRecords() ([]LogToSend, error) {
// 	query := fmt.Sprintf("SELECT Id,CreatedDate FROM %s WHERE CreatedDate > %s ORDER BY CreatedDate", sObject.SObjectType, sObject.LatestTimestamp)
// 	result, err := slr.client.Query(query)
// 	if err != nil {
// 		return nil, fmt.Errorf("error querying Lastpass API: %w", err)
// 	}

// 	debugLogger.Println("Got", len(result.Records), "records of sObject", sObject.SObjectType)
// 	return result.Records, nil
// }

// func (slr *LastpassLogsReceiver) CollectSObjectRecord(record *simpleforce.SObject) ([]byte, *string, error) {
// 	id := record.ID()
// 	data := record.Get(id)

// 	jsonData, err := json.Marshal(data)
// 	if err != nil {
// 		return nil, nil, fmt.Errorf("error marshaling data from Lastpass API: %w", err)
// 	}

// 	// jsonData, err = slr.addCustomFields(jsonData, record.Type(), id)
// 	if err != nil {
// 		return nil, nil, fmt.Errorf("error adding custom fields to data: %w", err)
// 	}

// 	createdDate := record.StringField("CreatedDate")
// 	createdDate = strings.Replace(createdDate, "+0000", "Z", 1)

// 	debugLogger.Println("Collected data of sObject", record.Type(), "record ID", id)
// 	return jsonData, &createdDate, nil
// }

// func (slr *LastpassLogsReceiver) EnrichEventLogFileSObjectData(data *simpleforce.SObject, jsonData []byte) ([][]byte, error) {
// 	eventLogRows, err := slr.getEventLogFileContent(data)
// 	if err != nil {
// 		return nil, fmt.Errorf("error getting EventLogFile sObject log file content: %w", err)
// 	}

// 	debugLogger.Println("Got", len(eventLogRows), "logs from EventLogFile sObject ID", data.ID())

// 	var jsonsData [][]byte
// 	for _, eventLogRow := range eventLogRows {
// 		newJsonData, err := addEventLogToJsonData(eventLogRow, jsonData)
// 		if err != nil {
// 			return nil, fmt.Errorf("error adding event log content to JSON data: %w", err)
// 		}

// 		jsonsData = append(jsonsData, newJsonData)
// 	}

// 	debugLogger.Println("Enriched sObject data with", len(jsonsData), "logs from EventLogFile sObject ID", data.ID())
// 	return jsonsData, nil
// }

// func (slr *LastpassLogsReceiver) getEventLogFileContent(data *simpleforce.SObject) ([]map[string]interface{}, error) {
// 	apiPath := data.StringField("LogFile")
// 	logFileContent, err := slr.getFileContent(apiPath)
// 	if err != nil {
// 		return nil, fmt.Errorf("error getting event log file content: %w", err)
// 	}

// 	trimmedLogFileContent := strings.Replace(string(logFileContent), "\n\n", "\n", -1)
// 	debugLogger.Println("Got EventLogFile sObject log file content ID", data.ID())

// 	reader := strings.NewReader(trimmedLogFileContent)
// 	csvReader := csv.NewReader(reader)

// 	csvData, err := csvReader.ReadAll()
// 	if err != nil {
// 		return nil, fmt.Errorf("error reading CSV data: %w", err)
// 	}

// 	var logEvents []map[string]interface{}
// 	for rowIndex, row := range csvData {
// 		if rowIndex == 0 {
// 			continue
// 		}

// 		logEvent := make(map[string]interface{})
// 		for fieldIndex, field := range row {
// 			key := csvData[0][fieldIndex]
// 			logEvent[key] = field
// 		}

// 		logEvents = append(logEvents, logEvent)
// 	}

// 	return logEvents, nil
// }

func (slr *LastpassLogsReceiver) getLogs(lastPassApiKey string) ([]byte, error) {
	// httpClient := &http.Client{}
	// req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", strings.TrimRight(slr.client.GetLoc(), "/"), apiPath), nil)
	// req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	// req.Header.Add("Accept", "application/json")
	// req.Header.Add("Authorization", "Bearer "+slr.client.GetSid())

	var data struct {
		Status string         `json:"status"`
		Next   string         `json:"next"`
		Data   map[string]Log `json:"data"`
	}

	lastTimeEvent := ""

	// lastPassApiKey := os.Getenv("LASTPASS_KEY")
	customerId := os.Getenv("CUSTOMER_ID")
	enterpriseUrl := ""
	arrtoSend := fmt.Sprintf(`{
		"cid": %s,
		"provhash": "%s",
		"cmd": "reporting",
		"data": {
			"from": "%s",
			"to": "%s"
		}
		}`, customerId, lastPassApiKey, lastTimeEvent, time.Now())
	jsonStr := []byte(arrtoSend)

	req, err := http.NewRequest(http.MethodPost, enterpriseUrl, bytes.NewReader(jsonStr))
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	var resp *http.Response
	err = retry.Do(
		func() error {
			resp, err := client.Do(req)
			if err != nil {
				return err
			}

			// fmt.Println(resp.Body)
			defer resp.Body.Close()

			body, _ := ioutil.ReadAll(resp.Body)
			// fmt.Println("response Body:", string(body))
			errr := json.Unmarshal(body, &data)
			if errr != nil {
				panic(errr)
			}
			// Marshal back to json (as original)
			// out, _ := json.Marshal(&data)
			// fmt.Println(out))
			json.NewDecoder(resp.Body).Decode(&data)

			if err != nil {
				fmt.Println(err)
				panic(err)
			}
			dataToSend := []LogToSend{}
			for key, value := range data.Data {
				fmt.Println(key)

				var logToSend LogToSend = parseLog(key, value)
				dataToSend = append(dataToSend, logToSend)
				fmt.Println(key)

			}

			if resp.StatusCode < 200 || resp.StatusCode > 299 {
				buf := new(bytes.Buffer)
				buf.ReadFrom(resp.Body)
				return fmt.Errorf("ERROR: statuscode: %d, body: %s", resp.StatusCode, buf.String())
			}

			return nil
		},
		retry.RetryIf(
			func(err error) bool {
				result, matchErr := regexp.MatchString("statuscode: 5[0-9]{2}", err.Error())
				if matchErr != nil {
					return false
				}
				if result {
					return true
				}

				return false
			}),
		retry.DelayType(retry.BackOffDelay),
		retry.Attempts(3),
	)
	if err != nil {
		return nil, err
	}

	var content []byte
	content, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return content, nil
}

func addEventLogToJsonData(eventLog map[string]interface{}, jsonData []byte) ([]byte, error) {
	var jsonMap map[string]interface{}
	if err := json.Unmarshal(jsonData, &jsonMap); err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON data: %w", err)
	}

	jsonMap["LogFileContent"] = eventLog

	newJsonData, err := json.Marshal(jsonMap)
	if err != nil {
		return nil, fmt.Errorf("error marshaling JSON data: %w", err)
	}

	return newJsonData, nil
}
