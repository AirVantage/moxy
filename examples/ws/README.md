## Websockets example

1. Run mosquitto container locally:
2. Run moxy with dummy handler:
~~~
cd moxy
gb build all
export BE_URL=localhost:8080
export MQTT_URL=localhost:1884
./bin/moxy -auth=bin/moxy-dummyauth -t -v
~~~
3. Run static server for example files:
~~~
cd moxy/examples
python -m SimpleHTTPServer 8082
~~~
4. Open "localhost:8082/connect-sub"

  This will connect to moxy and subscribe to a topics.

5. Open "localhost:8082/connect-pub"

  This will connect to moxy, and publish on the same topic.

You'll get notifications in the console if the message is received.
