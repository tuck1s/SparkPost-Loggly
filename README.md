# SparkPost-Loggly
Adapt SparkPost webhooks event stream to Loggly http bulk endpoint.

- diagram

## Installation

This project is intended for AWS Lambda / API Gateway deployment.

Build the code locally using `make`. This creates `main.zip`.

In AWS, create a new Lambda function. Specify name, Runtime = Go 1.x, choose role as `lambda_basic_excecution`.

- config API gateway
- config Loggly
- config SparkPost webhooks output

## Testing the adapter deployment

#### Input
You can directly feed input to the adapter app using `curl` or Postman, so that you can see the https response
code coming back. Here is a minimal example, Replace the URL with your own.

```
curl -X POST https://xyzzy.execute-api.us-west-2.amazonaws.com/xyzzy \
  -d '[{"msys":{"message_event":{"type":"test1"}}},{"msys":{"message_event":{"type":"test2"}}}]'
```

You should see `{"response" : "ok"}`

#### Output

You can replace the Loggly URL with any standard webhooks receiving app, such as RequestBin, and look at the raw output.
This is done by editing the Lambda function env var. 

Note that adapter output is line-oriented, i.e. each event is separated by newlines with no surrounding `[ ]`.

```
{"message_event":{"type":"test1"}}
{"message_event":{"type":"test2"}}
```

#### Input to Loggly
You can also provide test inputs to the Loggly directly using `curl` or Postman,
see [this article](https://www.loggly.com/docs/http-bulk-endpoint/).
