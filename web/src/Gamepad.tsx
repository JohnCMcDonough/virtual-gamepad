import React, { useEffect, useState } from 'react';
import { useConnection } from './useMqtt';
import { GamepadState, useGamepadState } from './useGamepad';


function gamepadStateToBuffer(gamepadState: GamepadState): Buffer {
  const buffer = Buffer.alloc(2 + (4 * 6));
  buffer[0] |= +gamepadState.ButtonNorth << 0;
  buffer[0] |= +gamepadState.ButtonSouth << 1;
  buffer[0] |= +gamepadState.ButtonWest << 2;
  buffer[0] |= +gamepadState.ButtonEast << 3;

  buffer[0] |= +gamepadState.ButtonBumperLeft << 4;
  buffer[0] |= +gamepadState.ButtonBumperRight << 5;

  buffer[0] |= +gamepadState.ButtonThumbLeft << 6;
  buffer[0] |= +gamepadState.ButtonThumbRight << 7;

  buffer[1] |= +gamepadState.ButtonSelect << 0;
  buffer[1] |= +gamepadState.ButtonStart << 1;

  buffer[1] |= +gamepadState.ButtonDpadUp << 2;
  buffer[1] |= +gamepadState.ButtonDpadDown << 3;
  buffer[1] |= +gamepadState.ButtonDpadLeft << 4;
  buffer[1] |= +gamepadState.ButtonDpadRight << 5;

  buffer[1] |= +gamepadState.ButtonMode << 6;

  buffer.writeFloatLE(gamepadState.AxisLeftX, 2);
  buffer.writeFloatLE(gamepadState.AxisLeftY, 6);

  buffer.writeFloatLE(gamepadState.AxisRightX, 10);
  buffer.writeFloatLE(gamepadState.AxisRightY, 14);

  buffer.writeFloatLE(gamepadState.AxisLeftTrigger, 18);
  buffer.writeFloatLE(gamepadState.AxisRightTrigger, 22);

  return buffer;
}

export const Gamepad: React.FC<{ index: number }> = ({ index }) => {
  const connection = useConnection();
  const gamepad = useGamepadState(index);
  const [previousMessage, setPreviousMessage] = useState<null | Buffer>(null);

  useEffect(() => {
    if (!gamepad) {
      return;
    }

    const newMessage = gamepadStateToBuffer(gamepad);
    if (previousMessage?.toString('base64') !== newMessage.toString('base64')) {
      console.log('publishing', newMessage.toString('base64'));
      if (connection && connection.connected) {
        connection.publish(`/gamepad/${index}/state`, newMessage);
      }
      setPreviousMessage(newMessage);
    }
  }, [connection, index, gamepad, previousMessage])

  return (
    <section>
      <h2>Gamepad {index}</h2>
      <pre style={{ textAlign: 'left' }}>
        {JSON.stringify(gamepad, null, 2)}
      </pre>
    </section>
  );
}