package main

// Based on https://github.com/aws-samples/lambda-go-samples
import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// Lambda function handler
func Handler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	// stdout and stderr are sent to AWS CloudWatch Logs
	fmt.Printf("Incoming: %s %s body %d bytes from SourceIP %s\n", req.HTTPMethod, req.Path, len(req.Body), req.RequestContext.Identity.SourceIP)

	switch req.HTTPMethod {
	case "POST":
		return rest_webhooks_post(req)
	default:
		return events.APIGatewayProxyResponse{
			Body:       "Unsupported method",
			StatusCode: 404,
		}, nil
	}
}

// For marshalling Loggly events
type logglyEvent struct {
	Event map[string]interface{} `json:"message_event"`
}

// Handler for an incoming webhooks POST. Make outgoing request
func rest_webhooks_post(session events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Received Webhook data is an array [].msys.xxx_event_key.event
	var sparkPostWebhookEvents []map[string]map[string]map[string]interface{}
	if err := json.Unmarshal([]byte(session.Body), &sparkPostWebhookEvents); err != nil {
		log.Fatal(err)
	}
	var logglyOutputLines string
	// Walk the contents, building a Loggly-compatible output
	for _, ev := range sparkPostWebhookEvents {
		// dereference msys.xxx_event_key, because "type" attribute is all we need to identify the event
		for _, event := range ev["msys"] {
			var se logglyEvent
			se.Event = event
			jsonb, err := json.Marshal(se)
			if err != nil {
				log.Fatalf("JSON marshaling failed : %s", err)
			}
			logglyOutputLines += string(jsonb) + "\n"
		}
	}

	var buf = bytes.NewBufferString(logglyOutputLines)
	var logglyUrl = strings.Trim(os.Getenv("LOGGLY_URL"), " ") // Trim leading and trailing spaces, if present
	client := &http.Client{}

	// Selectively copy across request method and headers
	req, _ := http.NewRequest(session.HTTPMethod, logglyUrl, buf)
	for hname, hval := range session.Headers {
		switch hname {
		case "Accept-Encoding", "Accept", "Authorization", "Content-Type":
			req.Header.Add(hname, hval)
		case "User-Agent":
			req.Header.Add(hname, "SparkPost-Loggly adapter")
		}
	}
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	resBuffer := new(bytes.Buffer)
	resBuffer.ReadFrom(res.Body)
	resStr := resBuffer.String()
	fmt.Printf("Outgoing: %s %s: Response %d %s", req.Method, logglyUrl, res.StatusCode, resStr)

	return events.APIGatewayProxyResponse{
		Body:       resStr,
		StatusCode: res.StatusCode,
	}, nil
}

func main() {
	if runtime.GOOS == "darwin" {
		// Simple code to simulate a request locally on OSX.  Takes a local JSON file as cmd-line arg
		if len(os.Args) < 2 {
			fmt.Println("Missing JSON filename in command line args")
		} else {
			requestFileName := os.Args[1]
			b, err := ioutil.ReadFile(requestFileName) // just pass the file name
			if err != nil {
				fmt.Println(err)
			}
			var req events.APIGatewayProxyRequest
			err = json.Unmarshal(b, &req)

			_, _ = Handler(req)
		}
	} else {
		// runtime code
		lambda.Start(Handler)
	}
}
