import { useEffect, useState } from 'react';

export type GamepadState = {
  // Button 1
  ButtonNorth: boolean;
  ButtonSouth: boolean;
  ButtonWest: boolean;
  ButtonEast: boolean;

  ButtonBumperLeft: boolean;
  ButtonBumperRight: boolean;
  ButtonThumbLeft: boolean;
  ButtonThumbRight: boolean;

  ButtonSelect: boolean;
  ButtonStart: boolean;

  ButtonDpadUp: boolean;
  ButtonDpadDown: boolean;
  ButtonDpadLeft: boolean;
  ButtonDpadRight: boolean;

  ButtonMode: boolean;

  // Axis 1
  AxisLeftX: number;
  AxisLeftY: number;
  // Axis 2
  AxisRightX: number;
  AxisRightY: number;

  AxisLeftTrigger: number;
  AxisRightTrigger: number;
};

export const useGamepadState = (index: number) => {
  const [gamepadState, setGamepadState] = useState<GamepadState | null>(null);
  useEffect(() => {
    let handle: number;
    const cb = () => {
      const gamepad = navigator.getGamepads()[index];
      setGamepadState(
        gamepad
          ? ({
              ButtonNorth: gamepad.buttons[3].pressed,
              ButtonSouth: gamepad.buttons[0].pressed,
              ButtonWest: gamepad.buttons[2].pressed,
              ButtonEast: gamepad.buttons[1].pressed,

              ButtonBumperLeft: gamepad.buttons[4].pressed,
              ButtonBumperRight: gamepad.buttons[5].pressed,
              
              ButtonThumbLeft: gamepad.buttons[10].pressed,
              ButtonThumbRight: gamepad.buttons[11].pressed,

              ButtonSelect: gamepad.buttons[8].pressed,
              ButtonStart: gamepad.buttons[9].pressed,
              
              ButtonDpadUp: gamepad.buttons[12].pressed,
              ButtonDpadDown: gamepad.buttons[13].pressed,
              ButtonDpadLeft: gamepad.buttons[14].pressed,
              ButtonDpadRight: gamepad.buttons[15].pressed,

              ButtonMode: gamepad.buttons[16].pressed,

              AxisLeftX: gamepad.axes[0],
              AxisLeftY: gamepad.axes[1],
              AxisRightX: gamepad.axes[2],
              AxisRightY: gamepad.axes[3],

              AxisLeftTrigger: gamepad.buttons[6].value,
              AxisRightTrigger: gamepad.buttons[7].value,

              // buttons: [...gamepad.buttons.map((b) => b.pressed)],
            } as GamepadState)
          : null
      );
      handle = requestAnimationFrame(cb);
    };
    handle = requestAnimationFrame(cb);
    return () => cancelAnimationFrame(handle);
  }, [index]);

  return gamepadState;
};
