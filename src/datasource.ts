import { getBackendSrv, isFetchError } from '@grafana/runtime';
import {
  CoreApp,
  DataQueryRequest,
  DataQueryResponse,
  DataSourceApi,
  DataSourceInstanceSettings,
  FieldType,
  MutableDataFrame,
} from '@grafana/data';

import {
  CatalystQuery,
  CatalystJsonData,
  CatalystSecureJsonData,
  DEFAULT_QUERY,
  CatalystAlertRow,
} from './types';

type InstanceSettings = DataSourceInstanceSettings<CatalystJsonData, CatalystSecureJsonData>;

const PAGE_SIZE = 100; // DNAC issues: fetch 100 per page as requested
const MAX_PAGES = 50;  // safety cap (adjust if needed)

export class DataSource extends DataSourceApi<CatalystQuery, CatalystJsonData> {
  constructor(private instanceSettings: InstanceSettings) {
    super(instanceSettings);
  }

  getDefaultQuery(_: CoreApp): Partial<CatalystQuery> {
    return DEFAULT_QUERY;
  }

  filterQuery(query: CatalystQuery): boolean {
    // Only execute when queryType=alerts (issues) is selected
    return !!query && query.queryType === 'alerts';
  }

  /**
   * Query DNAC issues and return a single table per target.
   * - Endpoint: /dna/intent/api/v1/issues
   * - Params: startTime, endTime (ms epoch), siteId, deviceId, macAddress, priority, issueStatus
   * - Pagination: offset + limit (100 per page)
   */
  async query(options: DataQueryRequest<CatalystQuery>): Promise<DataQueryResponse> {
    const { range } = options;
    const startMs = range!.from.valueOf();
    const endMs = range!.to.valueOf();

    const frames = await Promise.all(
      options.targets
        .filter((t) => !t.hide && t.queryType === 'alerts')
        .map(async (t) => {
          const allRows: CatalystAlertRow[] = [];

          let offset = 0;
          let page = 0;

          // Build static portion of params (time window + filters)
          const baseParams = new URLSearchParams();
          baseParams.set('startTime', String(startMs));
          baseParams.set('endTime', String(endMs));

          // Filters (only set if provided)
          if (t.siteId?.trim()) baseParams.set('siteId', t.siteId.trim());
          if (t.deviceId?.trim()) baseParams.set('deviceId', t.deviceId.trim());
          if (t.macAddress?.trim()) baseParams.set('macAddress', t.macAddress.trim());
          if (t.priority?.trim()) baseParams.set('priority', t.priority.trim()); // P1,P2,P3,P4 CSV acceptable if DNAC permits, otherwise single
          if (t.issueStatus?.trim()) baseParams.set('issueStatus', t.issueStatus.trim()); // ACTIVE, IGNORED, RESOLVED

          // Optional per-target cap (still fetch in pages of PAGE_SIZE)
          const hardLimit = t.limit && t.limit > 0 ? Math.min(t.limit, PAGE_SIZE * MAX_PAGES) : undefined;

          // Pagination loop
          for (;;) {
            if (page >= MAX_PAGES) {
              break; // safety stop
            }
            const params = new URLSearchParams(baseParams.toString());
            params.set('limit', String(PAGE_SIZE));
            params.set('offset', String(offset));

            const url = this._proxy(`/dnac/dna/intent/api/v1/issues?${params.toString()}`);
            const resp = await getBackendSrv().datasourceRequest<any>({ method: 'GET', url });

            // DNAC commonly returns { response: [ ...issues ], ... } or sometimes an array
            const pageList: any[] = Array.isArray(resp.data) ? resp.data : (resp.data?.response ?? []);
            if (!pageList.length) {
              break;
            }

            // Map to rows
            const pageRows: CatalystAlertRow[] = pageList.map((it: any) => {
              // DNAC issue fields can vary by version; we normalize common names
              const ts =
                (it.timestamp as number) ??
                (it.lastOccurredTime as number) ??
                (it.startTime as number) ??
                (it.firstOccurredTime as number);

              return {
                time: ts ?? startMs,
                id: (it.id ?? it.issueId ?? it.instanceId ?? it.alertId ?? '') as string,
                title: (it.name ?? it.title ?? it.issueTitle ?? '') as string,
                severity: (it.severity ?? it.priority ?? '') as string, // priority P1-4 often used; also expose as Severity column
                status: (it.status ?? it.issueStatus ?? '') as string,  // ACTIVE/IGNORED/RESOLVED
                category: (it.category ?? it.type ?? '') as string,
                device: (it.deviceId ?? it.deviceIp ?? it.device ?? '') as string,
                site: (it.siteId ?? it.site ?? '') as string,
                rule: (it.ruleId ?? '') as string,
                details: (it.description ?? it.details ?? it.issueDescription ?? '') as string,
                mac: (it.macAddress ?? it.clientMac ?? '') as string,
                priority: (it.priority ?? '') as string,
              };
            });

            allRows.push(...pageRows);

            // Stop if this was the last page
            if (pageList.length < PAGE_SIZE) {
              break;
            }

            // Stop if we reached a hard limit
            if (hardLimit && allRows.length >= hardLimit) {
              // Trim to exactly hardLimit if we overshot
              allRows.length = hardLimit;
              break;
            }

            offset += PAGE_SIZE;
            page += 1;
          }

          const frame = new MutableDataFrame({
            refId: t.refId,
            name: 'Catalyst Issues',
            fields: [
              { name: 'Time', type: FieldType.time, values: allRows.map((r) => r.time) },
              { name: 'ID', type: FieldType.string, values: allRows.map((r) => r.id) },
              { name: 'Title', type: FieldType.string, values: allRows.map((r) => r.title) },
              { name: 'Severity', type: FieldType.string, values: allRows.map((r) => r.severity) },
              { name: 'Priority', type: FieldType.string, values: allRows.map((r) => r.priority ?? '') },
              { name: 'Status', type: FieldType.string, values: allRows.map((r) => r.status) },
              { name: 'Category', type: FieldType.string, values: allRows.map((r) => r.category ?? '') },
              { name: 'Device', type: FieldType.string, values: allRows.map((r) => r.device ?? '') },
              { name: 'MAC', type: FieldType.string, values: allRows.map((r) => r.mac ?? '') },
              { name: 'Site', type: FieldType.string, values: allRows.map((r) => r.site ?? '') },
              { name: 'Rule', type: FieldType.string, values: allRows.map((r) => r.rule ?? '') },
              { name: 'Details', type: FieldType.string, values: allRows.map((r) => r.details ?? '') },
            ],
          });

          return frame;
        })
    );

    return { data: frames };
  }

  /**
   * Test datasource – minimal request to verify connectivity and auth/token.
   */
  async testDatasource() {
    const defaultErrorMessage = 'Cannot connect to Catalyst Center (issues)';

    try {
      // Tiny window + limit=1 to minimize load
      const now = Date.now();
      const params = new URLSearchParams({
        startTime: String(now - 5 * 60 * 1000),
        endTime: String(now),
        limit: '1',
        offset: '0',
      });

      const url = this._proxy(`/dnac/dna/intent/api/v1/issues?${params.toString()}`);
      const resp = await getBackendSrv().datasourceRequest({ method: 'GET', url });

      if (resp.status >= 200 && resp.status < 300) {
        return { status: 'success', message: 'Successfully connected to Catalyst Center (issues)' };
      }

      return {
        status: 'error',
        message: resp.statusText || defaultErrorMessage,
      };
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
   * Build Grafana proxy URL using plugin.json routes (create-plugin-README guidance).
   * Route base is "dnac" -> "${jsonData.baseUrl}" with X-Auth-Token header.
   */
  private _proxy(path: string) {
    const dsId = this.instanceSettings.id;
    // path should start with '/dnac/...'
    return `/api/datasources/proxy/${dsId}${path}`;
  }
}
