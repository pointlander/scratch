package mqtt

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net"
	"testing"
	"time"

	"github.com/TIBCOSoftware/flogo-lib/core/action"
	"github.com/TIBCOSoftware/flogo-lib/core/trigger"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

var jsonMetadata = getJsonMetadata()

func getJsonMetadata() string {
	jsonMetadataBytes, err := ioutil.ReadFile("trigger.json")
	if err != nil {
		panic("No Json Metadata found for trigger.json path")
	}
	return string(jsonMetadataBytes)
}

const testConfig string = `{
  "name": "tibco-mqtt",
  "settings": {
    "broker": "tcp://127.0.0.1:1883",
    "id": "flogoEngine",
    "user": "",
    "password": "",
    "store": "",
    "qos": "0",
    "cleansess": "false"
  },
  "handlers": [
    {
      "actionId": "device_info",
      "settings": {
        "topic": "test_start"
      }
    }
  ]
}`

var _ action.Runner = &TestRunner{}

type TestRunner struct {
	t *testing.T
}

// Run implements action.Runner.Run
func (tr *TestRunner) Run(context context.Context, action action.Action, uri string,
	options interface{}) (code int, data interface{}, err error) {
	tr.t.Logf("Ran Action: %v", uri)
	return 0, nil, nil
}

func (tr *TestRunner) RunAction(context context.Context, actionID string, inputGenerator action.InputGenerator,
	options map[string]interface{}) (results map[string]interface{}, err error) {
	tr.t.Logf("Ran Action: %v", actionID)
	return nil, nil
}

func TestInit(t *testing.T) {

	// New  factory
	md := trigger.NewMetadata(jsonMetadata)
	f := NewFactory(md)

	// New Trigger
	config := trigger.Config{}
	json.Unmarshal([]byte(testConfig), config)
	tgr := f.New(&config)

	runner := &TestRunner{t: t}

	tgr.Init(runner)
}

func TestEndpoint(t *testing.T) {

	_, err := net.Dial("tcp", "127.0.0.1:1883")
	if err != nil {
		t.Log("MQTT message broker is not available, skipping test...")
		return
	}

	// New  factory
	md := trigger.NewMetadata(jsonMetadata)
	f := NewFactory(md)

	// New Trigger
	config := trigger.Config{}
	json.Unmarshal([]byte(testConfig), &config)
	tgr := f.New(&config)

	runner := &TestRunner{t: t}

	tgr.Init(runner)

	tgr.Start()
	defer tgr.Stop()

	opts := MQTT.NewClientOptions()
	opts.AddBroker("tcp://127.0.0.1:1883")
	opts.SetClientID("flogo_test")
	opts.SetUsername("")
	opts.SetPassword("")
	opts.SetCleanSession(false)

	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	t.Log("---- doing first publish ----")

	token := client.Publish("test_start", 0, false, `{"message": "Test message payload!"}`)
	token.Wait()

	duration2 := time.Duration(2) * time.Second
	time.Sleep(duration2)

	t.Log("---- doing second publish ----")

	token = client.Publish("test_start", 0, false, `{"message": "Test message payload!"}`)
	token.Wait()

	duration5 := time.Duration(5) * time.Second
	time.Sleep(duration5)

	client.Disconnect(250)
	t.Log("Sample Publisher Disconnected")
}
