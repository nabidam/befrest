import { writable } from 'svelte/store';
import type { Device } from './proto';

export type ConnectionState = 'connecting' | 'ready' | 'joining' | 'connected' | 'error';

export const connection = writable<ConnectionState>('connecting');
export const self = writable<Device | null>(null);
export const devices = writable<Device[]>([]);
export const needsName = writable(true);
export const suggestedName = writable('');
export const connectionError = writable<string | null>(null);
