// validation.test.ts: Unit tests for validation utilities
import { validateConfig, validateQuery, assertValidConfig, assertValidQuery } from './validation';
import { ConfigurationError } from './errors';
import type { CatalystJsonData, CatalystQuery } from './types';

describe('validateConfig', () => {
  it('should validate a correct config', () => {
    const config: CatalystJsonData = {
      baseUrl: 'https://catalyst.example.com',
    };
    const result = validateConfig(config);
    expect(result.valid).toBe(true);
    expect(result.errors).toHaveLength(0);
  });

  it('should reject empty baseUrl', () => {
    const config: CatalystJsonData = {
      baseUrl: '',
    };
    const result = validateConfig(config);
    expect(result.valid).toBe(false);
    expect(result.errors).toContain('Base URL is required');
  });

  it('should reject missing baseUrl', () => {
    const config: CatalystJsonData = {};
    const result = validateConfig(config);
    expect(result.valid).toBe(false);
    expect(result.errors).toContain('Base URL is required');
  });

  it('should reject invalid URL format', () => {
    const config: CatalystJsonData = {
      baseUrl: 'not-a-valid-url',
    };
    const result = validateConfig(config);
    expect(result.valid).toBe(false);
    expect(result.errors).toContain('Base URL is not a valid URL');
  });

  it('should reject non-HTTP protocol', () => {
    const config: CatalystJsonData = {
      baseUrl: 'ftp://catalyst.example.com',
    };
    const result = validateConfig(config);
    expect(result.valid).toBe(false);
    expect(result.errors).toContain('Base URL must use HTTP or HTTPS protocol');
  });

  it('should accept HTTP URL', () => {
    const config: CatalystJsonData = {
      baseUrl: 'http://catalyst.example.com',
    };
    const result = validateConfig(config);
    expect(result.valid).toBe(true);
  });

  it('should accept HTTPS URL', () => {
    const config: CatalystJsonData = {
      baseUrl: 'https://catalyst.example.com',
    };
    const result = validateConfig(config);
    expect(result.valid).toBe(true);
  });
});

describe('validateQuery', () => {
  it('should validate a correct query', () => {
    const query: CatalystQuery = {
      refId: 'A',
      queryType: 'alerts',
      limit: 100,
    };
    const result = validateQuery(query);
    expect(result.valid).toBe(true);
    expect(result.errors).toHaveLength(0);
  });

  it('should reject missing queryType', () => {
    const query = {
      refId: 'A',
    } as CatalystQuery;
    const result = validateQuery(query);
    expect(result.valid).toBe(false);
    expect(result.errors).toContain('Query type is required');
  });

  it('should reject limit less than 1', () => {
    const query: CatalystQuery = {
      refId: 'A',
      queryType: 'alerts',
      limit: 0,
    };
    const result = validateQuery(query);
    expect(result.valid).toBe(false);
    expect(result.errors).toContain('Limit must be greater than 0');
  });

  it('should reject limit greater than 10000', () => {
    const query: CatalystQuery = {
      refId: 'A',
      queryType: 'alerts',
      limit: 10001,
    };
    const result = validateQuery(query);
    expect(result.valid).toBe(false);
    expect(result.errors).toContain('Limit must be less than or equal to 10000');
  });

  it('should reject invalid priority values', () => {
    const query: CatalystQuery = {
      refId: 'A',
      queryType: 'alerts',
      priority: ['P1', 'P5' as any, 'P6' as any],
    };
    const result = validateQuery(query);
    expect(result.valid).toBe(false);
    expect(result.errors[0]).toContain('Invalid priority values');
    expect(result.errors[0]).toContain('P5');
  });

  it('should accept valid priority values', () => {
    const query: CatalystQuery = {
      refId: 'A',
      queryType: 'alerts',
      priority: ['P1', 'P2', 'P3', 'P4'],
    };
    const result = validateQuery(query);
    expect(result.valid).toBe(true);
  });

  it('should reject invalid status values', () => {
    const query: CatalystQuery = {
      refId: 'A',
      queryType: 'alerts',
      status: ['ACTIVE', 'INVALID' as any],
    };
    const result = validateQuery(query);
    expect(result.valid).toBe(false);
    expect(result.errors[0]).toContain('Invalid status values');
    expect(result.errors[0]).toContain('INVALID');
  });

  it('should accept valid status values', () => {
    const query: CatalystQuery = {
      refId: 'A',
      queryType: 'alerts',
      status: ['ACTIVE', 'RESOLVED', 'IGNORED'],
    };
    const result = validateQuery(query);
    expect(result.valid).toBe(true);
  });
});

describe('assertValidConfig', () => {
  it('should not throw for valid config', () => {
    const config: CatalystJsonData = {
      baseUrl: 'https://catalyst.example.com',
    };
    expect(() => assertValidConfig(config)).not.toThrow();
  });

  it('should throw ConfigurationError for invalid config', () => {
    const config: CatalystJsonData = {
      baseUrl: '',
    };
    expect(() => assertValidConfig(config)).toThrow(ConfigurationError);
  });

  it('should include error details in thrown error', () => {
    const config: CatalystJsonData = {
      baseUrl: '',
    };
    try {
      assertValidConfig(config);
      fail('Should have thrown');
    } catch (error) {
      expect(error).toBeInstanceOf(ConfigurationError);
      const configError = error as ConfigurationError;
      expect(configError.message).toContain('Base URL is required');
    }
  });
});

describe('assertValidQuery', () => {
  it('should not throw for valid query', () => {
    const query: CatalystQuery = {
      refId: 'A',
      queryType: 'alerts',
      limit: 100,
    };
    expect(() => assertValidQuery(query)).not.toThrow();
  });

  it('should throw ConfigurationError for invalid query', () => {
    const query: CatalystQuery = {
      refId: 'A',
      queryType: 'alerts',
      limit: 0,
    };
    expect(() => assertValidQuery(query)).toThrow(ConfigurationError);
  });

  it('should include error details in thrown error', () => {
    const query: CatalystQuery = {
      refId: 'A',
      queryType: 'alerts',
      limit: -1,
    };
    try {
      assertValidQuery(query);
      fail('Should have thrown');
    } catch (error) {
      expect(error).toBeInstanceOf(ConfigurationError);
      const configError = error as ConfigurationError;
      expect(configError.message).toContain('Limit must be greater than 0');
    }
  });
});
