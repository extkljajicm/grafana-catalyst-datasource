import { DataSourceWithBackend } from '@grafana/runtime';
import type { CoreApp, DataSourceInstanceSettings, MetricFindValue } from '@grafana/data';
import type { CatalystQuery, CatalystJsonData, CatalystVariableQuery } from './types';
import { DEFAULT_QUERY as DEFAULTS } from './types';

type InstanceSettings = DataSourceInstanceSettings<CatalystJsonData>;

const PAGE_SIZE = 100;
const MAX_PAGES = 5; // variable helper only

export class DataSource extends DataSourceWithBackend<CatalystQuery, CatalystJsonData> {
  constructor(instanceSettings: InstanceSettings) {
    super(instanceSettings);
  }

  // Default query for new panels/targets
  getDefaultQuery(_: CoreApp): Partial<CatalystQuery> {
    return DEFAULTS;
  }

  // Prevent executing empty/invalid targets
  filterQuery(query: CatalystQuery): boolean {
    return !!query && query.queryType === 'alerts';
  }

  /**
   * Variables support:
   * - {type:'priorities'} -> P1..P4
   * - {type:'issueStatuses'} -> ACTIVE/IGNORED/RESOLVED
   * - {type:'sites'|'devices'|'macs'} -> dedup from /resources/issues (last 24h)
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
   * Fetch recent issues via backend resource and extract unique values for the given keys.
   * Uses last 24h and up to 5 pages (500 rows).
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

      // DataSourceWithBackend#getResource returns a Promise in Grafana 12
      const data: any = await this.getResource<any>(`issues?${params.toString()}`);
      const arr: any[] = Array.isArray(data) ? data : (data?.response ?? []);
      if (!arr.length) break;

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

      if (arr.length < PAGE_SIZE) break;
      offset += PAGE_SIZE;
    }

    return Array.from(out).sort().map((v) => ({ text: v, value: v }));
  }
}

export default DataSource;
