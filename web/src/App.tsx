import React from 'react';
import './App.css';
import { MQTTConnectionProvider } from './useMqtt';
import { Gamepad } from './Gamepad';

function App() {
  return (
    <div className="App">
      <MQTTConnectionProvider>
        {navigator.getGamepads().map((_, index) => (
          <Gamepad key={index} index={index} />
        ))}
      </MQTTConnectionProvider>
    </div>
  );
}

export default App;
