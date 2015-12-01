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

type datrParser struct {
    Value string
}

func (d *datrParser) UnmarshalJSON (raw []byte) error {
    d.Value = strings.Trim(string(raw), `"`)

    if d.Value == "" {
        return errors.New("Invalid datr format")
    }

    return nil
}

func decodePayload (raw []byte) (error, *Payload) {
    payload := &Payload{raw, nil, nil, nil}
    customStruct := &struct{
        Stat *struct{
            Time timeParser `json:"time"`
        } `json:"stat"`
        RXPK *[]struct{
            Time timeParser `json:"time"`
            Datr datrParser `json:"datr"`
        } `json:"rxpk"`
        TXPK *struct{
            Time timeParser `json:"time"`
            Datr datrParser `json:"datr"`
        } `json:"txpk"`
    }{}

    err := json.Unmarshal(raw, payload)
    err = json.Unmarshal(raw, customStruct)

    if err != nil {
        return err, nil
    }

    if customStruct.Stat != nil {
        payload.Stat.Time = customStruct.Stat.Time.Value
    }

    if customStruct.RXPK != nil {
        for i, x := range(*customStruct.RXPK) {
            (*payload.RXPK)[i].Time = x.Time.Value
            (*payload.RXPK)[i].Datr = x.Datr.Value
        }
    }

    if customStruct.TXPK != nil {
        payload.TXPK.Time = customStruct.TXPK.Time.Value
        payload.TXPK.Datr = customStruct.TXPK.Datr.Value
    }

    return nil, payload
}
