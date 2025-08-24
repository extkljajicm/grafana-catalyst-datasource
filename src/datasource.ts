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

export class DataSource extends DataSourceApi<CatalystQuery, CatalystJsonData> {
  constructor(private instanceSettings: InstanceSettings) {
    super(instanceSettings);
  }

  getDefaultQuery(_: CoreApp): Partial<CatalystQuery> {
    return DEFAULT_QUERY;
  }

  filterQuery(query: CatalystQuery): boolean {
    return query?.queryType === 'alerts';
  }

  async query(options: DataQueryRequest<CatalystQuery>): Promise<DataQueryResponse> {
    const { range } = options;
    const startMs = range!.from.valueOf();
    const endMs = range!.to.valueOf();

    const frames = await Promise.all(
      options.targets
        .filter((t) => !t.hide && t.queryType === 'alerts')
        .map(async (t) => {
          const targetLimit = clampNumber(t.limit ?? 100, 1, 10000);
          const pageSize = clampNumber(t.pageSize ?? Math.min(200, targetLimit), 1, 2000);

          // build initial params from the target
          const params = new URLSearchParams();

          // DNAC commonly accepts epoch ms; adjust here if your cluster expects seconds
          params.set('startTime', String(startMs));
          params.set('endTime', String(endMs));

          // optional filters: harmless if DNAC ignores them
          if (t.severity?.trim()) params.set('severity', t.severity.trim()); // e.g. "P1,P2,P3"
          if (t.status?.trim()) params.set('status', t.status.trim());       // e.g. "ACTIVE,CLEARED"
          if (t.text?.trim()) params.set('text', t.text.trim());             // free text if supported

          // pagination defaults
          // try to be API-agnostic: some DNACs use pageSize, some limit; we set both.
          params.set('pageSize', String(pageSize));
          if (!params.has('limit')) params.set('limit', String(pageSize));
          if (!params.has('offset')) params.set('offset', '0');

          const collected: CatalystAlertRow[] = [];
          let safetyHops = 0;

          // pagination loop
          while (collected.length < targetLimit && safetyHops < 200) {
            safetyHops++;

            const url = this._proxy(`/dnac/alerts?${params.toString()}`);
            const resp = await getBackendSrv().datasourceRequest<any>({ method: 'GET', url });

            const { items, totalCount } = extractItems(resp.data);
            if (items.length === 0) {
              break;
            }

            // map items -> CatalystAlertRow
            for (const a of items) {
              collected.push({
                time: (a?.timestamp ?? a?.startTime ?? startMs) as number,
                id: (a?.id ?? a?.instanceId ?? a?.alertId ?? '') as string,
                title: (a?.name ?? a?.title ?? '') as string,
                severity: (a?.severity ?? a?.priority ?? '') as string,
                status: (a?.status ?? a?.state ?? '') as string,
                category: a?.category ?? '',
                device: a?.deviceIp ?? a?.device ?? '',
                site: a?.siteId ?? a?.site ?? '',
                rule: a?.ruleId ?? '',
                details: a?.description ?? a?.details ?? '',
              });

              if (collected.length >= targetLimit) {
                break;
              }
            }

            if (collected.length >= targetLimit) {
              break;
            }

            // figure out "next page" using common DNAC patterns
            const progressed = advancePagination(params, resp.data, pageSize, totalCount);
            if (!progressed) {
              break; // no more pages
            }
          }

          // trim to targetLimit (in case we overshot in the last page)
          const rows = collected.slice(0, targetLimit);

          const frame = new MutableDataFrame({
            refId: t.refId,
            name: 'Catalyst Alerts',
            fields: [
              { name: 'Time', type: FieldType.time, values: rows.map((r) => r.time) },
              { name: 'ID', type: FieldType.string, values: rows.map((r) => r.id) },
              { name: 'Title', type: FieldType.string, values: rows.map((r) => r.title) },
              { name: 'Severity', type: FieldType.string, values: rows.map((r) => r.severity) },
              { name: 'Status', type: FieldType.string, values: rows.map((r) => r.status) },
              { name: 'Category', type: FieldType.string, values: rows.map((r) => r.category ?? '') },
              { name: 'Device', type: FieldType.string, values: rows.map((r) => r.device ?? '') },
              { name: 'Site', type: FieldType.string, values: rows.map((r) => r.site ?? '') },
              { name: 'Rule', type: FieldType.string, values: rows.map((r) => r.rule ?? '') },
              { name: 'Details', type: FieldType.string, values: rows.map((r) => r.details ?? '') },
            ],
          });

          return frame;
        })
    );

    return { data: frames };
  }

