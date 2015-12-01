package protocol

import (
    "errors"
    "time"
    "strings"
    "encoding/json"
)

type timeParser struct {
    Value time.Time
}

func (t *timeParser) UnmarshalJSON (raw []byte) error {
    var err error
    value := strings.Trim(string(raw), `"`)
    t.Value, err = time.Parse("2006-01-02 15:04:05 GMT", value)
    if err != nil { t.Value, err = time.Parse(time.RFC3339, value) }
    if err != nil { t.Value, err = time.Parse(time.RFC3339Nano, value) }
    if err != nil { return errors.New("Unkown date format. Unable to parse time") }
    return nil
}

func decodePayload (raw []byte) (error, *Payload) {
    payload := &Payload{raw, nil, nil, nil}
    timeStruct := &struct{
        Stat *struct{ Time timeParser `json:"time"` } `json:"stat"`
        RXPK *[]struct{ Time timeParser `json:"time"` } `json:"rxpk"`
        TXPK *struct{ Time timeParser `json:"time"` } `json:"txpk"`
    }{}

    err := json.Unmarshal(raw, payload)
    err = json.Unmarshal(raw, timeStruct)

    if err != nil {
        return err, nil
    }

    if timeStruct.Stat != nil {
        payload.Stat.Time = timeStruct.Stat.Time.Value
    }

    if timeStruct.RXPK != nil {
        for i, x := range(*timeStruct.RXPK) {
            (*payload.RXPK)[i].Time = x.Time.Value
        }
    }

    if timeStruct.TXPK != nil {
        payload.TXPK.Time = timeStruct.TXPK.Time.Value
    }

    return nil, payload
}
