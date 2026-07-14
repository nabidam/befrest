import { beforeEach, describe, expect, it, vi } from 'vitest';

const { offerFiles } = vi.hoisted(() => ({ offerFiles: vi.fn() }));

vi.mock('./ws', () => ({ offerFiles }));

import { offerSelectedFiles, registerOfferFiles } from './upload';
import { connection } from './stores';
import type { Transfer } from './proto';

describe('offerSelectedFiles', () => {
  beforeEach(() => {
    offerFiles.mockReset();
    connection.set('connecting');
  });

  it('uses the same offer path as the file picker and retains files for the created transfer', () => {
    const file = new File(['photo'], 'photo.jpg', { type: 'image/jpeg' });
    offerFiles.mockReturnValue(true);

    expect(offerSelectedFiles('receiver-1', [file])).toBe(true);
    expect(offerFiles).toHaveBeenCalledWith('receiver-1', [file]);

    const transfer = {
      id: 'transfer-1',
      receiverId: 'receiver-1',
    } as Transfer;
    registerOfferFiles(transfer);
  });

  it('queues files selected while the mobile file picker has disconnected the socket', () => {
    offerFiles.mockReturnValue(false);

    expect(offerSelectedFiles('receiver-2', [new File(['x'], 'x.txt')])).toBe(true);

    offerFiles.mockReturnValue(true);
    connection.set('connected');

    expect(offerFiles).toHaveBeenCalledTimes(2);
    expect(offerFiles).toHaveBeenLastCalledWith('receiver-2', [expect.any(File)]);
  });
});
