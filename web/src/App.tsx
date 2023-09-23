import React from 'react';
import './App.css';
import { MQTTConnectionProvider } from './useMqtt';
import { Gamepad } from './Gamepad';
import Video from './Video';

function App() {
  return (
    <div className="App">
      <MQTTConnectionProvider>
        <Video width="1280px" height="720px" />
        <div className="gamepads">
          {navigator.getGamepads().map((_, index) => (
            <Gamepad key={index} index={index} />
          ))}
        </div>
      </MQTTConnectionProvider>
    </div>
  );
}

export default App;
