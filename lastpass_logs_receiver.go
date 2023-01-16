package lastpass_logs_receiver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/avast/retry-go"
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

// func lastpass_logs_receiver() []LogToSend {
// 	var data struct {
// 		Status string         `json:"status"`
// 		Next   string         `json:"next"`
// 		Data   map[string]Log `json:"data"`
// 	}

// 	lastTimeEvent := ""
// 	lastPassApiKey := os.Getenv("LASTPASS_KEY")
// 	customerId := os.Getenv("CUSTOMER_ID")
// 	enterpriseUrl := os.Getenv("LASTPASS_URL")

// 	arrtoSend := fmt.Sprintf(`{
// 		"cid": %s,
// 		"provhash": "%s",
// 		"cmd": "reporting",
// 		"data": {
// 			"from": "%s",
// 			"to": "%s"
// 		}
// 		}`, customerId, lastPassApiKey, lastTimeEvent, time.Now())
// 	jsonStr := []byte(arrtoSend)

// 	req, err := http.NewRequest(http.MethodPost, enterpriseUrl, bytes.NewReader(jsonStr))
// 	if err != nil {
// 		fmt.Println(err)
// 		panic(err)
// 	}
// 	req.Header.Set("Content-Type", "application/json")

// 	client := http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		fmt.Println(err)
// 		panic(err)
// 	}
// 	defer resp.Body.Close()

// 	body, _ := ioutil.ReadAll(resp.Body)
// 	errr := json.Unmarshal(body, &data)
// 	if errr != nil {
// 		panic(errr)
// 	}
// 	// Marshal back to json (as original)
// 	// out, _ := json.Marshal(&data)
// 	// fmt.Println(out))
// 	json.NewDecoder(resp.Body).Decode(&data)

// 	if err != nil {
// 		fmt.Println(err)
// 		panic(err)
// 	}

// 	dataToSend := []LogToSend{}

// 	for key, value := range data.Data {
// 		fmt.Println(key)

// 		var logToSend LogToSend = parseLog(key, value)
// 		dataToSend = append(dataToSend, logToSend)
// 		fmt.Println(key)

// 	}
// 	return dataToSend

// }

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

type LastPassLogsReceiver struct {
	securityToken           string
	client                  http.Client
	currentTimeMinusOneHour string
}

type RequestBody struct {
	cid      int
	provhash string
	cmd      string
	data     FilterDate
}
type FilterDate struct {
	from string
	to   string
}

func NewLastPassLogsReceiver(
	url string,
	securityToken string,
) (*LastPassLogsReceiver, error) {
	if securityToken == "" {
		return nil, fmt.Errorf("security token must have a value")
	}

	currentTimeMinusOneHour := time.Now().Format("2006-01-02 15:04:05")

	client := http.Client{}

	return &LastPassLogsReceiver{
		securityToken:           securityToken,
		client:                  client,
		currentTimeMinusOneHour: currentTimeMinusOneHour,
	}, nil
}

func (slr *LastPassLogsReceiver) GetLogs(lastPassApiKey string, lastTimeEvent string) ([]LogToSend, error) {
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

	// lastPassApiKey := os.Getenv("LASTPASS_KEY")
	customerId := os.Getenv("CUSTOMER_ID")
	// LASTPASS_API_URL = "https://lastpass.com/enterpriseapi.php"

	enterpriseUrl := os.Getenv("LASTPASS_URL")
	arrtoSend := fmt.Sprintf(`{
		"cid": %s,
		"provhash": "%s",
		"cmd": "reporting",
		"data": {
			"from": "%s",
			"to": "%s",
		},
		}`, customerId, lastPassApiKey, lastTimeEvent, time.Now())
	// requestBody := RequestBody{}
	// json.Unmarshal([]byte(arrtoSend), &requestBody)

	// payloadBuf := new(bytes.Buffer)
	// err := json.NewEncoder(payloadBuf).Encode([]byte(arrtoSend))
	// if err != nil {
	// 	fmt.Println(err)
	// }

	req, err := http.NewRequest(http.MethodPost, enterpriseUrl, strings.NewReader(arrtoSend))
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
	dataToSend := []LogToSend{}
	// var resp *http.Response
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
			for key, value := range data.Data {
				var logToSend LogToSend = parseLog(key, value)
				dataToSend = append(dataToSend, logToSend)

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

	return dataToSend, nil
}

// func addEventLogToJsonData(eventLog map[string]interface{}, jsonData []byte) ([]byte, error) {
// 	var jsonMap map[string]interface{}
// 	if err := json.Unmarshal(jsonData, &jsonMap); err != nil {
// 		return nil, fmt.Errorf("error unmarshaling JSON data: %w", err)
// 	}

// 	jsonMap["LogFileContent"] = eventLog

// 	newJsonData, err := json.Marshal(jsonMap)
// 	if err != nil {
// 		return nil, fmt.Errorf("error marshaling JSON data: %w", err)
// 	}

// 	return newJsonData, nil
// }
