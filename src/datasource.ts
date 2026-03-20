// This file implements the main frontend logic for the Catalyst datasource plugin.
// It extends DataSourceWithBackend, which means most query logic is handled by the Go backend.
// This class is responsible for:
// - Providing default query values for new panels
// - Filtering out invalid queries before sending to the backend
// - Implementing template variable queries (metricFindQuery)
// - Efficiently extracting unique values for template variables from recent issues

import { DataSourceWithBackend } from '@grafana/runtime';
import type { CoreApp, DataSourceInstanceSettings, MetricFindValue } from '@grafana/data';
import { DEFAULT_QUERY as DEFAULTS, type CatalystQuery, type CatalystJsonData, type CatalystVariableQuery } from './types';

type InstanceSettings = DataSourceInstanceSettings<CatalystJsonData>;

// PAGE_SIZE: Number of rows to fetch per page when populating template variables.
// MAX_PAGES: Maximum number of pages to fetch (prevents excessive API calls).
const PAGE_SIZE = 100;
const MAX_PAGES = 5; // variable helper only

export class DataSource extends DataSourceWithBackend<CatalystQuery, CatalystJsonData> {
  constructor(instanceSettings: InstanceSettings) {
    super(instanceSettings);
  }

  // Returns the default query structure for new panels/targets.
  getDefaultQuery(_: CoreApp): Partial<CatalystQuery> {
    return DEFAULTS;
  }

  // Prevents Grafana from executing empty or invalid queries.
  // Only queries of type 'alerts' are allowed.
  filterQuery(query: CatalystQuery): boolean {
    return !!query && query.queryType === 'alerts';
  }

  /**
   * Implements template variable support for the plugin.
   * Supports the following variable types:
   * - priorities: Returns static list P1..P4
   * - issueStatuses: Returns static list ACTIVE/IGNORED/RESOLVED
   * - sites, devices, macs: Fetches recent issues and extracts unique values for the given keys
   */
  async metricFindQuery(raw?: CatalystVariableQuery | string): Promise<MetricFindValue[]> {
    const q = (raw || { type: 'priorities' }) as CatalystVariableQuery;

    switch (q.type) {
      case 'priorities':
        return ['P1', 'P2', 'P3', 'P4'].map((v) => ({ text: v, value: v }));

      case 'issueStatuses':
        return ['ACTIVE', 'IGNORED', 'RESOLVED'].map((v) => ({ text: v, value: v }));

      case 'sites':
        return this.uniqueFromIssues(['siteId'], q.search);

      case 'devices':
        return this.uniqueFromIssues(['deviceId', 'deviceIp', 'device'], q.search);

      case 'macs':
        return this.uniqueFromIssues(['macAddress', 'clientMac'], q.search);

      default:
        return [];
    }
  }

  /**
   * Helper function to fetch recent issues via the backend resource handler
   * and extract unique values for a given set of keys. Used for dynamic template variables.
   * - Queries for issues in the last 24 hours
   * - Paginates up to MAX_PAGES to avoid excessive API calls
   * - Applies optional search filter to returned values
   */
  private async uniqueFromIssues(keys: string[], search?: string): Promise<MetricFindValue[]> {
    const out = new Set<string>();
    const now = Date.now();
    const start = now - 24 * 60 * 60 * 1000;
    const s = (search ?? '').toLowerCase().trim();

    let offset = 0;
    for (let page = 0; page < MAX_PAGES; page++) {
      const params = new URLSearchParams({
        startTime: String(start),
        endTime: String(now),
        limit: String(PAGE_SIZE),
        offset: String(offset),
      });

      // getResource calls the backend's CallResource handler, which proxies to the Catalyst Center API.
      const data: any = await this.getResource<any>(`issues?${params.toString()}`);
      const arr: any[] = Array.isArray(data) ? data : (data?.response ?? []);
      if (!arr.length) {break;}

      for (const it of arr) {
        for (const k of keys) {
          const v = it?.[k];
          if (typeof v === 'string' && v.trim()) {
            const val = v.trim();
            if (!s || val.toLowerCase().includes(s)) {
              out.add(val);
            }
            break;
          }
        }
      }

      if (arr.length < PAGE_SIZE) {break;}
      offset += PAGE_SIZE;
    }

    return Array.from(out).sort().map((v) => ({ text: v, value: v }));
  }
}

export default DataSource;
