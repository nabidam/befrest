import { writable } from 'svelte/store';
import type { Device, InviteInfoMessage, Transfer } from './proto';

export type ConnectionState = 'connecting' | 'ready' | 'joining' | 'connected' | 'error';

export const connection = writable<ConnectionState>('connecting');
export const self = writable<Device | null>(null);
export const devices = writable<Device[]>([]);
export const needsName = writable(true);
export const suggestedName = writable('');
export const connectionError = writable<string | null>(null);
export const invite = writable<InviteInfoMessage | null>(null);
export const devicesLoaded = writable(false);

export interface IncomingOffer {
  transfer: Transfer;
  from: Device;
}

export interface TransferStatus {
  transfer: Transfer;
  direction: 'sending' | 'receiving';
  index: number;
  sent: number;
  size: number;
  totalSent: number;
  totalSize: number;
}

export interface Toast {
  id: number;
  message: string;
  tone: 'success' | 'info';
}

export const offers = writable<IncomingOffer[]>([]);
export const transfers = writable<Record<string, TransferStatus>>({});
export const toasts = writable<Toast[]>([]);
