package edgo

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"path/filepath"
	"strings"
)

type Json map[string]interface{}

func ParseJournalLine(contents []byte) (Json, error) {
	jsonMap := make(Json)
	err := json.Unmarshal(contents, &jsonMap)
	return jsonMap, err
}

func IsStatusFile(filename string) bool {
	base := strings.ToLower(filepath.Base(filename))
	switch base {
	case "cargo.json":
		return true
	case "market.json":
		return true
	case "modulesinfo.json":
		return true
	case "navroute.json":
		return true
	case "outfitting.json":
		return true
	case "shipyard.json":
		return true
	case "status.json":
		return true
	case "shiplocker.json":
		return true
	default:
		return false
	}
}

func GetStatusInterface(filename string) interface{} {
	base := strings.ToLower(filepath.Base(filename))
	switch base {
	case "cargo.json":
		return &Cargo{}
	case "market.json":
		return &Market{}
	case "modulesinfo.json":
		return &ModulesInfo{}
	case "navroute.json":
		return &NavRoute{}
	case "outfitting.json":
		return &Outfitting{}
	case "shipyard.json":
		return &Shipyard{}
	case "status.json":
		return &Status{}
	case "shiplocker.json":
		return make(map[string]interface{}) // TODO
	default:
		return make(map[string]interface{})
	}
}

func ParseStatusContents(filename string, content []byte) (interface{}, error) {
	if !IsStatusFile(filename) {
		return nil, errors.New("Not a status file")
	}
	obj := GetStatusInterface(filename)
	err := json.Unmarshal(content, obj)
	return obj, err
}

func ParseStatusData(filename string) (interface{}, error) {
	if !IsStatusFile(filename) {
		return nil, errors.New("Not a status file")
	}
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	obj := GetStatusInterface(filename)
	err = json.Unmarshal(content, obj)
	return obj, err
}

func GetEventNameByte(contents []byte) string {
	// The quick and dirty way is to find the event field and extract
	// the value from it.
	idx := bytes.Index(contents, []byte(`"event"`))
	if idx != -1 {
		tmp := contents[idx+7 : len(contents)-1]
		f := bytes.SplitN(tmp, []byte(`"`), 3)
		return string(f[1])
	}
	return ""
}

func GetEventName(i interface{}) string {
	switch v := i.(type) {
	case []byte:
		return GetEventNameByte(v)
	case string:
		return GetEventNameByte([]byte(v))
	case Json:
		return v["event"].(string)
	case *Cargo:
		return v.Event
	case *Market:
		return v.Event
	case *ModulesInfo:
		return v.Event
	case *NavRoute:
		return v.Event
	case *Outfitting:
		return v.Event
	case *Shipyard:
		return v.Event
	case *Status:
		return v.Event
	case *Base:
		return v.Event
	default:
		return ""
	}
}

func GetEventTimestampByte(contents []byte) string {
	// The quick and dirty way is to find the event field and extract
	// the value from it.
	idx := bytes.Index(contents, []byte(`"timestamp"`))
	if idx != -1 {
		tmp := contents[idx+11 : len(contents)-1]
		f := bytes.SplitN(tmp, []byte(`"`), 3)
		return string(f[1])
	}
	return ""
}

func GetEventTimestamp(i interface{}) string {
	switch v := i.(type) {
	case []byte:
		return GetEventTimestampByte(v)
	case string:
		return GetEventTimestampByte([]byte(v))
	case Json:
		return v["timestamp"].(string)
	case *Cargo:
		return v.Timestamp
	case *Market:
		return v.Timestamp
	case *ModulesInfo:
		return v.Timestamp
	case *NavRoute:
		return v.Timestamp
	case *Outfitting:
		return v.Timestamp
	case *Shipyard:
		return v.Timestamp
	case *Status:
		return v.Timestamp
	case *Base:
		return v.Timestamp
	default:
		return ""
	}
}
