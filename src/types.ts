import type { DataQuery, DataSourceJsonData } from '@grafana/data';

/**
 * Query types supported by this datasource.
 * Keep lean for MVP: just "alerts" (DNAC issues).
 */
export type QueryType = 'alerts';

/**
 * Panel query model (frontend <-> datasource).
 * All fields are optional except `queryType`.
 *
 * Notes:
 * - DNAC /dna/intent/api/v1/issues accepts:
 *   startTime(ms), endTime(ms), siteId, deviceId, macAddress,
 *   priority(P1..P4), issueStatus(ACTIVE|IGNORED|RESOLVED), aiDriven(YES|NO)
 *
 * - The legacy fields (severity/status/text) are kept for backward-compat
 *   and may be mapped internally to DNAC params where it makes sense.
 */
export interface CatalystQuery extends DataQuery {
  queryType: QueryType;

  /** DNAC filters */
  siteId?: string;
  deviceId?: string;
  macAddress?: string;
  priority?: string;      // e.g., "P1,P2" or "P3"
  issueStatus?: string;   // e.g., "ACTIVE", "IGNORED", "RESOLVED"
  aiDriven?: string;      // "YES" | "NO"

  /**
   * Maximum number of issues to return (panel-side cap, independent of paging).
   * Server pages are fetched in chunks (e.g., 100).
   */
  limit?: number;

  /** ------------------------------------------------------------------
   *  Legacy/Generic fields (kept for compatibility with earlier drafts)
   * ------------------------------------------------------------------ */
  /**
   * @deprecated prefer `priority`
   * Comma-separated list of severities. Example: "P1,P2,P3"
   */
  severity?: string;

  /**
   * @deprecated prefer `issueStatus`
   * Comma-separated list of statuses. Example: "ACTIVE,CLEARED"
   */
  status?: string;

  /**
   * @deprecated DNAC does not have a generic text filter here; keep only if you map it yourself.
   */
  text?: string;
}

/**
 * Defaults for brand-new queries created in the query editor.
 */
export const DEFAULT_QUERY: Partial<CatalystQuery> = {
  queryType: 'alerts',
  limit: 100,
};

/**
 * Normalized row shape used when building data frames.
 * (Not required by Grafana, but useful for internal typing/mapping.)
 */
export interface CatalystAlertRow {
  time: number;     // epoch ms
  id: string;
  title: string;
  severity: string; // maps from DNAC "priority" where available
  status: string;   // maps from DNAC "issueStatus"
  category?: string;
  device?: string;
  site?: string;
  rule?: string;
  details?: string;

  // DNAC-specific convenience fields
  mac?: string;
  priority?: string; // original DNAC priority value (P1..P4)
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
 * Keep minimal for now; extend as needed for dropdowns.
 */
export type CatalystVariableQuery =
  | { type: 'priorities' }
  | { type: 'issueStatuses' }
  | { type: 'sites'; search?: string }
  | { type: 'devices'; search?: string }
  | { type: 'macs'; search?: string };
