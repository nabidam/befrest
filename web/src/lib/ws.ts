import {
  MSG_DEVICES,
  MSG_ERROR,
  MSG_HELLO,
  MSG_NEED_NAME,
  MSG_SET_NAME,
  MSG_WELCOME,
  type Device,
  type ServerMessage,
} from './proto';
import { connection, connectionError, devices, needsName, self, suggestedName } from './stores';

const DEVICE_ID_KEY = 'befrest.deviceId';
const NAME_KEY = 'befrest.name';

let socket: WebSocket | null = null;
let started = false;

function storedIdentity(): { deviceId?: string; name?: string } {
  const deviceId = localStorage.getItem(DEVICE_ID_KEY) ?? undefined;
  const name = localStorage.getItem(NAME_KEY)?.trim() || undefined;
  return { deviceId, name };
}

function socketURL(): string {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  return `${protocol}//${window.location.host}/ws`;
}

function send(message: object): boolean {
  if (socket?.readyState !== WebSocket.OPEN) {
    connection.set('error');
    connectionError.set("Can't reach the hub. Are you on the same wifi?");
    return false;
  }
  socket.send(JSON.stringify(message));
  return true;
}

function handleMessage(message: ServerMessage): void {
  switch (message.type) {
    case MSG_NEED_NAME:
      suggestedName.set(message.suggested);
      needsName.set(true);
      connection.set('ready');
      return;
    case MSG_WELCOME:
      localStorage.setItem(DEVICE_ID_KEY, message.deviceId);
      localStorage.setItem(NAME_KEY, message.self.name);
      self.set(message.self);
      needsName.set(false);
      connectionError.set(null);
      connection.set('connected');
      return;
    case MSG_DEVICES:
      devices.set(message.devices);
      return;
    case MSG_ERROR:
      connectionError.set(message.message);
      connection.set('error');
  }
}

export function connect(): void {
  if (started) return;
  started = true;

  const identity = storedIdentity();
  needsName.set(!identity.name);
  connection.set('connecting');
  socket = new WebSocket(socketURL());

  socket.addEventListener('open', () => {
    send({ type: MSG_HELLO, ...identity });
  });
  socket.addEventListener('message', (event) => {
    try {
      handleMessage(JSON.parse(String(event.data)) as ServerMessage);
    } catch {
      connectionError.set('The hub sent an invalid response.');
      connection.set('error');
    }
  });
  socket.addEventListener('error', () => {
    connectionError.set("Can't reach the hub. Are you on the same wifi?");
    connection.set('error');
  });
  socket.addEventListener('close', () => {
    if (socket) {
      connectionError.set("Can't reach the hub. Are you on the same wifi?");
      connection.set('error');
    }
  });
}

export function setName(name: string): void {
  const trimmedName = name.trim();
  if (!trimmedName || trimmedName.length > 32) return;
  connection.set('joining');
  connectionError.set(null);
  send({ type: MSG_SET_NAME, name: trimmedName });
}
