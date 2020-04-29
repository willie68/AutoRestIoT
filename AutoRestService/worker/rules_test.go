package worker

import (
	"encoding/json"
	"reflect"
	"testing"
)

type TestRuleStruct struct {
	RuleName string
	RuleSrc  string
	JsonSrc  string
	JsonExp  string
}

var TestingRules = []TestRuleStruct{
	TestRuleStruct{
		RuleName: "tasmotaTemp",
		RuleSrc: `[
				{"operation": "shift",
					"spec": {
						"Temperature": "DS18B20.Temperature",
						"TempUnit": "TempUnit",
						"Time": "Time"
					}
				}
			]`,
		JsonSrc: `{"Time":"2020-04-27T08:47:07","DS18B20":{"Id":"0114556E95AA","Temperature":26.9},"TempUnit":"C"}`,
		JsonExp: `{"Time":"2020-04-27T08:47:07", "Temperature":26.9,"TempUnit":"C"}`,
	},
	TestRuleStruct{
		RuleName: "hmBWMMotion",
		RuleSrc: `[
				{"operation": "shift",
					"spec": {
						"Device": "hm.channelName",
						"Motion": "hm.valueStable",
						"Time": "ts"
					}
				}
			]`,
		JsonSrc: `{"val":false,"ts":1587971766000,"lc":1587971766000,"hm":{"ccu":"localhost","iface":"BidCos-RF","device":"LTK0028082","deviceName":"BWM Gartenhütte","deviceType":"HM-Sen-MDIR-O-2","channel":"LTK0028082:1","channelName":"BWM Gartenhütte","channelType":"MOTION_DETECTOR","channelIndex":1,"datapoint":"MOTION","datapointName":"BidCos-RF.LTK0028082:1.MOTION","datapointType":"BOOL","datapointMin":false,"datapointMax":true,"datapointDefault":false,"valueStable":false,"rooms":["Garten"],"room":"Garten","functions":["Homekit"],"function":"Homekit","ts":1587971766000,"lc":1587971766000,"change":false,"cache":true,"working":false,"uncertain":false,"stable":true}}`,
		JsonExp: `{"Device":"BWM Gartenhütte", "Motion":false,"Time":1587971766000}`,
	},
	TestRuleStruct{
		RuleName: "hmBWMBright",
		RuleSrc: `[
				{"operation": "shift",
					"spec": {
						"Device": "hm.channelName",
						"Brightness": "hm.valueStable",
						"Time": "ts"
					}
				}
			]`,
		JsonSrc: `{"val":181,"ts":1587973274220,"lc":1587973274220,"hm":{"ccu":"localhost","iface":"BidCos-RF","device":"LTK0028082","deviceName":"BWM Gartenhütte","deviceType":"HM-Sen-MDIR-O-2","channel":"LTK0028082:1","channelName":"BWM Gartenhütte","channelType":"MOTION_DETECTOR","channelIndex":1,"datapoint":"BRIGHTNESS","datapointName":"BidCos-RF.LTK0028082:1.BRIGHTNESS","datapointType":"INTEGER","datapointMin":0,"datapointMax":255,"datapointDefault":0,"valuePrevious":180,"valueStable":181,"rooms":["Garten"],"room":"Garten","functions":["Homekit"],"function":"Homekit","ts":1587973274220,"tsPrevious":1587972873227,"lc":1587973274220,"change":true,"cache":false,"uncertain":false,"stable":true}}`,
		JsonExp: `{"Device":"BWM Gartenhütte", "Brightness":181,"Time":1587973274220}`,
	},
	TestRuleStruct{
		RuleName: "hmTermTemp",
		RuleSrc: `[
				{"operation": "shift",
					"spec": {
						"Device": "hm.deviceName",
						"Temperature": "hm.valueStable",
						"Time": "ts"
					}
				}
			]`,
		JsonSrc: `{"val":22.6,"ts":1587973561640,"lc":1587973561640,"hm":{"ccu":"localhost","iface":"BidCos-RF","device":"OEQ1670535","deviceName":"Thermostat Bad","deviceType":"HM-TC-IT-WM-W-EU","channel":"OEQ1670535:1","channelName":"HM-TC-IT-WM-W-EU OEQ1670535:1","channelType":"WEATHER_TRANSMIT","channelIndex":1,"datapoint":"TEMPERATURE","datapointName":"BidCos-RF.OEQ1670535:1.TEMPERATURE","datapointType":"FLOAT","datapointMin":-10,"datapointMax":50,"datapointDefault":0,"datapointControl":"NONE","valuePrevious":22.7,"valueStable":22.6,"rooms":["Bad"],"room":"Bad","functions":["Heizung"],"function":"Heizung","ts":1587973561640,"tsPrevious":1587973426887,"lc":1587973561640,"change":true,"cache":false,"uncertain":false,"stable":true}}`,
		JsonExp: `{"Device":"Thermostat Bad", "Temperature":22.6,"Time":1587973561640}`,
	},
	TestRuleStruct{
		RuleName: "hmTermHumi",
		RuleSrc: `[
			{"operation": "shift",
				"spec": {
					"Device": "hm.deviceName",
					"Humidity": "hm.valueStable",
					"Time": "ts"
				}
			}, {
             "operation": "timestamp",
			 "spec": {
				 "Time": {
					"inputFormat": "$unixext",
    				"outputFormat": "2006-01-02T15:04:05"
				  }
			   }
			}
		]`,
		JsonSrc: `{"val":40, "ts":1587973561647,"lc":1587973561647,"hm":{"ccu":"localhost","iface":"BidCos-RF","device":"OEQ1670535","deviceName":"Thermostat Bad","deviceType":"HM-TC-IT-WM-W-EU","channel":"OEQ1670535:1","channelName":"HM-TC-IT-WM-W-EU OEQ1670535:1","channelType":"WEATHER_TRANSMIT","channelIndex":1,"datapoint":"HUMIDITY","datapointName":"BidCos-RF.OEQ1670535:1.HUMIDITY","datapointType":"INTEGER","datapointMin":0,"datapointMax":99,"datapointDefault":0,"datapointControl":"NONE","valuePrevious":39,"valueStable":40,"rooms":["Bad"],"room":"Bad","functions":["Heizung"],"function":"Heizung","ts":1587973561647,"tsPrevious":1587973426898,"lc":1587973561647,"change":true,"cache":false,"uncertain":false,"stable":true}}`,
		JsonExp: `{"Device":"Thermostat Bad", "Humidity":40,"Time":"2020-04-27T09:46:01"}`,
	},
}

