package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/logzio/logzio-go"
	receiver "github.com/resdenia/lastpass-receiver"
)

const (
	envNameLastPassURL       = "LASTPASS_URL"
	envNameFromTimestamp     = "FROM_TIMESTAMP"
	envNameInterval          = "INTERVAL"
	envNameCustomerId        = "CUSTOMER_ID"
	envNameLogzioListenerURL = "LOGZIO_LISTENER_URL"
	envNameLogzioToken       = "LOGZIO_TOKEN"
	lastPassApiKey           = "LASTPASS_KEY"
	defaultInterval          = 5
	defaultLogzioListenerURL = "https://listener.logz.io:8071"
	enterpriseUrl            = "https://lastpass.com/enterpriseapi.php"
	timeFormat               = "2006-01-02 15:04:05"
)

var (
	debugLogger = log.New(os.Stderr, "DEBUG: ", log.Ldate|log.Ltime)
	infoLogger  = log.New(os.Stderr, "INFO: ", log.Ldate|log.Ltime)
	errorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime)
)

type lastPassCollector struct {
	receiver *receiver.LastPassLogsReceiver
	shipper  *logzio.LogzioSender
	interval int
}

func goDotEnvVariable(key string) string {

	// load .env file
	err := godotenv.Load(".env")

	if err != nil {
		errorLogger.Println("Error loading .env file")
	}

	return os.Getenv(key)
}

func newLastPassCollector() (*lastPassCollector, error) {
	rec, err := createLastPassReceiver()
	if err != nil {
		return nil, fmt.Errorf("error creating LastPass receiver: %w", err)
	}

	shipper, err := createLogzioSender()
	if err != nil {
		return nil, fmt.Errorf("error creating Logz.io sender: %w", err)
	}

	intervalStr := goDotEnvVariable(envNameInterval)
	interval, err := strconv.Atoi(intervalStr)
	if err != nil {
		infoLogger.Println("Interval is not a number. Used default value -", defaultInterval, "seconds")
		interval = defaultInterval
	}

	if interval <= 0 {
		infoLogger.Println("Interval is not a positive number. Used default value -", defaultInterval, "seconds")
		interval = defaultInterval
	}

	return &lastPassCollector{
		receiver: rec,
		shipper:  shipper,
		interval: interval,
	}, nil
}

func createLastPassReceiver() (*receiver.LastPassLogsReceiver, error) {

	rec, err := receiver.NewLastPassLogsReceiver(
		enterpriseUrl,
		goDotEnvVariable(lastPassApiKey),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating Salesforce logs receiver object: %w", err)
	}

	return rec, nil
}

func createLogzioSender() (*logzio.LogzioSender, error) {
	logzioListenerURL := goDotEnvVariable(envNameLogzioListenerURL)
	if logzioListenerURL == "" {
		logzioListenerURL = defaultLogzioListenerURL
	}

	logzioToken := goDotEnvVariable(envNameLogzioToken)
	if logzioToken == "" {
		return nil, fmt.Errorf("Logz.io token must have a value")
	}

	shipper, err := logzio.New(
		fmt.Sprintf("%s&type=lastPass", logzioToken),
		logzio.SetDebug(os.Stderr),
		logzio.SetUrl(logzioListenerURL),
		logzio.SetDrainDuration(time.Second*5),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating Logz.io sender object: %w", err)
	}

	return shipper, nil
}

func (sfc *lastPassCollector) sendDataToLogzio(data []byte) bool {
	if err := sfc.shipper.Send(data); err != nil {
		errorLogger.Println("error sending lastPass report log ", err)
		return false
	}

	return true
}

func (sfc *lastPassCollector) collect(lastTime string) {
	var waitGroup sync.WaitGroup
	customerIDStr := goDotEnvVariable(envNameCustomerId)
	customerId, err := strconv.Atoi(customerIDStr)
	if err != nil {
		// print it out
		errorLogger.Println(err)
	}
	go func() {
		defer waitGroup.Done()

		logsToSend, err := sfc.receiver.GetLogs(goDotEnvVariable(lastPassApiKey), lastTime, customerId)
		if err != nil {
			// print it out
			errorLogger.Println(err)
		}

		waitGroup.Add(1)
		for _, log := range logsToSend {
			byteLog, _ := json.Marshal(log)

			sfc.sendDataToLogzio(byteLog)
		}
		dataLastTime := []byte(lastTime)

		// the WriteFile method returns an error if unsuccessful
		err = ioutil.WriteFile("lastTime.txt", dataLastTime, 0777)
		// handle this error
		if err != nil {
			// print it out
			errorLogger.Println(err)
		}
	}()

	waitGroup.Wait()
}

func main() {
	collector, err := newLastPassCollector()
	if err != nil {
		panic(err)
	}
	lastTime := time.Now().Format(timeFormat)
	for {
		collector.collect(lastTime)
		debugLogger.Println("Finished collecting. Collector will run in", collector.interval, "seconds")
		lastTime = time.Now().Format(timeFormat)

		time.Sleep(time.Duration(collector.interval) * time.Second)
	}
}
