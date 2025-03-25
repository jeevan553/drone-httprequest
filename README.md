The _httprequest_ plugin in Drone facilitates the automation of HTTP requests during CI/CD pipelines. It allows the user to send HTTP/HTTPS requests to specified endpoints and perform actions based on the response, similar to the Jenkins HTTP Request Plugin. Users can set headers, authenticate requests, define expected response codes, and configure timeouts.

# Building

Build the plugin binary:

```text
scripts/build.sh
```

Build the plugin image:

```text
docker build -t local/httprequest -f docker/Dockerfile .
```

# Testing

Execute the plugin from your current working directory:

```text
Simple GET request
docker run --rm \
-e PLUGIN_URL='https://httpbin.org/get' \
-e PLUGIN_HTTP_METHOD='GET' \
-e PLUGIN_HEADERS='Content-Type:application/json' \
-e PLUGIN_TIMEOUT='30' \
  -w /drone/src  -v $(pwd):/drone/src local/httprequest


GET request with response logged in to a file
docker run --rm \
-e PLUGIN_URL='https://httpbin.org/get' \
-e PLUGIN_HTTP_METHOD='GET' \
-e PLUGIN_HEADERS='Content-Type:application/json' \
-e PLUGIN_TIMEOUT='30' \
-e PLUGIN_OUTPUT_FILE='/tmp/output_6008.txt' \
  -w /drone/src  -v $(pwd):/drone/src local/httprequest


GET request with proxy usage
docker run --rm \
-e PLUGIN_URL='https://httpbin.org/get' \
-e PLUGIN_HTTP_METHOD='GET' \
-e PLUGIN_HEADERS='Content-Type:application/json' \
-e PLUGIN_TIMEOUT='30' \
-e PLUGIN_IGNORE_SSL='true' \
-e PLUGIN_PROXY='http://localhost:8888' \
  -w /drone/src  -v $(pwd):/drone/src local/httprequest

Simple POST request
docker run --rm \
-e PLUGIN_URL='https://httpbin.org/post' \
-e PLUGIN_HTTP_METHOD='POST' \
-e PLUGIN_HEADERS='Content-Type:application/json' \
-e PLUGIN_REQUEST_BODY='{"name":"drone"}' \
-e PLUGIN_TIMEOUT='30' \
  -w /drone/src  -v $(pwd):/drone/src local/httprequest

```

## Harness CI Example:
```yaml
              - step:
                  type: Plugin
                  name: httprequest-Test
                  identifier: Email_Plugin
                  spec:
                    connectorRef: account.harnessImage
                    image: plugins/httprequest:linux-amd64
                    settings:
                      access_token: <JFROG_ACCESS_TOKEN>
                      url: https://httpbin.org/get
                      headers: Content-Type:application/json
                      ssl_cert_path: <ssl-cert-location-path>
```

### Maven Build and Publish reference
[Go to Maven reference](./docs/MAVEN_README.md)

### Gradle Build and Publish reference
[Go to Gradle reference](./docs/GRADLE_README.md)

## Community and Support
[Harness Community Slack](https://join.slack.com/t/harnesscommunity/shared_invite/zt-y4hdqh7p-RVuEQyIl5Hcx4Ck8VCvzBw) - Join the #drone slack channel to connect with our engineers and other users running Drone CI.

[Harness Community Forum](https://community.harness.io/) - Ask questions, find answers, and help other users.

[Report and Track A Bug](https://community.harness.io/c/bugs/17) - Find a bug? Please report in our forum under Drone Bugs. Please provide screenshots and steps to reproduce. 

[Events](https://www.meetup.com/harness/) - Keep up to date with Drone events and check out previous events [here](https://www.youtube.com/watch?v=Oq34ImUGcHA&list=PLXsYHFsLmqf3zwelQDAKoVNmLeqcVsD9o).