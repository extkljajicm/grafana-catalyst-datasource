// errors.ts: Centralized error handling for the Catalyst datasource plugin
// Provides structured error types with actionable messages for users

/**
 * Base error class for all Catalyst datasource errors
 */
export class CatalystError extends Error {
  constructor(
    message: string,
    public readonly code: string,
    public readonly userMessage?: string,
    public readonly remediation?: string
  ) {
    super(message);
    this.name = 'CatalystError';
  }
}

/**
 * Authentication-related errors
 */
export class AuthenticationError extends CatalystError {
  constructor(message: string, remediation?: string) {
    super(
      message,
      'AUTH_ERROR',
      'Authentication failed. Please check your credentials.',
      remediation || 'Verify your username, password, or API token in the datasource configuration.'
    );
    this.name = 'AuthenticationError';
  }
}

/**
 * Connection-related errors
 */
export class ConnectionError extends CatalystError {
  constructor(message: string, remediation?: string) {
    super(
      message,
      'CONNECTION_ERROR',
      'Unable to connect to Catalyst Center.',
      remediation || 'Verify the Base URL is correct and that Grafana can reach your Catalyst Center instance.'
    );
    this.name = 'ConnectionError';
  }
}

/**
 * Configuration-related errors
 */
export class ConfigurationError extends CatalystError {
  constructor(message: string, remediation?: string) {
    super(
      message,
      'CONFIG_ERROR',
      'Invalid datasource configuration.',
      remediation || 'Check your datasource settings and ensure all required fields are filled correctly.'
    );
    this.name = 'ConfigurationError';
  }
}

/**
 * Query-related errors
 */
export class QueryError extends CatalystError {
  constructor(message: string, remediation?: string) {
    super(
      message,
      'QUERY_ERROR',
      'Query execution failed.',
      remediation || 'Check your query parameters and try again. View the error details for more information.'
    );
    this.name = 'QueryError';
  }
}

/**
 * API-related errors (rate limiting, timeouts, etc.)
 */
export class APIError extends CatalystError {
  constructor(
    message: string,
    public readonly statusCode?: number,
    remediation?: string
  ) {
    super(
      message,
      'API_ERROR',
      statusCode ? `API request failed with status ${statusCode}.` : 'API request failed.',
      remediation || 'The Catalyst Center API returned an error. Check the API status and try again.'
    );
    this.name = 'APIError';
  }
}

/**
 * Parse HTTP error response and convert to appropriate CatalystError
 */
export function parseError(error: any): CatalystError {
  // Handle network errors
  if (error.message?.includes('fetch') || error.message?.includes('network')) {
    return new ConnectionError(error.message);
  }

  // Handle HTTP status codes
  if (error.status || error.statusCode) {
    const status = error.status || error.statusCode;
    
    if (status === 401 || status === 403) {
      return new AuthenticationError(
        error.message || 'Unauthorized',
        'Your authentication token may have expired. Try updating your credentials or API token.'
      );
    }
    
    if (status === 404) {
      return new APIError(
        error.message || 'Resource not found',
        status,
        'The requested API endpoint was not found. Verify your Catalyst Center version and API support.'
      );
    }
    
    if (status === 429) {
      return new APIError(
        error.message || 'Rate limit exceeded',
        status,
        'Too many requests. Please wait a moment and try again.'
      );
    }
    
    if (status >= 500) {
      return new APIError(
        error.message || 'Server error',
        status,
        'The Catalyst Center server encountered an error. Please try again later.'
      );
    }
    
    return new APIError(error.message || 'Request failed', status);
  }

  // Handle configuration errors
  if (error.message?.includes('config') || error.message?.includes('setting')) {
    return new ConfigurationError(error.message);
  }

  // Default to QueryError
  return new QueryError(error.message || 'An unexpected error occurred');
}

/**
 * Format error for display to users
 */
export function formatErrorMessage(error: CatalystError): string {
  const parts = [error.userMessage || error.message];
  
  if (error.remediation) {
    parts.push('\n\nHow to fix:', error.remediation);
  }
  
  return parts.join(' ');
}
