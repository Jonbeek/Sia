package record

import (
    "common"
    "encoding/json"
)

type recordencoded struct {
    Type string
    Payload string
}

func Encode(r common.Record) string {
    e := recordencoded{r.Type(), r.MarshalString()}
    s, err := json.Marshal(e)
    if err != nil {
        panic(err.Error())
    }
    return string(s)
}
