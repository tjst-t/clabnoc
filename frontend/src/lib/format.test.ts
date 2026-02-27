import { describe, it, expect } from 'vitest';
import { formatBytes } from './format';

describe('formatBytes', () => {
  it('returns "0 B" for zero bytes', () => {
    expect(formatBytes(0)).toBe('0 B');
  });

  it('returns "-- B" for negative values', () => {
    expect(formatBytes(-1)).toBe('-- B');
    expect(formatBytes(-1024)).toBe('-- B');
  });

  it('returns "-- B" for NaN', () => {
    expect(formatBytes(NaN)).toBe('-- B');
  });

  it('returns "-- B" for Infinity', () => {
    expect(formatBytes(Infinity)).toBe('-- B');
    expect(formatBytes(-Infinity)).toBe('-- B');
  });

  it('formats small byte values in B', () => {
    expect(formatBytes(1)).toBe('1.0 B');
    expect(formatBytes(512)).toBe('512.0 B');
    expect(formatBytes(1023)).toBe('1023.0 B');
  });

  it('formats KiB range values', () => {
    expect(formatBytes(1024)).toBe('1.0 KiB');
    expect(formatBytes(1536)).toBe('1.5 KiB');
    expect(formatBytes(1024 * 500)).toBe('500.0 KiB');
  });

  it('formats MiB range values', () => {
    expect(formatBytes(1024 ** 2)).toBe('1.0 MiB');
    expect(formatBytes(100 * 1024 ** 2)).toBe('100.0 MiB');
    expect(formatBytes(104857600)).toBe('100.0 MiB');
  });

  it('formats GiB range values', () => {
    expect(formatBytes(1024 ** 3)).toBe('1.0 GiB');
    expect(formatBytes(1073741824)).toBe('1.0 GiB');
    expect(formatBytes(2.5 * 1024 ** 3)).toBe('2.5 GiB');
  });

  it('formats TiB range values', () => {
    expect(formatBytes(1024 ** 4)).toBe('1.0 TiB');
    expect(formatBytes(3 * 1024 ** 4)).toBe('3.0 TiB');
  });

  it('formats PiB range values', () => {
    expect(formatBytes(1024 ** 5)).toBe('1.0 PiB');
  });

  it('clamps extremely large values to PiB unit', () => {
    // 1024 PiB = 1 EiB, but should stay expressed in PiB
    expect(formatBytes(1024 ** 6)).toBe('1024.0 PiB');
    expect(formatBytes(1024 ** 7)).toBe('1048576.0 PiB');
  });
});
