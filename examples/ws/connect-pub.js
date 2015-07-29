/* global Paho  */

 var client = new Paho.MQTT.Client("localhost", 8081, "connect-pub");
 client.onConnectionLost = onConnectionLost;
 client.onMessageArrived = onMessageArrived;
 client.connect({
   onSuccess: onConnect,
   mqttVersion : 3
 });

 function onConnect() {
   // Once a connection has been made, make a subscription and send a message.
   console.log("onConnect");
   client.send("topic0", "Pouet", 0);
 }

 function onConnectionLost(responseObject) {
   console.log("In OnConnectionLost");
   if (responseObject.errorCode !== 0) {
     console.log("onConnectionLost:" + responseObject.errorMessage);
   }
 }

 function onMessageArrived(message) {
   console.log("Message arrived", message);
 }
