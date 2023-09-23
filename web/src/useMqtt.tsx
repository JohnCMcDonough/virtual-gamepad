import React, { useEffect, useState, useContext } from 'react';
import * as mqtt from 'mqtt';
import MQTTEmitter from 'mqtt-emitter';

type ContextValue = { client: mqtt.MqttClient; emitter: MQTTEmitter } | null;
const ConnectionContext = React.createContext<ContextValue>(null);
ConnectionContext.displayName = 'MQTT';

export const MQTTConnectionProvider: React.FunctionComponent<React.PropsWithChildren> = (props) => {
  const [value, setValue] = useState<ContextValue>(null);
  useEffect(() => {
    if (!value) {
      // eslint-disable-next-line no-restricted-globals
      let url = `wss://100.64.0.12:8081/`;
      // let url = `wss://localhost:8080/`;
      if (window.location.protocol.indexOf('https:') !== 0) {
        url = url.replace('wss:', 'ws:');
      }
      const client = mqtt.connect(url, {
        username: 'web',
      });
      const emitter = new MQTTEmitter();
      client.on('message', emitter.emit.bind(emitter));
      client.on('packetsend', p => console.log('mqtt packetsend', p));
      client.on('packetreceive', p => console.log('mqtt packetreceive', p));
      emitter.onadd = client.subscribe.bind(client);
      emitter.onremove = client.unsubscribe.bind(client);
      setValue({
        client,
        emitter,
      });
    }
  }, [value]);

  return (
    <ConnectionContext.Provider value={value}>
      {props.children}
    </ConnectionContext.Provider>
  );
};

export function useConnection() {
  return useContext(ConnectionContext)?.client;
}

export function useEmitter() {
  return useContext(ConnectionContext)?.emitter;
}

export function useLatestMessageFromSubscription<T extends object>(
  topic: string
) {
  const connection = useConnection();
  const emitter = useEmitter();

  const [latestMessage, setLatestMessage] = useState<T | null>(null);

  useEffect(() => {
    if (connection && emitter) {
      const handler: mqtt.OnMessageCallback = (data) => {
        setLatestMessage(JSON.parse(data.toString()));
      };
      emitter.on(topic, handler);
      return () => {
        emitter.removeListener(topic, handler);
      };
    }
  }, [connection, emitter, topic]);

  if (!connection || !emitter) {
    console.warn(
      'Subscription made to ' + topic + ' without connection being created'
    );
    return null;
  }

  return latestMessage;
}
