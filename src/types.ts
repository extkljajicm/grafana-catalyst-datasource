import type { DataQuery, DataSourceJsonData } from '@grafana/data';

/**
 * Query types supported by this datasource.
 * Keep lean for MVP: just "alerts".
 */
export type QueryType = 'alerts';

/**
 * Panel query model (frontend <-> datasource).
 * All fields are optional except `queryType`.
 */
export interface CatalystQuery extends DataQuery {
  queryType: QueryType;

  /**
   * Comma-separated list of severities.
   * Example: "P1,P2,P3"
   */
  severity?: string;

  /**
   * Comma-separated list of statuses.
   * Example: "ACTIVE,CLEARED"
   */
  status?: string;

  /**
   * Optional free-text filter.
   * For MVP we’ll pass this as a generic text filter (you can later
   * repurpose it for site/device/category once we lock your DNAC schema).
   */
  text?: string;

  /**
   * Maximum number of alerts to return (after pagination/merge).
   */
  limit?: number;
}

/**
 * Defaults for brand-new queries created in the query editor.
 */
export const DEFAULT_QUERY: Partial<CatalystQuery> = {
  queryType: 'alerts',
  limit: 100,
};

/**
 * Optional shape we normalize alert rows to when building data frames.
 * (Not required by Grafana, but useful for internal typing/mapping.)
 */
export interface CatalystAlertRow {
  time: number;         // epoch ms
  id: string;
  title: string;
  severity: string;
  status: string;
  category?: string;
  device?: string;
  site?: string;
  rule?: string;
  details?: string;
}

/**
 * Data source JSON options (saved non-securely in Grafana).
 * `baseUrl` is used by the proxy route in plugin.json.
 */
export interface CatalystJsonData extends DataSourceJsonData {
  /**
   * Example: https://dnac.example.com/dna/intent/api/v1
   * NOTE: Must match what you configured in plugin.json routes.url
   */
  baseUrl?: string;
}

/**
 * Secure options (never sent to the browser in plaintext).
 * Used to inject headers via Grafana's proxy route.
 */
export interface CatalystSecureJsonData {
  /**
   * X-Auth-Token for Catalyst Center.
   */
  apiToken?: string;
}

/**
 * (Optional) Variable/templating query model.
 * Keep minimal for now; we can extend as we add dropdowns.
 */
export type CatalystVariableQuery =
  | { type: 'severities' }
  | { type: 'statuses' }
  | { type: 'sites'; search?: string }
  | { type: 'devices'; search?: string };
