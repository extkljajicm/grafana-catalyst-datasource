// cache.ts: Simple in-memory cache with TTL support
// Provides caching for API responses to reduce redundant requests

export interface CacheEntry<T> {
  data: T;
  timestamp: number;
  ttl: number; // Time to live in milliseconds
}

export class SimpleCache<T> {
  private cache = new Map<string, CacheEntry<T>>();

  /**
   * Get a value from the cache
   * Returns undefined if the key doesn't exist or has expired
   */
  get(key: string): T | undefined {
    const entry = this.cache.get(key);
    if (!entry) {
      return undefined;
    }

    // Check if entry has expired
    const now = Date.now();
    if (now - entry.timestamp > entry.ttl) {
      this.cache.delete(key);
      return undefined;
    }

    return entry.data;
  }

  /**
   * Set a value in the cache with a TTL
   * @param key Cache key
   * @param data Data to cache
   * @param ttl Time to live in milliseconds (default: 5 minutes)
   */
  set(key: string, data: T, ttl: number = 5 * 60 * 1000): void {
    this.cache.set(key, {
      data,
      timestamp: Date.now(),
      ttl,
    });
  }

  /**
   * Check if a key exists and is not expired
   */
  has(key: string): boolean {
    return this.get(key) !== undefined;
  }

  /**
   * Remove a specific key from the cache
   */
  delete(key: string): void {
    this.cache.delete(key);
  }

  /**
   * Clear all entries from the cache
   */
  clear(): void {
    this.cache.clear();
  }

  /**
   * Remove all expired entries from the cache
   */
  prune(): void {
    const now = Date.now();
    for (const [key, entry] of this.cache.entries()) {
      if (now - entry.timestamp > entry.ttl) {
        this.cache.delete(key);
      }
    }
  }

  /**
   * Get the number of entries in the cache (including expired)
   */
  size(): number {
    return this.cache.size;
  }
}

/**
 * Create a cache key from various parameters
 */
export function createCacheKey(prefix: string, params: Record<string, any>): string {
  const sortedKeys = Object.keys(params).sort();
  const keyParts = sortedKeys.map((k) => `${k}=${JSON.stringify(params[k])}`);
  return `${prefix}:${keyParts.join('&')}`;
}
