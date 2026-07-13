import {
  MSG_ACCEPT,
  MSG_DEVICES,
  MSG_DECLINE,
  MSG_ERROR,
  MSG_FILE_READY,
  MSG_INVITE_INFO,
  MSG_HELLO,
  MSG_NEED_NAME,
  MSG_OFFER,
  MSG_OFFER_CANCEL,
  MSG_OFFER_CANCELLED,
  MSG_OFFER_CREATED,
  MSG_PROGRESS,
  MSG_SET_NAME,
  MSG_TRANSFER_ACCEPTED,
  MSG_TRANSFER_DECLINED,
  MSG_TRANSFER_DONE,
  MSG_TRANSFER_CANCEL,
  MSG_TRANSFER_FAILED,
  MSG_WELCOME,
  type OfferRequestMessage,
  type Device,
  type Transfer,
  type ServerMessage,
} from './proto';
import { beginUpload, registerOfferFiles } from './upload';
import { downloadFile } from './download';
import { formatOfferCancellation, formatTransferFailure } from './format';
import { connection, connectionError, devices, devicesLoaded, invite, needsName, offers, self, suggestedName, toasts, transfers, type Toast, type TransferStatus } from './stores';

const DEVICE_ID_KEY = 'befrest.deviceId';
const NAME_KEY = 'befrest.name';

let socket: WebSocket | null = null;
let started = false;
let nextToastID = 0;
let reconnectAttempt = 0;
let reconnectTimer: number | undefined;
let loadingTimer: number | undefined;

function storedIdentity(): { deviceId?: string; name?: string } {
  const deviceId = localStorage.getItem(DEVICE_ID_KEY) ?? undefined;
  const name = localStorage.getItem(NAME_KEY)?.trim() || undefined;
  return { deviceId, name };
}

function hostToken(): string | undefined {
  const token = new URLSearchParams(window.location.search).get('hostToken');
  return token || undefined;
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

function addToast(message: string, tone: Toast['tone'] = 'info'): void {
  const toast = { id: ++nextToastID, message, tone };
  toasts.update((items) => [...items, toast]);
  window.setTimeout(() => {
    toasts.update((items) => items.filter((item) => item.id !== toast.id));
  }, 4_000);
}

function setTransfer(transfer: Transfer, direction: TransferStatus['direction']): void {
  transfers.update((items) => ({
    ...items,
    [transfer.id]: {
      transfer,
      direction,
      index: 0,
      sent: 0,
      size: transfer.files[0]?.size ?? 0,
      totalSent: 0,
      totalSize: transfer.files.reduce((total, file) => total + file.size, 0),
    },
  }));
}

function clearTransfer(transferID: string): TransferStatus | undefined {
  let current: TransferStatus | undefined;
  transfers.update((items) => {
    current = items[transferID];
    const { [transferID]: _, ...remaining } = items;
    return remaining;
  });
  return current;
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
      devicesLoaded.set(true);
      return;
    case MSG_INVITE_INFO:
      invite.set(message);
      return;
    case MSG_ERROR:
      connectionError.set(message.message);
      connection.set('error');
      addToast(message.message);
      return;
    case MSG_OFFER_CREATED:
      registerOfferFiles(message.transfer);
      setTransfer(message.transfer, 'sending');
      return;
    case MSG_OFFER:
      offers.update((items) => [...items, { transfer: message.transfer, from: message.from }]);
      setTransfer(message.transfer, 'receiving');
      return;
    case MSG_TRANSFER_ACCEPTED:
      transfers.update((items) => {
        const transfer = items[message.transferId];
        if (!transfer) return items;
        return {
          ...items,
          [message.transferId]: {
            ...transfer,
            transfer: { ...transfer.transfer, state: 'accepted' },
          },
        };
      });
      void beginUpload(message.transferId);
      return;
    case MSG_TRANSFER_DECLINED: {
      const transfer = clearTransfer(message.transferId);
      addToast(`${transfer?.transfer.files[0]?.name ?? 'Transfer'} was declined`);
      return;
    }
    case MSG_OFFER_CANCELLED: {
      let sender = 'The sender';
      offers.update((items) => {
        const offer = items.find((item) => item.transfer.id === message.transferId);
        sender = offer?.from.name ?? sender;
        return items.filter((item) => item.transfer.id !== message.transferId);
      });
      clearTransfer(message.transferId);
      addToast(formatOfferCancellation(message.reason, sender), 'failure');
      return;
    }
    case MSG_FILE_READY:
      downloadFile(message.transferId, message.index);
      return;
    case MSG_PROGRESS:
      transfers.update((items) => {
        const transfer = items[message.transferId];
        if (!transfer) return items;
        return {
          ...items,
          [message.transferId]: {
            ...transfer,
            transfer: { ...transfer.transfer, state: 'streaming' },
            index: message.index,
            sent: message.sent,
            size: message.size,
            totalSent: message.totalSent,
            totalSize: message.totalSize,
          },
        };
      });
      return;
    case MSG_TRANSFER_DONE: {
      const transfer = clearTransfer(message.transferId);
      addToast(transfer?.direction === 'receiving' ? 'Saved to Downloads ✓' : 'Sent ✓', 'success');
      return;
    }
    case MSG_TRANSFER_FAILED: {
      const transfer = clearTransfer(message.transferId);
      const counterpartID = transfer?.direction === 'sending' ? transfer.transfer.receiverId : transfer?.transfer.senderId;
      let counterpart = 'The other device';
      if (counterpartID) {
        devices.update((items) => {
          counterpart = items.find((device) => device.id === counterpartID)?.name ?? counterpart;
          return items;
        });
      }
      addToast(formatTransferFailure(message.reason, counterpart), 'failure');
      return;
    }
  }
}

