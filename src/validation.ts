// validation.ts: Configuration and query validation utilities
// Provides validation functions to ensure proper configuration before API calls

import { ConfigurationError } from './errors';
import type { CatalystJsonData, CatalystQuery } from './types';

/**
 * Validates the datasource configuration
 */
export function validateConfig(config: CatalystJsonData): { valid: boolean; errors: string[] } {
  const errors: string[] = [];

  // Base URL validation
  if (!config.baseUrl || config.baseUrl.trim() === '') {
    errors.push('Base URL is required');
  } else {
    try {
      const url = new URL(config.baseUrl);
      if (!url.protocol.startsWith('http')) {
        errors.push('Base URL must use HTTP or HTTPS protocol');
      }
    } catch {
      errors.push('Base URL is not a valid URL');
    }
  }

  return {
    valid: errors.length === 0,
    errors,
  };
}

/**
 * Validates a query before execution
 */
export function validateQuery(query: CatalystQuery): { valid: boolean; errors: string[] } {
  const errors: string[] = [];

  // Check query type
  if (!query.queryType) {
    errors.push('Query type is required');
  }

  // Validate limit if provided
  if (query.limit !== undefined && query.limit !== null) {
    if (query.limit < 1) {
      errors.push('Limit must be greater than 0');
    }
    if (query.limit > 10000) {
      errors.push('Limit must be less than or equal to 10000');
    }
  }

  // Validate priority values
  if (query.priority && query.priority.length > 0) {
    const validPriorities = ['P1', 'P2', 'P3', 'P4'];
    const invalid = query.priority.filter((p) => !validPriorities.includes(p));
    if (invalid.length > 0) {
      errors.push(`Invalid priority values: ${invalid.join(', ')}`);
    }
  }

  // Validate status values
  if (query.status && query.status.length > 0) {
    const validStatuses = ['ACTIVE', 'RESOLVED', 'IGNORED'];
    const invalid = query.status.filter((s) => !validStatuses.includes(s));
    if (invalid.length > 0) {
      errors.push(`Invalid status values: ${invalid.join(', ')}`);
    }
  }

  return {
    valid: errors.length === 0,
    errors,
  };
}

/**
 * Throws ConfigurationError if validation fails
 */
export function assertValidConfig(config: CatalystJsonData): void {
  const result = validateConfig(config);
  if (!result.valid) {
    throw new ConfigurationError(
      `Invalid configuration: ${result.errors.join(', ')}`,
      'Please fix the configuration errors in the datasource settings.'
    );
  }
}

/**
 * Throws ConfigurationError if query validation fails
 */
export function assertValidQuery(query: CatalystQuery): void {
  const result = validateQuery(query);
  if (!result.valid) {
    throw new ConfigurationError(
      `Invalid query: ${result.errors.join(', ')}`,
      'Please correct the query parameters and try again.'
    );
  }
}
