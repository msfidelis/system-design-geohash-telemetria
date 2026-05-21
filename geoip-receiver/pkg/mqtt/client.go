package mqtt

import (
	"time"

	mqttio "github.com/eclipse/paho.mqtt.golang"
)

func Connect(addr, clientID string) (mqttio.Client, error) {
	opts := mqttio.NewClientOptions().
		AddBroker(addr).
		SetClientID(clientID).
		SetConnectRetry(true).
		SetConnectRetryInterval(3 * time.Second)

	client := mqttio.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	return client, nil
}
