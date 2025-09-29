import type { DataQuery, DataSourceJsonData } from '@grafana/data';

/**
 * The only query type supported in this version of the plugin.
 * This corresponds to fetching issues/alerts from the Catalyst Center API.
 */
export type QueryType = 'alerts' | 'siteHealth';

// Define specific, strict types for query parameters to improve type safety.
export type CatalystPriority = 'P1' | 'P2' | 'P3' | 'P4';
export type CatalystIssueStatus = 'ACTIVE' | 'RESOLVED' | 'IGNORED';

/**
 * Represents the query structure that is sent from the frontend query editor
 * to the backend.
 *
 * NOTE:
 * - The Catalyst Center API uses `priority` (P1..P4) and `issueStatus` (ACTIVE/IGNORED/RESOLVED).
 * - The `severity` and `status` fields are included as optional aliases for backward
 *   compatibility or UI convenience. The backend is responsible for normalizing these
 *   to the correct API parameters.
 */
export interface CatalystQuery extends DataQuery {
  queryType: QueryType;

  // Common fields
  endpoint?: string;
  siteId?: string;
  limit?: number;

  // Alerts-specific fields
  deviceId?: string;
  macAddress?: string;
  priority?: CatalystPriority[];
  issueStatus?: CatalystIssueStatus;
  aiDriven?: string; // Should be 'true' or 'false' as a string.
  severity?: string; // alias for priority
  status?: string; // alias for issueStatus
  enrich?: boolean;

  // SiteHealth-specific fields
  metric?: string;
  startTime?: string;
  endTime?: string;
}

/**
 * Defines the default values for a new query in the query editor.
 */
export const DEFAULT_QUERY: Partial<CatalystQuery> = {
  queryType: 'alerts',
  limit: 25,
  enrich: false,
};

/**
 * Represents the non-sensitive configuration data for the datasource instance,
 * stored as JSON in the Grafana database.
 */
export interface CatalystJsonData extends DataSourceJsonData {
  /**
   * The base URL of the Catalyst Center API.
   * Example: https://catalyst.example.com
   */
  baseUrl?: string;
  /**
   * The selected API endpoint for queries (e.g., 'alerts', 'siteHealth').
   */
  endpoint?: string;

  /**
   * If true, the backend will not verify the TLS certificate of the API endpoint.
   * This is intended for development or lab environments with self-signed certs.
   */
  insecureSkipVerify?: boolean;
}

/**
 * Represents the sensitive configuration data for the datasource instance.
 * These values are encrypted by Grafana and stored in secureJsonData.
 */
export interface CatalystSecureJsonData {
  username?: string;
  password?: string;
  /**
   * An optional, manually provided authentication token (X-Auth-Token).
   * If this is set, the backend will use it directly and bypass the
   * username/password authentication flow.
   */
  apiToken?: string;
}

/**
 * Defines the structure of queries used in the template variable editor.
 * Each type corresponds to a function that can be called to populate a variable.
 */
export type CatalystVariableQuery =
  | { type: 'priorities' }
  | { type: 'issueStatuses' }
  | { type: 'sites'; search?: string }
  | { type: 'devices'; search?: string }
  | { type: 'macs'; search?: string };
