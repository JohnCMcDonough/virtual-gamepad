import 'webrtc-adapter';
import React, { useEffect, useRef, useState } from 'react';
import { useConnection, useEmitter } from './useMqtt';
import { OnMessageCallback } from 'mqtt/*';

const audio = new MediaStream();
const video = new MediaStream();
const pc = new RTCPeerConnection({
  // iceServers: [{
  //   // urls: ["stun:stun.l.google.com:19302"]
  //   urls: []
  // }]
});

const Video: React.FunctionComponent<{
  width: string,
  height?: string,
  combined?: boolean,
}> = ({ width, height, combined = false }) => {
  const videoRef = useRef<HTMLVideoElement>(null);
  const audioRef = useRef<HTMLAudioElement>(null);
  const [connecting, setConnecting] = useState<boolean>(true);
  const [disconnected, setDisconnected] = useState<boolean>(false);
  const mqttConnection = useConnection();
  const mqttEmitter = useEmitter();
  const [sessionId] = useState<string>(Math.random().toString(36).substring(7));

  useEffect(() => {
    if (!videoRef.current || !mqttConnection || !mqttEmitter) {
      return () => { };
    }

    pc.addTransceiver('video', { direction: 'recvonly' });
    pc.addTransceiver('audio', { direction: 'recvonly' });

    pc.ontrack = (event) => {
      console.log('Got track', event.track);
      if (event.track.kind === 'audio' && !combined) {
        audio.addTrack(event.track);
        audioRef.current!.srcObject = audio;
      } else {
        video.addTrack(event.track);
        videoRef.current!.srcObject = video;
      }
    }

    pc.onicecandidate = async (event) => {
      console.log('onicecandidate', event);
    }

    pc.onicegatheringstatechange = async () => {
      switch (pc.iceGatheringState) {
        case "gathering":
          console.log("gathering")
          break;
        case "complete":
          await mqttConnection.publishAsync(`webrtc/${sessionId}/offer`, pc.localDescription!.sdp!);
          break;
      }
    }

    pc.onnegotiationneeded = async () => {
      console.log('onnegotiationneeded');
      let offer = await pc.createOffer();
      await pc.setLocalDescription(offer);
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

    const topic = `webrtc/${sessionId}/answer`;
    mqttConnection.subscribe(topic);

    const handler: OnMessageCallback = async (payload) => {
      console.log('Got answer', payload.toString());
      try {
        await pc.setRemoteDescription(new RTCSessionDescription({
          type: 'answer',
          sdp: payload.toString(),
        }))
        videoRef.current!.muted = false;
        audioRef.current!.muted = false;
      }
      catch (e) {
        console.error(e);
      }
    }
    mqttEmitter.on(topic, handler);

    return () => {
      mqttEmitter.removeListener(topic, handler);
      mqttConnection.unsubscribe(topic);
    }
  }, [videoRef, setDisconnected, mqttConnection, mqttEmitter, sessionId]);

  if (disconnected) {
    return <span>WebRTC has disconnected. Refresh to try again.</span>
  }

  return (
    <>
      {connecting && <span>Connecting...</span>}
      <audio ref={audioRef} autoPlay playsInline controls={true} muted></audio>
      <video ref={videoRef} width={width} height={height} autoPlay controls={true} muted playsInline style={{ display: connecting ? 'hidden' : 'block' }} />
    </>
  )
}

export default Video;
