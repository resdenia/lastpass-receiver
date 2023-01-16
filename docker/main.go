package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/logzio/logzio-go"
	receiver "github.com/resdenia/lastpass-receiver"
)

const (
	envNameLastPassURL       = "LASTPASS_URL"
	envNameFromTimestamp     = "FROM_TIMESTAMP"
	envNameInterval          = "INTERVAL"
	envNameLogzioListenerURL = "LOGZIO_LISTENER_URL"
	envNameLogzioToken       = "LOGZIO_TOKEN"
	lastPassApiKey           = "LAST_PASS_API_KEY"
	defaultInterval          = 5
	defaultLogzioListenerURL = "https://listener.logz.io:8071"
	enterpriseUrl            = "https://lastpass.com/enterpriseapi.php"
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

func newLastPassCollector() (*lastPassCollector, error) {
	rec, err := createLastPassReceiver()
	if err != nil {
		return nil, fmt.Errorf("error creating Salesforce receiver: %w", err)
	}

	shipper, err := createLogzioSender()
	if err != nil {
		return nil, fmt.Errorf("error creating Logz.io sender: %w", err)
	}

	intervalStr := os.Getenv(envNameInterval)
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
	// sObjectTypesStr := os.Getenv(envNameSObjectTypes)
	// sObjectTypes := strings.Split(strings.Replace(sObjectTypesStr, " ", "", -1), ",")
	// latestTimestamp := os.Getenv(envNameFromTimestamp)

	// var sObjects []*receiver.SObjectToCollect
	// for _, sObjectType := range sObjectTypes {
	// 	sObjects = append(sObjects, &receiver.SObjectToCollect{
	// 		SObjectType:     sObjectType,
	// 		LatestTimestamp: latestTimestamp,
	// 	})
	// }

	// customFieldsStr := os.Getenv(envNameCustomFields)
	// customFields := make(map[string]string)

	// if customFieldsStr != "" {
	// 	fields := strings.Split(customFieldsStr, ",")

	// 	for _, field := range fields {
	// 		if !strings.Contains(field, ":") {
	// 			return nil, fmt.Errorf("each field in %s must have ':' separator between the field key and value", envNameCustomFields)
	// 		}

	// 		fieldKeyAndValue := strings.Split(field, ":")
	// 		customFields[fieldKeyAndValue[0]] = fieldKeyAndValue[1]
	// 	}
	// }

	rec, err := receiver.NewLastPassLogsReceiver(
		enterpriseUrl,
		os.Getenv(lastPassApiKey),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating Salesforce logs receiver object: %w", err)
	}

	return rec, nil
}

func createLogzioSender() (*logzio.LogzioSender, error) {
	logzioListenerURL := os.Getenv(envNameLogzioListenerURL)
	if logzioListenerURL == "" {
		logzioListenerURL = defaultLogzioListenerURL
	}

	logzioToken := os.Getenv(envNameLogzioToken)
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
		errorLogger.Println("error sending sObject ", err)
		return false
	}

	return true
}

func (sfc *lastPassCollector) collect(lastTime string) {
	var waitGroup sync.WaitGroup

	// for _, sObject := range sfc.receiver.SObjects {
	// 	debugLogger.Println("sObject type:", sObject.SObjectType, "- from timestamp:", sObject.LatestTimestamp)
	// 	waitGroup.Add(1)

	// 	go func(sObject *receiver.SObjectToCollect) {
	// 		defer waitGroup.Done()

	// 		records, err := sfc.receiver.GetSObjectRecords(sObject)
	// 		if err != nil {
	// 			errorLogger.Println("error getting sObject ", sObject.SObjectType, " records: ", err)
	// 			return
	// 		}

	// 		for _, record := range records {
	// 			data, createdDate, err := sfc.receiver.CollectSObjectRecord(&record)
	// 			if err != nil {
	// 				errorLogger.Println("error collecting sObject ", sObject.SObjectType, " record ID ", record.ID(), ": ", err)
	// 				return
	// 			}

	// 			if strings.ToLower(sObject.SObjectType) == receiver.EventLogFileSObjectName {
	// 				enrichedData, err := sfc.receiver.EnrichEventLogFileSObjectData(&record, data)
	// 				if err != nil {
	// 					errorLogger.Println("error enriching EventLogFile sObject ", " record ID ", record.ID(), ": ", err)
	// 					return
	// 				}

	// 				for _, data = range enrichedData {
	// 					if !sfc.sendDataToLogzio(data, sObject.SObjectType, record.ID()) {
	// 						return
	// 					}
	// 				}
	// 			} else {
	// 				if !sfc.sendDataToLogzio(data, sObject.SObjectType, record.ID()) {
	// 					return
	// 				}
	// 			}

	// 			sObject.LatestTimestamp = *createdDate
	// 		}
	// 	}(sObject)
	// }

	logsToSend := sfc.receiver.GetLogs(lastPassApiKey, lastTime)
	for _, log := range logsToSend {

	}
	dataLastTime := []byte(lastTime)

	// the WriteFile method returns an error if unsuccessful
	err := ioutil.WriteFile("lastTime.txt", dataLastTime, 0777)
	// handle this error
	if err != nil {
		// print it out
		fmt.Println(err)
	}
	waitGroup.Wait()
}

func main() {
	collector, err := newLastPassCollector()
	if err != nil {
		panic(err)
	}
	lastTime := time.Now().Format("2006-01-02 15:04:05")

	for {
		collector.collect(lastTime)
		debugLogger.Println("Finished collecting. Collector will run in", collector.interval, "seconds")
		lastTime = time.Now().Format("2006-01-02 15:04:05")
		time.Sleep(time.Duration(collector.interval) * time.Second)
	}
}
