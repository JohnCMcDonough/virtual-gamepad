import 'webrtc-adapter';
import React, { useEffect, useRef, useState } from 'react';
import { useConnection, useEmitter } from './useMqtt';
import { OnMessageCallback } from 'mqtt/*';

const stream = new MediaStream();
const pc = new RTCPeerConnection({
  iceServers: [{
    urls: ["stun:stun.l.google.com:19302"]
  }]
});

pc.addTransceiver('video', {
  direction: 'recvonly',
})
pc.addTransceiver('audio', {
  direction: 'recvonly',
})

const Video: React.FunctionComponent<{
  width: string,
  height?: string,
}> = ({ width, height }) => {
  const videoRef = useRef<HTMLVideoElement>(null);
  const [connecting, setConnecting] = useState<boolean>(true);
  const [disconnected, setDisconnected] = useState<boolean>(false);
  const mqttConnection = useConnection();
  const mqttEmitter = useEmitter();
  const [sessionId] = useState<string>(Math.random().toString(36).substring(7));

  useEffect(() => {
    if (!videoRef.current || !mqttConnection || !mqttEmitter) {
      return () => { };
    }

    pc.ontrack = (event) => {
      console.log('Got track', event.track);
      stream.addTrack(event.track);
      videoRef.current!.srcObject = stream;
    }

    pc.onnegotiationneeded = async () => {
      let offer = await pc.createOffer();
      await pc.setLocalDescription(offer);

      await mqttConnection.publishAsync(`/webrtc/${sessionId}/offer`, offer.sdp!, {
        retain: true
      })
    }

    pc.onconnectionstatechange = (e) => {
      console.log('onconnectionstatechange', e, pc.connectionState);
      if (pc.connectionState === 'connected') {
        setConnecting(false);
      }
      else if (pc.connectionState === 'disconnected') {
        setDisconnected(true);
      }
    }

    const topic = `/webrtc/${sessionId}/answer`;
    mqttConnection.subscribe(topic);

    const handler: OnMessageCallback = (_, payload) => {
      pc.setRemoteDescription(new RTCSessionDescription({
        type: 'answer',
        sdp: payload.toString(),
      }))
    }
    mqttEmitter.on(topic, handler);

    return () => {
      mqttEmitter.off(topic, handler);
      mqttConnection.unsubscribe(topic);
    }
  }, [videoRef, setDisconnected, mqttConnection, mqttEmitter, sessionId, connecting]);

  if (!mqttConnection || connecting) {
    return <span>Connecting...</span>
  }

  if (disconnected) {
    return <span>WebRTC has disconnected. Refresh to try again.</span>
  }

  return (
    <video ref={videoRef} width={width} height={height} autoPlay controls={false} playsInline />
  )
}

export default Video;
