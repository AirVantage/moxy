Moxy
====

A general purpose reverse proxy for MQTT


Compilation
-----------

You need [Go 1.4](https://golang.org)

You need to install GB:
```
go get github.com/constabulary/gb/...
```

run

```
gb build all
```

And voilà..

Testing
-------

* Start mosquitto container :
~~~
docker run -d -p 1884:1883 --user=mosquitto --name=mosquitto airvantage/mosquitto
~~~
* Build and start moxy with dumy authentication plugin
~~~
gb build all
./bin/moxy -auth=bin/moxy-dummyauth -t -v
~~~
* Run python tests (in another terminal)
~~~
python3 ./test/interoperability/client_test.py --hostname localhost --port 1883 -z -d
~~~


