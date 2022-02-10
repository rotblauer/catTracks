#!/usr/bin/env bash

set -e

main() {

    curl "https://api.catonmap.info/catsnaps?tstart=$(($(date +%s) - 7*24*60*60))" > /tmp/catsnaps.json

    # {"uuid":"3582fb4e0c347601","pushToken":"","version":"gcps/v0.0.0+1","id":1640127468000000000,"name":"sofia-moto-fdb7","lat":48.0432994,"long":-118.3448815,"accuracy":3.6,"vAccuracy":2.5,"elevation":368,"speed":0,"tilt":0,"heading":239,"heartrate":0,"time":"2021-12-21T22:57:48Z","floor":0,"notes":"{\"activity\":\"Walking\",\"numberOfSteps\":34761,\"averageActivePace\":0,\"currentPace\":0,\"currentCadence\":0,\"distance\":0,\"customNote\":\"\",\"floorsAscended\":0,\"floorsDescended\":0,\"currentTripStart\":\"0001-01-01T00:00:00Z\",\"pressure\":0,\"visit\":\"\",\"heartRateS\":\"\",\"heartRateRawS\":\"\",\"batteryStatus\":\"{\\\"level\\\":0.24,\\\"status\\\":\\\"unplugged\\\"}\",\"networkInfo\":\"\",\"imgb64\":\"\",\"imgS3\":\"rotblauercatsnaps/WfJMAWGAZHdnDGdTUBziNAkiKZsAnodn\"}","COVerified":true,"remoteaddr":""},{"uuid":"3582fb4e0c347601","pushToken":"","version":"gcps/v0.0.0+1","id":1640127595000000000,"name":"sofia-moto-fdb7","lat":48.0425815,"long":-118.3442455,"accuracy":4.3,"vAccuracy":5.5,"elevation":375.3,"speed":0,"tilt":0,"heading":91,"heartrate":0,"time":"2021-12-21T22:59:55Z","floor":0,"notes":"{\"activity\":\"Bike\",\"numberOfSteps\":34866,\"averageActivePace\":0,\"currentPace\":0,\"currentCadence\":0,\"distance\":0,\"customNote\":\"\",\"floorsAscended\":0,\"floorsDescended\":0,\"currentTripStart\":\"0001-01-01T00:00:00Z\",\"pressure\":0,\"visit\":\"\",\"heartRateS\":\"\",\"heartRateRawS\":\"\",\"batteryStatus\":\"{\\\"level\\\":0.23,\\\"status\\\":\\\"unplugged\\\"}\",\"networkInfo\":\"\",\"imgb64\":\"\",\"imgS3\":\"rotblauercatsnaps/XgUujqrDonKeMlFUhVpNDTQmZwhIUImT\"}","COVerified":true,"remoteaddr":""},{"uuid":"F3FAE270-8BE6-4A16-8320-07B66578B722","pushToken":"unset","version":"V.customizableCatTrackHat","id":1640288638232000000,"name":"Rye8","lat":46.84600830078125,"long":-92.03269958496094,"accuracy":0,"vAccuracy":0,"elevation":0,"speed":-1,"tilt":0,"heading":-1,"heartrate":0,"time":"2021-12-23T19:43:58.232Z","floor":0,"notes":"{\"activity\":\"Unknown\",\"numberOfSteps\":2174,\"averageActivePace\":0.6070001057081406,\"currentPace\":0.694575846195221,\"currentCadence\":2.2329909801483154,\"distance\":2110.764665362425,\"customNote\":\"\",\"floorsAscended\":34,\"floorsDescended\":1,\"currentTripStart\":\"2021-12-23T18:51:05.046Z\",\"pressure\":96.22677612304688,\"visit\":\"{\\\"validVisit\\\":false}\",\"heartRateS\":\"\",\"heartRateRawS\":\"\",\"batteryStatus\":\"{\\\"level\\\":0.75,\\\"status\\\":\\\"unplugged\\\"}\",\"networkInfo\":\"{}\",\"imgb64\":\"\",\"imgS3\":\"rotblauercatsnaps/PEQmEfgLHVupLtTnhZlpeTQJoEYXmTNH\"}","COVerified":true,"remoteaddr":""},{"uuid":"F3FAE270-8BE6-4A16-8320-07B66578B722","pushToken":"unset","version":"V.customizableCatTrackHat","id":1640288653065000000,"name":"Rye8","lat":46.845977783203125,"long":-92.03270721435547,"accuracy":0,"vAccuracy":0,"elevation":0,"speed":-1,"tilt":0,"heading":-1,"heartrate":0,"time":"2021-12-23T19:44:13.065Z","floor":0,"notes":"{\"activity\":\"Unknown\",\"numberOfSteps\":2174,\"averageActivePace\":0.6070001057081406,\"currentPace\":0.694575846195221,\"currentCadence\":2.2329909801483154,\"distance\":2110.764665362425,\"customNote\":\"\",\"floorsAscended\":34,\"floorsDescended\":1,\"currentTripStart\":\"2021-12-23T18:51:05.046Z\",\"pressure\":96.2248764038086,\"visit\":\"{\\\"validVisit\\\":false}\",\"heartRateS\":\"\",\"heartRateRawS\":\"\",\"batteryStatus\":\"{\\\"level\\\":0.74000000953674316,\\\"status\\\":\\\"unplugged\\\"}\",\"networkInfo\":\"{}\",\"imgb64\":\"\",\"imgS3\":\"rotblauercatsnaps/KBroUvoVSikTbNooZEOkwNplXgHtMHln\"}","COVerified":true,"remoteaddr":""}]

    cat /tmp/catsnaps.json | jq '.[]|.notes' | sed 's/.*catsnaps\///gm' | sed 's/\\.*//g' > /tmp/catsnaps-ids.txt

    # cat catsnaps.json | jq '.[]|.notes' | sed 's/.*catsnaps\///gm' | sed 's/\\.*//g'
    # FXRcJobhRZBpVDjEMpAlJLxQAqYkjNoB
    # LOGdDcwdfHXrzrnRStcBuFbStJGCQMFL
    # pOifIntKloTGIAmJQdUaAgwyXFNQCzft
    # fEZfqhtYeteTGHQiqjywBswLuVuzysJZ
    # okejRZVzbJNDCQAZHPWqnbLafvXLWRfR
    # MaYLMIKZmgFfKaGRxyudtzKSWDANosru
    # zJOLHlTUZYEIYQTEQqfReLVASEAwYMmH
    # uNobMXvcffVrrGLQBRvCNcEXnfSJodzg
    # JDNnQOGzGqZSFpmZFSIitUfbgQksLCWs
    # vsiQacPpgzrvwgLgBXLQdFGGeXDtFLqm
    # erWbCMVcIdslbQSstbNABvnGCZNsxcCj
    # rBbBhqXepfVSUbrRlWLcXgPJAlCQuAIz
    # gUaekscQOIDNbTNQolvEpIVNPDYhPSPP
    # WfJMAWGAZHdnDGdTUBziNAkiKZsAnodn
    # XgUujqrDonKeMlFUhVpNDTQmZwhIUImT
    # PEQmEfgLHVupLtTnhZlpeTQJoEYXmTNH
    # KBroUvoVSikTbNooZEOkwNplXgHtMHln

    [[ $(whoami) == "pi" ]] && mkdir -p "${HOME}/shared/pictures/tv/"

    cat /tmp/catsnaps-ids.txt | while read -r line; do

        # https://s3.us-east-2.amazonaws.com/rotblauercatsnaps/FXRcJobhRZBpVDjEMpAlJLxQAqYkjNoB

        # If the file (image) doesn't already exist, download it.
        #
        echo "https://s3.us-east-2.amazonaws.com/rotblauercatsnaps/${line}"
        if [[ ! -f "/home/pi/shared/pictures/tv/catsnap_${line}.png" ]]; then
            echo "/home/pi/shared/pictures/tv/catsnap_${line}.png Does not yet exist locally, downloading..."
            wget -O "/home/pi/shared/pictures/tv/catsnap_${line}.png" "https://s3.us-east-2.amazonaws.com/rotblauercatsnaps/${line}"
        else
            echo "/home/pi/shared/pictures/tv/catsnap_${line}.png Exists locally, skipping"
        fi
    done
}

main
