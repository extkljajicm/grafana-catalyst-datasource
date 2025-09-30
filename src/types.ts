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
  queryType?: string;
  limit?: number;
  priority?: string[];
  status?: string[];
  device?: string;
  mac?: string;
  site?: string;
  rule?: string;
  enrich?: boolean;
  siteType?: string;
  parentSiteName?: string;
  siteName?: string;
  siteId?: string;
  parentSiteId?: string;
  metric?: string[];
  endpoint?: string;
  deviceId?: string;
  macAddress?: string;
  issueStatus?: CatalystIssueStatus;
  aiDriven?: string;
}

/**
 * Defines the default values for a new query in the query editor.
 */
export const DEFAULT_QUERY: Partial<CatalystQuery> = {
  limit: 100,
  priority: ['P1', 'P2'],
  status: ['ACTIVE'],
  enrich: true,
  siteType: 'area',
  parentSiteName: '',
  siteName: '',
  siteId: '',
  parentSiteId: '',
  metric: [],
  endpoint: 'issues',
  deviceId: '',
  macAddress: '',
  issueStatus: 'ACTIVE',
  aiDriven: '',
};

export const metricOptions = [
  { label: 'Client Count', value: 'clientCount' },
  { label: 'Health Score', value: 'healthScore' },
  { label: 'Access Point Count', value: 'accessPointCount' },
  { label: 'Switch Count', value: 'switchCount' },
  { label: 'Router Count', value: 'routerCount' },
];

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
