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

// const enterpriseUrl = "https://lastpass.com/enterpriseapi.php"

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
	timestamp  string
}
type JsonData struct {
	Status string         `json:"status"`
	Next   string         `json:"next"`
	Data   map[string]Log `json:"data"`
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

type LastPassLogsReceiver struct {
	securityToken string
	client        http.Client
	currentTime   string
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

	currentTime := time.Now().Format("2006-01-02 15:04:05")

	client := http.Client{}

	return &LastPassLogsReceiver{
		securityToken: securityToken,
		client:        client,
		currentTime:   currentTime,
	}, nil
}

func (slr *LastPassLogsReceiver) GetLogs(lastPassApiKey string, lastTimeEvent string, customerId int) ([]LogToSend, error) {

	var data struct {
		Status string         `json:"status"`
		Next   string         `json:"next"`
		Data   map[string]Log `json:"data"`
	}

	arrtoSend := fmt.Sprintf(`{"cid": %d,"provhash": "%s","cmd": "reporting","data": {"from": "%s","to": "%s"}}`, customerId, lastPassApiKey, lastTimeEvent, slr.currentTime)
	enterpriseUrl := os.Getenv("LASTPASS_URL")

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
	err = retry.Do(
		func() error {
			resp, err := client.Do(req)
			if err != nil {
				return err
			}

			defer resp.Body.Close()

			body, _ := ioutil.ReadAll(resp.Body)
			errr := json.Unmarshal(body, &data)
			if errr != nil {
				panic(errr)
			}

			json.NewDecoder(resp.Body).Decode(&data)

			if err != nil {
				fmt.Println(err)
				panic(err)
			}
			for key, value := range data.Data {
				var logToSend LogToSend = parseLog(key, value)
				addTimestamp(&logToSend)
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

func addTimestamp(logToSend *LogToSend) {
	times := fmt.Sprintf("%s.000Z", strings.Replace(logToSend.Time, " ", "T", 1))
	logToSend.timestamp = times
}