export function connect(): void {
  if (started) return;
  started = true;

  openSocket();
}

function openSocket(): void {
  if (reconnectTimer !== undefined) {
    window.clearTimeout(reconnectTimer);
    reconnectTimer = undefined;
  }

  const identity = storedIdentity();
  needsName.set(!identity.name);
  connection.set('connecting');
  devices.set([]);
  devicesLoaded.set(false);
  if (loadingTimer !== undefined) window.clearTimeout(loadingTimer);
  loadingTimer = window.setTimeout(() => devicesLoaded.set(true), 2_000);
  const nextSocket = new WebSocket(socketURL());
  socket = nextSocket;

  nextSocket.addEventListener('open', () => {
    reconnectAttempt = 0;
    send({ type: MSG_HELLO, ...identity, hostToken: hostToken() });
  });
  nextSocket.addEventListener('message', (event) => {
    try {
      handleMessage(JSON.parse(String(event.data)) as ServerMessage);
    } catch {
      connectionError.set('The hub sent an invalid response.');
      connection.set('error');
    }
  });
  nextSocket.addEventListener('close', () => {
    if (socket !== nextSocket) return;
    socket = null;
    connectionError.set("Can't reach the hub. Are you on the same wifi?");
    connection.set('connecting');
    scheduleReconnect();
  });
}

function scheduleReconnect(): void {
  if (reconnectTimer !== undefined) return;
  const delay = Math.min(500 * 2 ** reconnectAttempt, 8_000);
  reconnectAttempt += 1;
  reconnectTimer = window.setTimeout(() => {
    reconnectTimer = undefined;
    openSocket();
  }, delay);
}

export function setName(name: string): void {
  const trimmedName = name.trim();
  if (!trimmedName || trimmedName.length > 32) return;
  connection.set('joining');
  connectionError.set(null);
  send({ type: MSG_SET_NAME, name: trimmedName });
}

export function offerFiles(to: string, files: File[]): boolean {
  const message: OfferRequestMessage = {
    type: MSG_OFFER,
    to,
    files: files.map((file) => ({ name: file.name, size: file.size })),
  };
  return send(message);
}

export function respondToOffer(transferID: string, accepted: boolean): void {
  offers.update((items) => items.filter((offer) => offer.transfer.id !== transferID));
  send({ type: accepted ? MSG_ACCEPT : MSG_DECLINE, transferId: transferID });
}

export function cancelOffer(transferID: string): void {
  send({ type: MSG_OFFER_CANCEL, transferId: transferID });
}

export function cancelTransfer(transferID: string): void {
  send({ type: MSG_TRANSFER_CANCEL, transferId: transferID });
}
