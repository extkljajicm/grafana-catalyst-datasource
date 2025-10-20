// logger.ts: Structured logging utility for the Catalyst datasource
// Provides consistent logging across the plugin with configurable log levels

export enum LogLevel {
  DEBUG = 0,
  INFO = 1,
  WARN = 2,
  ERROR = 3,
  NONE = 4,
}

export interface LogContext {
  [key: string]: any;
}

export class Logger {
  private level: LogLevel;
  private prefix: string;

  constructor(prefix = 'CatalystDatasource', level: LogLevel = LogLevel.INFO) {
    this.prefix = prefix;
    this.level = level;
  }

  setLevel(level: LogLevel): void {
    this.level = level;
  }

  private shouldLog(level: LogLevel): boolean {
    return level >= this.level;
  }

  private formatMessage(level: string, message: string, context?: LogContext): string {
    const timestamp = new Date().toISOString();
    let formatted = `[${timestamp}] [${this.prefix}] [${level}] ${message}`;
    
    if (context && Object.keys(context).length > 0) {
      formatted += ` ${JSON.stringify(context)}`;
    }
    
    return formatted;
  }

  debug(message: string, context?: LogContext): void {
    if (this.shouldLog(LogLevel.DEBUG)) {
      // eslint-disable-next-line no-console
      console.debug(this.formatMessage('DEBUG', message, context));
    }
  }

  info(message: string, context?: LogContext): void {
    if (this.shouldLog(LogLevel.INFO)) {
      // eslint-disable-next-line no-console
      console.info(this.formatMessage('INFO', message, context));
    }
  }

  warn(message: string, context?: LogContext): void {
    if (this.shouldLog(LogLevel.WARN)) {
      // eslint-disable-next-line no-console
      console.warn(this.formatMessage('WARN', message, context));
    }
  }

  error(message: string, error?: Error | any, context?: LogContext): void {
    if (this.shouldLog(LogLevel.ERROR)) {
      const errorContext = {
        ...context,
        ...(error && {
          error: error.message || String(error),
          stack: error.stack,
        }),
      };
      // eslint-disable-next-line no-console
      console.error(this.formatMessage('ERROR', message, errorContext));
    }
  }

  /**
   * Create a child logger with a sub-prefix
   */
  child(subPrefix: string): Logger {
    return new Logger(`${this.prefix}:${subPrefix}`, this.level);
  }
}

// Default logger instance
export const logger = new Logger();

// Configure log level from environment or localStorage if available
try {
  if (typeof window !== 'undefined' && window.localStorage) {
    const savedLevel = window.localStorage.getItem('catalyst_log_level');
    if (savedLevel) {
      const level = LogLevel[savedLevel as keyof typeof LogLevel];
      if (level !== undefined) {
        logger.setLevel(level);
      }
    }
  }
} catch (e) {
  // Ignore errors accessing localStorage
}
