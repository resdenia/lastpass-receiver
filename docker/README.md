# LastPass Logz.io

Send report logs from LastPass to logz.io.

## Usage

```shell
docker pull logzio/logzio-salesforce-collector
```

## Run The Docker Container

```shell
 docker run --name logzio-salesforce-collector \
 --env PROVISIONING_HASH="<<PROVISIONING_HASH>>" \
 --env CUSTOM_FIELDS="<<CUSTOM_FIELDS>>" \
 --env FROM_TIMESTAMP="<<FROM_TIMESTAMP>>" \
 --env INTERVAL="<<INTERVAL>>" \
 --env LOGZIO_LISTENER_URL="<<LOGZIO_LISTENER_URL>>" \
 --env LOGZIO_TOKEN="<<LOGZIO_TOKEN>>" \
logzio/logzio-lastpass-collector
```

### Environment Variables

| Name                | Description                                                                                                | Required? | Default                       |
| ------------------- | ---------------------------------------------------------------------------------------------------------- | --------- | ----------------------------- |
| CLIENT_ID           | Salesforce App Client ID.                                                                                  | Yes       | -                             |
| USERNAME            | Salesforce account username (your email)                                                                   | Yes       | -                             |
| PASSWORD            | Salesforce account password                                                                                | Yes       | -                             |
| SECURITY_TOKEN      | Salesforce account security token                                                                          | Yes       | -                             |
| SOBJECT_TYPES       | List of sObject types to collect. Each type must be separated by comma, for example: "TYPE1,TYPE2,TYPE3".  | Yes       | -                             |
| LOGZIO_TOKEN        | Logz.io logs token.                                                                                        | Yes       | -                             |
| SALESFORCE_URL      | Salesforce URL.                                                                                            | No        | https://login.salesforce.com  |
| API_VERSION         | Salesforce API version.                                                                                    | No        | 55.0                          |
| CUSTOM_FIELDS       | Custom fields to add to each sObject data. Must be in the following pattern: "FIELD1:VALUE1,FIELD2:VALUE2" | No        | -                             |
| FROM_TIMESTAMP      | Timestamp from when to collect data. Must be in the following format: 2006-01-02T15:04:05.000Z .           | No        | Current time minus 1 hour     |
| INTERVAL            | The time interval to collect Salesforce data (in seconds).                                                 | No        | 5 (seconds)                   |
| LOGZIO_LISTENER_URL | Logz.io listener logs URL.                                                                                 | No        | https://listener.logz.io:8071 |

## Searching in Logz.io

All logs will have the type `salesforce`.

## Changelog

1.0.0 - Initial Release

## Licease
