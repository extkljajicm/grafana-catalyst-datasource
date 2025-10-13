// errors.test.ts: Unit tests for error handling
import {
  CatalystError,
  AuthenticationError,
  ConnectionError,
  ConfigurationError,
  QueryError,
  APIError,
  parseError,
  formatErrorMessage,
} from './errors';

describe('CatalystError', () => {
  it('should create a base error with all properties', () => {
    const error = new CatalystError('Test error', 'TEST_CODE', 'User message', 'Remediation');
    expect(error.message).toBe('Test error');
    expect(error.code).toBe('TEST_CODE');
    expect(error.userMessage).toBe('User message');
    expect(error.remediation).toBe('Remediation');
    expect(error.name).toBe('CatalystError');
  });
});

describe('AuthenticationError', () => {
  it('should create auth error with default remediation', () => {
    const error = new AuthenticationError('Auth failed');
    expect(error.code).toBe('AUTH_ERROR');
    expect(error.userMessage).toBe('Authentication failed. Please check your credentials.');
    expect(error.remediation).toContain('username');
  });

  it('should create auth error with custom remediation', () => {
    const error = new AuthenticationError('Auth failed', 'Custom fix');
    expect(error.remediation).toBe('Custom fix');
  });
});

describe('ConnectionError', () => {
  it('should create connection error with default remediation', () => {
    const error = new ConnectionError('Connection failed');
    expect(error.code).toBe('CONNECTION_ERROR');
    expect(error.userMessage).toBe('Unable to connect to Catalyst Center.');
  });
});

describe('ConfigurationError', () => {
  it('should create config error', () => {
    const error = new ConfigurationError('Config invalid');
    expect(error.code).toBe('CONFIG_ERROR');
  });
});

describe('QueryError', () => {
  it('should create query error', () => {
    const error = new QueryError('Query failed');
    expect(error.code).toBe('QUERY_ERROR');
  });
});

describe('APIError', () => {
  it('should create API error with status code', () => {
    const error = new APIError('Request failed', 500);
    expect(error.code).toBe('API_ERROR');
    expect(error.statusCode).toBe(500);
    expect(error.userMessage).toContain('500');
  });

  it('should create API error without status code', () => {
    const error = new APIError('Request failed');
    expect(error.statusCode).toBeUndefined();
  });
});

describe('parseError', () => {
  it('should parse network errors as ConnectionError', () => {
    const error = parseError({ message: 'fetch failed' });
    expect(error).toBeInstanceOf(ConnectionError);
  });

  it('should parse 401 as AuthenticationError', () => {
    const error = parseError({ status: 401, message: 'Unauthorized' });
    expect(error).toBeInstanceOf(AuthenticationError);
  });

  it('should parse 403 as AuthenticationError', () => {
    const error = parseError({ statusCode: 403, message: 'Forbidden' });
    expect(error).toBeInstanceOf(AuthenticationError);
  });

  it('should parse 404 as APIError', () => {
    const error = parseError({ status: 404, message: 'Not found' });
    expect(error).toBeInstanceOf(APIError);
    expect((error as APIError).statusCode).toBe(404);
  });

  it('should parse 429 as APIError with rate limit message', () => {
    const error = parseError({ status: 429 });
    expect(error).toBeInstanceOf(APIError);
    expect(error.remediation).toContain('many requests');
  });

  it('should parse 500+ as APIError', () => {
    const error = parseError({ status: 500, message: 'Server error' });
    expect(error).toBeInstanceOf(APIError);
    expect((error as APIError).statusCode).toBe(500);
  });

  it('should parse config errors as ConfigurationError', () => {
    const error = parseError({ message: 'Invalid config setting' });
    expect(error).toBeInstanceOf(ConfigurationError);
  });

  it('should default to QueryError for unknown errors', () => {
    const error = parseError({ message: 'Unknown error' });
    expect(error).toBeInstanceOf(QueryError);
  });

  it('should handle errors without message', () => {
    const error = parseError({});
    expect(error).toBeInstanceOf(QueryError);
    expect(error.message).toContain('unexpected');
  });
});

describe('formatErrorMessage', () => {
  it('should format error with userMessage and remediation', () => {
    const error = new AuthenticationError('Auth failed');
    const formatted = formatErrorMessage(error);
    expect(formatted).toContain('Authentication failed');
    expect(formatted).toContain('How to fix:');
    expect(formatted).toContain('username');
  });

  it('should format error with only userMessage', () => {
    const error = new CatalystError('Test', 'CODE', 'User message');
    const formatted = formatErrorMessage(error);
    expect(formatted).toBe('User message');
  });

  it('should use message if userMessage is missing', () => {
    const error = new CatalystError('Original message', 'CODE');
    const formatted = formatErrorMessage(error);
    expect(formatted).toBe('Original message');
  });
});
