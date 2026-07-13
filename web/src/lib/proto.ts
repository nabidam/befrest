export const MSG_HELLO = 'hello';
export const MSG_SET_NAME = 'set-name';
export const MSG_WELCOME = 'welcome';
export const MSG_NEED_NAME = 'need-name';
export const MSG_DEVICES = 'devices';
export const MSG_ERROR = 'error';
export const MSG_OFFER = 'offer';
export const MSG_OFFER_CREATED = 'offer-created';
export const MSG_ACCEPT = 'accept';
export const MSG_DECLINE = 'decline';
export const MSG_TRANSFER_ACCEPTED = 'transfer-accepted';
export const MSG_TRANSFER_DECLINED = 'transfer-declined';
export const MSG_FILE_READY = 'file-ready';
export const MSG_PROGRESS = 'progress';
export const MSG_TRANSFER_DONE = 'transfer-done';
export const MSG_INVITE_INFO = 'invite-info';

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
  hostToken?: string;
}

export interface InviteInfoMessage {
  type: typeof MSG_INVITE_INFO;
  urls: { mdns: string; ip: string };
  port: number;
  reachabilityHint?: string;
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

export interface FileMeta {
  index: number;
  name: string;
  size: number;
  sent: number;
}

export interface Transfer {
  id: string;
  senderId: string;
  receiverId: string;
  files: FileMeta[];
  state: 'offered' | 'accepted' | 'streaming' | 'done' | 'declined';
  createdAt: string;
}

export interface OfferMessage {
  type: typeof MSG_OFFER;
  transfer: Transfer;
  from: Device;
}

export interface OfferRequestMessage {
  type: typeof MSG_OFFER;
  to: string;
  files: Pick<FileMeta, 'name' | 'size'>[];
}

export interface TransferIDMessage {
  type:
    | typeof MSG_ACCEPT
    | typeof MSG_DECLINE
    | typeof MSG_TRANSFER_ACCEPTED
    | typeof MSG_TRANSFER_DECLINED
    | typeof MSG_TRANSFER_DONE;
  transferId: string;
}

export interface OfferCreatedMessage {
  type: typeof MSG_OFFER_CREATED;
  transfer: Transfer;
}

export interface FileReadyMessage {
  type: typeof MSG_FILE_READY;
  transferId: string;
  index: number;
}

export interface ProgressMessage {
  type: typeof MSG_PROGRESS;
  transferId: string;
  index: number;
  sent: number;
  size: number;
  totalSent: number;
  totalSize: number;
}

export interface ClientTransferIDMessage {
  type: typeof MSG_ACCEPT | typeof MSG_DECLINE;
  transferId: string;
}

export type ServerMessage =
  | WelcomeMessage
  | NeedNameMessage
  | DevicesMessage
  | ErrorMessage
  | OfferMessage
  | OfferCreatedMessage
  | TransferIDMessage
  | FileReadyMessage
  | ProgressMessage
  | InviteInfoMessage;
