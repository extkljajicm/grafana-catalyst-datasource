import { getBackendSrv, isFetchError, getTemplateSrv } from '@grafana/runtime';
import {
  CoreApp,
  DataQueryRequest,
  DataQueryResponse,
  DataSourceApi,
  DataSourceInstanceSettings,
  FieldType,
  MutableDataFrame,
  MetricFindValue,
} from '@grafana/data';

import {
  CatalystQuery,
  CatalystJsonData,
  CatalystAlertRow,
  CatalystVariableQuery,
  DEFAULT_QUERY,
} from './types';

type InstanceSettings = DataSourceInstanceSettings<CatalystJsonData>;

/**
 * DataSource for Cisco Catalyst Center (DNAC) alerts via issues API.
 * Proxies through Grafana using the route key "dnac" defined in plugin.json.
 */
export class DataSource extends DataSourceApi<CatalystQuery, CatalystJsonData> {
  constructor(private instanceSettings: InstanceSettings) {
    super(instanceSettings);
  }

  getDefaultQuery(_: CoreApp): Partial<CatalystQuery> {
    return DEFAULT_QUERY;
  }

  filterQuery(query: CatalystQuery): boolean {
    return !!query && query.queryType === 'alerts';
  }

  /**
   * MAIN QUERY: /dna/intent/api/v1/issues via proxy route "dnac"
   * Pagination: offset & limit (page size fixed at 100).
   * Time units: ms
   * Filters: siteId, deviceId, macAddress, priority, issueStatus
   */
  async query(options: DataQueryRequest<CatalystQuery>): Promise<DataQueryResponse> {
    const range = options.range;
    if (!range) {
      return { data: [] };
    }

    const startMs = range.from.valueOf();
    const endMs = range.to.valueOf();
    const templateSrv = getTemplateSrv();

    const frames = await Promise.all(
      options.targets
        .filter((t) => !t.hide && t.queryType === 'alerts')
        .map(async (t) => {
          const PAGE_SIZE = 100;
          const hardCap = Math.max(1, Math.min(t.limit ?? 100, 5000)); // safety cap
          let collected = 0;
          let offset = 0;

          const rows: CatalystAlertRow[] = [];

          while (collected < hardCap) {
            const params = new URLSearchParams();

            // Time window: ms
            params.set('startTime', String(startMs));
            params.set('endTime', String(endMs));

            // Filters: expand templates
            if (t.siteId) {
              const v = templateSrv.replace(t.siteId, options.scopedVars, 'csv');
              if (v.trim()) params.set('siteId', v.trim());
            }
            if (t.deviceId) {
              const v = templateSrv.replace(t.deviceId, options.scopedVars, 'csv');
              if (v.trim()) params.set('deviceId', v.trim());
            }
            if (t.macAddress) {
              const v = templateSrv.replace(t.macAddress, options.scopedVars, 'csv');
              if (v.trim()) params.set('macAddress', v.trim());
            }
            if (t.priority) {
              const v = templateSrv.replace(t.priority, options.scopedVars, 'csv');
              if (v.trim()) params.set('priority', v.trim()); // P1,P2,P3,P4
            }
            if (t.issueStatus) {
              const v = templateSrv.replace(t.issueStatus, options.scopedVars, 'csv');
              if (v.trim()) params.set('issueStatus', v.trim()); // ACTIVE,IGNORED,RESOLVED
            }

            // Pagination window
            params.set('offset', String(offset));
            params.set('limit', String(Math.min(PAGE_SIZE, hardCap - collected)));

            const url = this._proxy(`/dnac/dna/intent/api/v1/issues?${params.toString()}`);
            const resp = await getBackendSrv().datasourceRequest<any>({ method: 'GET', url });

            const arr: any[] = Array.isArray(resp.data) ? resp.data : resp.data?.response ?? [];

            for (const a of arr) {
              rows.push({
                time: (a?.timestamp ?? a?.firstOccurredTime ?? startMs) as number,
                id: (a?.issueId ?? a?.id ?? a?.instanceId ?? '') as string,
                title: (a?.name ?? a?.title ?? '') as string,
                severity: (a?.priority ?? a?.severity ?? '') as string, // DNAC calls it priority
                status: (a?.issueStatus ?? a?.status ?? '') as string,
                category: a?.category ?? '',
                device: a?.deviceId ?? a?.device ?? '',
                site: a?.siteId ?? a?.site ?? '',
                rule: a?.ruleId ?? '',
                details: a?.description ?? a?.details ?? '',
              });
            }

            collected += arr.length;
            if (arr.length < PAGE_SIZE) {
              break; // no more pages
            }
            offset += PAGE_SIZE;
          }

          const frame = new MutableDataFrame({
            refId: t.refId,
            name: 'Catalyst Issues',
            fields: [
              { name: 'Time', type: FieldType.time, values: rows.map((r) => r.time) },
              { name: 'Issue ID', type: FieldType.string, values: rows.map((r) => r.id) },
              { name: 'Title', type: FieldType.string, values: rows.map((r) => r.title) },
              { name: 'Priority', type: FieldType.string, values: rows.map((r) => r.severity) },
              { name: 'Status', type: FieldType.string, values: rows.map((r) => r.status) },
              { name: 'Category', type: FieldType.string, values: rows.map((r) => r.category ?? '') },
              { name: 'Device ID', type: FieldType.string, values: rows.map((r) => r.device ?? '') },
              { name: 'Site ID', type: FieldType.string, values: rows.map((r) => r.site ?? '') },
              { name: 'Rule', type: FieldType.string, values: rows.map((r) => r.rule ?? '') },
              { name: 'Details', type: FieldType.string, values: rows.map((r) => r.details ?? '') },
            ],
          });

          return frame;
        })
    );

    return { data: frames };
  }

