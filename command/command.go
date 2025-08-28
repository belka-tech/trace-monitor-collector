package command

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/mailru/easyjson"
)

var (
	ErrorEmptyJson = errors.New("FromJson: empty json")
)

type Command struct {
	Pid        string          `json:"pid"`
	Method     string          `json:"method"`
	TraceId    string          `json:"traceId"`
	Data       json.RawMessage `json:"data"`
	SentAt     time.Time       `json:"sentAt"`
	RawCommand []byte          `json:"-"`
}

func FromJson(js []byte) (Command, error) {
	if len(js) == 0 {
		return Command{}, ErrorEmptyJson
	}

	cmd := Command{}
	err := easyjson.Unmarshal(js, &cmd)
	if err != nil {
		return cmd, err
	}
	cmd.RawCommand = js

	return cmd, nil
}
