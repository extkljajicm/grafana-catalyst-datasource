import type { DataQuery, DataSourceJsonData } from '@grafana/data';

/**
 * Single query type for MVP.
 */
export type QueryType = 'alerts';

/**
 * Panel query model.
 *
 * NOTE:
 * - DNAC uses `priority` (P1..P4) and `issueStatus` (ACTIVE/IGNORED/RESOLVED).
 * - Some UI code might still refer to `severity` and `status`. We keep them here
 *   as optional aliases so the editors compile. The datasource/backend will
 *   normalize to `priority` / `issueStatus`.
 */
export interface CatalystQuery extends DataQuery {
  queryType: QueryType;

  // DNAC filters
  siteId?: string;
  deviceId?: string;
  macAddress?: string;
  priority?: string;     // P1,P2,P3,P4
  issueStatus?: string;  // ACTIVE,IGNORED,RESOLVED
  aiDriven?: string;     // YES,NO

  // UI-friendly aliases (optional). Frontend can map these to DNAC fields.
  severity?: string;     // alias for priority
  status?: string;       // alias for issueStatus

  // Hard cap on results (applied after pagination merge)
  limit?: number;
}

/**
 * Defaults for new queries.
 */
export const DEFAULT_QUERY: Partial<CatalystQuery> = {
  queryType: 'alerts',
  limit: 100,
};

export interface CatalystJsonData extends DataSourceJsonData {
  /**
   * Example: https://dnac.example.com/dna/intent/api/v1
   */
  baseUrl?: string;

  /**
   * When true, the backend will not verify TLS certs.
   * For lab/self-signed environments only.
   */
  insecureSkipVerify?: boolean;
}

export interface CatalystSecureJsonData {
  username?: string;
  password?: string;
  /**
   * Optional manual token override (X-Auth-Token). If provided, backend will
   * prefer this over performing the username/password token exchange.
   */
  apiToken?: string;
}

/**
 * Variable editor query model for templating.
 */
export type CatalystVariableQuery =
  | { type: 'priorities' }
  | { type: 'issueStatuses' }
  | { type: 'sites'; search?: string }
  | { type: 'devices'; search?: string }
  | { type: 'macs'; search?: string };