func TestTasmotaRule(t *testing.T) {
	for _, testRule := range TestingRules {

		err := Rules.Register("mcs", testRule.RuleName, testRule.RuleSrc)
		if err != nil {
			t.Errorf("can't transforn: %v", err)
			return
		}

		jsonDest, err := Rules.TransformJSON("mcs", testRule.RuleName, []byte(testRule.JsonSrc))
		if err != nil {
			t.Errorf("can't transforn: %v", err)
			return
		}

		areEqual, _ := checkJSONBytesEqual(jsonDest, []byte(testRule.JsonExp))

		if !areEqual {
			t.Error("Transformed data does not match expectation.")
			t.Log("Source:   ", testRule.JsonSrc)
			t.Log("Expected:   ", testRule.JsonExp)
			t.Log("Actual:     ", string(jsonDest))
			t.FailNow()
		}
		t.Log("transformation OK.", string(jsonDest))
	}
}

func TestSimpleRule(t *testing.T) {
	jsonConfig := `[{
  "operation": "shift",
  "spec": {
    "Temperature": "DS18B20.Temperature",
    "TempUnit": "TempUnit"
  }
}]`

	Rules.Register("mcs", "test.me", jsonConfig)

	jsonObject := `{
  "doc": {
    "uid": 12345,
    "guid": ["guid0", "guid2", "guid4"],
    "guidObjects": [{"id": "guid0"}, {"id": "guid2"}, {"id": "guid4"}]
  },
  "top-level-key": null
}`

	jsonDest, err := Rules.TransformJSON("mcs", "test.me", []byte(jsonObject))
	if err != nil {
		t.Errorf("can't transforn: %v", err)
		return
	}
	t.Log("transformation OK.", string(jsonDest))
}

func TestUnknownRule(t *testing.T) {
	jsonObject := `{
  "doc": {
    "uid": 12345,
    "guid": ["guid0", "guid2", "guid4"],
    "guidObjects": [{"id": "guid0"}, {"id": "guid2"}, {"id": "guid4"}]
  },
  "top-level-key": null
}`

	_, err := Rules.TransformJSON("mcs", "test.me2", []byte(jsonObject))
	if err != ErrRuleNotDefined {
		t.Errorf("something goes wrong: %v", err)
		return
	}
	t.Log("unknown rule throws error.")
}

func checkJSONBytesEqual(item1, item2 []byte) (bool, error) {
	var out1, out2 interface{}

	err := json.Unmarshal(item1, &out1)
	if err != nil {
		return false, nil
	}

	err = json.Unmarshal(item2, &out2)
	if err != nil {
		return false, nil
	}

	return reflect.DeepEqual(out1, out2), nil
}
