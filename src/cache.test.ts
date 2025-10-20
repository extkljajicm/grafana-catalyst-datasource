// cache.test.ts: Unit tests for caching utilities
import { SimpleCache, createCacheKey } from './cache';

describe('SimpleCache', () => {
  let cache: SimpleCache<string>;

  beforeEach(() => {
    cache = new SimpleCache<string>();
  });

  describe('get and set', () => {
    it('should store and retrieve a value', () => {
      cache.set('key1', 'value1');
      expect(cache.get('key1')).toBe('value1');
    });

    it('should return undefined for non-existent key', () => {
      expect(cache.get('nonexistent')).toBeUndefined();
    });

    it('should overwrite existing value', () => {
      cache.set('key1', 'value1');
      cache.set('key1', 'value2');
      expect(cache.get('key1')).toBe('value2');
    });
  });

  describe('TTL expiration', () => {
    beforeEach(() => {
      jest.useFakeTimers();
    });

    afterEach(() => {
      jest.useRealTimers();
    });

    it('should return value before TTL expires', () => {
      cache.set('key1', 'value1', 1000);
      jest.advanceTimersByTime(500);
      expect(cache.get('key1')).toBe('value1');
    });

    it('should return undefined after TTL expires', () => {
      cache.set('key1', 'value1', 1000);
      jest.advanceTimersByTime(1001);
      expect(cache.get('key1')).toBeUndefined();
    });

    it('should use default TTL of 5 minutes', () => {
      cache.set('key1', 'value1');
      jest.advanceTimersByTime(4 * 60 * 1000); // 4 minutes
      expect(cache.get('key1')).toBe('value1');
      jest.advanceTimersByTime(2 * 60 * 1000); // +2 minutes = 6 minutes total
      expect(cache.get('key1')).toBeUndefined();
    });

    it('should delete expired entry on access', () => {
      cache.set('key1', 'value1', 1000);
      expect(cache.size()).toBe(1);
      jest.advanceTimersByTime(1001);
      cache.get('key1');
      expect(cache.size()).toBe(0);
    });
  });

  describe('has', () => {
    it('should return true for existing non-expired key', () => {
      cache.set('key1', 'value1');
      expect(cache.has('key1')).toBe(true);
    });

    it('should return false for non-existent key', () => {
      expect(cache.has('nonexistent')).toBe(false);
    });

    it('should return false for expired key', () => {
      jest.useFakeTimers();
      cache.set('key1', 'value1', 1000);
      jest.advanceTimersByTime(1001);
      expect(cache.has('key1')).toBe(false);
      jest.useRealTimers();
    });
  });

  describe('delete', () => {
    it('should delete a specific key', () => {
      cache.set('key1', 'value1');
      cache.set('key2', 'value2');
      cache.delete('key1');
      expect(cache.get('key1')).toBeUndefined();
      expect(cache.get('key2')).toBe('value2');
    });
  });

  describe('clear', () => {
    it('should clear all entries', () => {
      cache.set('key1', 'value1');
      cache.set('key2', 'value2');
      cache.clear();
      expect(cache.get('key1')).toBeUndefined();
      expect(cache.get('key2')).toBeUndefined();
      expect(cache.size()).toBe(0);
    });
  });

  describe('prune', () => {
    beforeEach(() => {
      jest.useFakeTimers();
    });

    afterEach(() => {
      jest.useRealTimers();
    });

    it('should remove expired entries', () => {
      cache.set('key1', 'value1', 1000);
      cache.set('key2', 'value2', 2000);
      cache.set('key3', 'value3', 3000);
      
      jest.advanceTimersByTime(1500);
      cache.prune();
      
      expect(cache.get('key1')).toBeUndefined();
      expect(cache.get('key2')).toBe('value2');
      expect(cache.get('key3')).toBe('value3');
      expect(cache.size()).toBe(2);
    });

    it('should not affect non-expired entries', () => {
      cache.set('key1', 'value1', 5000);
      cache.set('key2', 'value2', 5000);
      
      jest.advanceTimersByTime(1000);
      cache.prune();
      
      expect(cache.size()).toBe(2);
    });
  });

  describe('size', () => {
    it('should return 0 for empty cache', () => {
      expect(cache.size()).toBe(0);
    });

    it('should return correct count', () => {
      cache.set('key1', 'value1');
      cache.set('key2', 'value2');
      expect(cache.size()).toBe(2);
    });

    it('should include expired entries until pruned', () => {
      jest.useFakeTimers();
      cache.set('key1', 'value1', 1000);
      jest.advanceTimersByTime(1001);
      expect(cache.size()).toBe(1); // Still counted until accessed or pruned
      cache.prune();
      expect(cache.size()).toBe(0);
      jest.useRealTimers();
    });
  });
});

describe('createCacheKey', () => {
  it('should create a key from prefix and params', () => {
    const key = createCacheKey('issues', { limit: 100, offset: 0 });
    expect(key).toContain('issues');
    expect(key).toContain('limit');
    expect(key).toContain('100');
  });

  it('should sort parameters for consistent keys', () => {
    const key1 = createCacheKey('issues', { b: 2, a: 1 });
    const key2 = createCacheKey('issues', { a: 1, b: 2 });
    expect(key1).toBe(key2);
  });

  it('should handle different parameter values', () => {
    const key1 = createCacheKey('issues', { limit: 100 });
    const key2 = createCacheKey('issues', { limit: 200 });
    expect(key1).not.toBe(key2);
  });

  it('should handle complex parameter values', () => {
    const key = createCacheKey('issues', {
      priority: ['P1', 'P2'],
      status: ['ACTIVE'],
      limit: 100,
    });
    expect(key).toContain('priority');
    expect(key).toContain('P1');
    expect(key).toContain('P2');
  });

  it('should handle empty params', () => {
    const key = createCacheKey('issues', {});
    expect(key).toBe('issues:');
  });
});
