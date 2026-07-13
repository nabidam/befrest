export const MSG_HELLO = 'hello';
export const MSG_SET_NAME = 'set-name';
export const MSG_WELCOME = 'welcome';
export const MSG_NEED_NAME = 'need-name';
export const MSG_DEVICES = 'devices';
export const MSG_ERROR = 'error';

export type DeviceKind = 'mobile' | 'desktop';

export interface Device {
  id: string;
  name: string;
  rawName: string;
  kind: DeviceKind;
  isHost: boolean;
  connectedAt: string;
}

export interface HelloMessage {
  type: typeof MSG_HELLO;
  deviceId?: string;
  name?: string;
}

export interface SetNameMessage {
  type: typeof MSG_SET_NAME;
  name: string;
}

export interface WelcomeMessage {
  type: typeof MSG_WELCOME;
  deviceId: string;
  self: Device;
  isHost: boolean;
}

export interface NeedNameMessage {
  type: typeof MSG_NEED_NAME;
  suggested: string;
}

export interface DevicesMessage {
  type: typeof MSG_DEVICES;
  devices: Device[];
}

export interface ErrorMessage {
  type: typeof MSG_ERROR;
  code: 'target-gone' | 'bad-request';
  message: string;
}

export type ServerMessage = WelcomeMessage | NeedNameMessage | DevicesMessage | ErrorMessage;