  async testDatasource() {
    const defaultErrorMessage = 'Cannot connect to Catalyst Center (alerts)';

    try {
      const url = this._proxy('/dnac/alerts?limit=1&pageSize=1');
      const resp = await getBackendSrv().datasourceRequest({ method: 'GET', url });

      if (resp.status >= 200 && resp.status < 300) {
        return { status: 'success', message: 'Successfully connected to Catalyst Center (alerts)' };
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

  private _proxy(path: string) {
    const dsId = this.instanceSettings.id;
    return `/api/datasources/proxy/${dsId}${path}`;
  }
}

/* --------------------------- helpers (route-agnostic) --------------------------- */

/** Try hard to pull items and total count from various DNAC-like response shapes */
function extractItems(data: any): { items: any[]; totalCount?: number } {
  if (!data) return { items: [] };

  // common shapes
  if (Array.isArray(data)) return { items: data };
  if (Array.isArray(data?.response)) return { items: data.response, totalCount: firstNumber(data, ['total', 'totalCount', 'count']) };
  if (Array.isArray(data?.items)) return { items: data.items, totalCount: firstNumber(data, ['total', 'totalCount', 'count']) };
  if (Array.isArray(data?.records)) return { items: data.records, totalCount: firstNumber(data, ['total', 'totalCount', 'count']) };

  // nested wrappers (e.g., data.data or data.result)
  if (Array.isArray(data?.data)) return { items: data.data, totalCount: firstNumber(data, ['total', 'totalCount', 'count']) };
  if (Array.isArray(data?.result)) return { items: data.result, totalCount: firstNumber(data, ['total', 'totalCount', 'count']) };

  // paged object with "content"
  if (Array.isArray(data?.content)) return { items: data.content, totalCount: firstNumber(data, ['totalElements', 'total', 'totalCount', 'count']) };

  // fallback: try to find the first array in top-level props
  for (const k of Object.keys(data)) {
    const v = (data as any)[k];
    if (Array.isArray(v)) return { items: v };
  }

  return { items: [] };
}

/**
 * Advance pagination params based on common patterns:
 * - cursor-based: nextCursor, response.nextCursor, pagination.cursor.next, links.next.href
 * - offset/limit-based: offset += pageSize until total reached (if totalCount provided)
 */
function advancePagination(params: URLSearchParams, data: any, pageSize: number, totalCount?: number): boolean {
  // 1) cursor patterns
  const cursorCandidates: Array<string | undefined> = [
    data?.nextCursor,
    data?.response?.nextCursor,
    data?.pagination?.nextCursor,
    data?.page?.next,
    data?.cursor?.next,
  ];

  for (const c of cursorCandidates) {
    if (typeof c === 'string' && c.length > 0) {
      // ensure "cursor" param is used commonly; also clear offset-based controls
      params.set('cursor', c);
      params.delete('offset');
      return true;
    }
  }

  // links.next.href (URL) pattern
  const nextHref =
    data?.links?.next?.href ??
    data?._links?.next?.href ??
    data?.pagination?.links?.next?.href ??
    data?.page?.links?.next?.href;
  if (typeof nextHref === 'string' && nextHref.length > 0) {
    // extract query part from href and replace params
    try {
      const u = new URL(nextHref, 'http://dummy'); // base is ignored if href is absolute
      params.clear();
      for (const [k, v] of u.searchParams.entries()) {
        params.append(k, v);
      }
      return true;
    } catch {
      // ignore parsing errors; fall through to offset mode
    }
  }

  // 2) offset/limit mode: if we know total, keep going until offset >= total
  const limit = firstNumberFromParams(params, ['pageSize', 'limit']) ?? pageSize;
  const currentOffset = firstNumberFromParams(params, ['offset', 'skip', 'start']) ?? 0;

  if (typeof totalCount === 'number') {
    const nextOffset = currentOffset + limit;
    if (nextOffset < totalCount) {
      // try to set the best-known offset key; we set multiple to be API-agnostic
      params.set('offset', String(nextOffset));
      params.set('skip', String(nextOffset));
      params.set('start', String(nextOffset));
      // keep size/limit hints
      params.set('pageSize', String(limit));
      params.set('limit', String(limit));
      // make sure cursor is not set
      params.delete('cursor');
      return true;
    }
    return false; // done
  }

  // if totalCount is unknown: try one more page by bumping offset (some APIs don't return total)
  const nextOffset = currentOffset + limit;
  // guardrail: if offset grew but server gave us < pageSize items, assume last page
  const inferredLastPage = inferLastPageFromItemCount(data, limit);
  if (inferredLastPage) {
    return false;
  }

  params.set('offset', String(nextOffset));
  params.set('skip', String(nextOffset));
  params.set('start', String(nextOffset));
  params.set('pageSize', String(limit));
  params.set('limit', String(limit));
  params.delete('cursor');
  return true;
}

function inferLastPageFromItemCount(data: any, limit: number): boolean {
  const { items } = extractItems(data);
  // if the server returned fewer than requested, it's likely the last page
  return items.length < limit;
}

function firstNumber(obj: any, keys: string[]): number | undefined {
  for (const k of keys) {
    const v = obj?.[k];
    if (typeof v === 'number') return v;
    if (typeof v === 'string' && v.trim() !== '' && !Number.isNaN(Number(v))) return Number(v);
  }
  return undefined;
}

function firstNumberFromParams(params: URLSearchParams, keys: string[]): number | undefined {
  for (const k of keys) {
    if (params.has(k)) {
      const v = params.get(k)!;
      const n = Number(v);
      if (!Number.isNaN(n)) return n;
    }
  }
  return undefined;
}

function clampNumber(n: number, min: number, max: number) {
  return Math.max(min, Math.min(max, n));
}