  /**
   * Variable support (templating). Supports:
   * - { type: 'priorities' } -> P1,P2,P3,P4
   * - { type: 'issueStatuses' }   -> ACTIVE, IGNORED, RESOLVED
   * - { type: 'sites' }      -> distinct siteId values from recent issues
   * - { type: 'devices' }    -> distinct deviceId values from recent issues
   */
  async metricFindQuery?(query: CatalystVariableQuery, options?: any): Promise<MetricFindValue[]> {
    if (!query) {
      return [];
    }

    switch (query.type) {
      case 'priorities':
        return ['P1', 'P2', 'P3', 'P4'].map((v) => ({ text: v, value: v }));
      case 'issueStatuses':
        return ['ACTIVE', 'IGNORED', 'RESOLVED'].map((v) => ({ text: v, value: v }));
      case 'sites':
        return this._distinctFromIssues('siteId', options);
      case 'devices':
        return this._distinctFromIssues('deviceId', options);
      default:
        return [];
    }
  }

  /**
   * Simple connectivity check.
   */
  async testDatasource() {
    const defaultErrorMessage = 'Cannot connect to Catalyst Center issues API';
    try {
      const url = this._proxy('/dnac/dna/intent/api/v1/issues?limit=1&offset=0');
      const resp = await getBackendSrv().datasourceRequest<any>({ method: 'GET', url });
      if (resp.status >= 200 && resp.status < 300) {
        return { status: 'success', message: 'Successfully connected to Catalyst Center (issues)' };
      }
      return { status: 'error', message: resp.statusText || defaultErrorMessage };
    } catch (err: unknown) {
      let message = defaultErrorMessage;
      if (typeof err === 'string') {
        message = err;
      } else if (isFetchError(err)) {
        message = `Fetch error: ${err.statusText || defaultErrorMessage}`;
        const code = (err as any)?.data?.error?.code;
        const detail = (err as any)?.data?.error?.message;
        if (code) message += `: ${code}${detail ? `. ${detail}` : ''}`;
      }
      return { status: 'error', message };
    }
  }

  /**
   * Helper: build Grafana proxy URL (plugin.json route key "dnac")
   */
  private _proxy(path: string): string {
    return `/api/datasources/proxy/${this.instanceSettings.id}${path}`;
  }

  /**
   * Helper: fetch distinct values for a key from /issues over a short window.
   */
  private async _distinctFromIssues(key: 'siteId' | 'deviceId', options?: any): Promise<MetricFindValue[]> {
    const now = Date.now();
    const startMs = now - 7 * 24 * 60 * 60 * 1000; // last 7 days
    const endMs = now;
    const PAGE_SIZE = 100;

    const out = new Set<string>();
    let offset = 0;

    while (out.size < 500) {
      const params = new URLSearchParams();
      params.set('startTime', String(startMs));
      params.set('endTime', String(endMs));
      params.set('offset', String(offset));
      params.set('limit', String(PAGE_SIZE));

      const url = this._proxy(`/dnac/dna/intent/api/v1/issues?${params.toString()}`);
      const resp = await getBackendSrv().datasourceRequest<any>({ method: 'GET', url });
      const arr: any[] = Array.isArray(resp.data) ? resp.data : resp.data?.response ?? [];

      for (const a of arr) {
        const v = a?.[key];
        if (typeof v === 'string' && v.trim()) {
          out.add(v.trim());
        }
      }

      if (arr.length < PAGE_SIZE) {
        break;
      }
      offset += PAGE_SIZE;
    }

    return Array.from(out)
      .sort()
      .map((v) => ({ text: v, value: v }));
  }
}
