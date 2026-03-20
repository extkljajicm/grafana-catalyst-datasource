// QueryEditor: Main query configuration UI for Catalyst datasource in Grafana.
// Allows users to filter alerts/issues by site, device, MAC, priority, status, AI-driven, and more.
// Uses debounced local state to avoid excessive backend requests.
import React, { useEffect, useRef, useState } from 'react';
import { Field, Input, Select, Switch } from '@grafana/ui';
import type { QueryEditorProps, SelectableValue } from '@grafana/data';
import { DataSource } from '../datasource';
import { CatalystQuery, CatalystJsonData, CatalystPriority, CatalystIssueStatus } from '../types';

// Props: Provided by Grafana plugin system
type Props = QueryEditorProps<DataSource, CatalystQuery, CatalystJsonData>;

// useDebounced: Custom hook to debounce value changes for smoother UX and reduced backend load
function useDebounced<T>(value: T, delayMs: number): T {
  const [debounced, setDebounced] = useState(value);
  useEffect(() => {
    const h = setTimeout(() => setDebounced(value), delayMs);
    return () => clearTimeout(h);
  }, [value, delayMs]);
  return debounced;
}

// Priority dropdown options
const PRIORITY_OPTIONS: Array<SelectableValue<CatalystPriority>> = [
  { label: 'P1', value: 'P1' },
  { label: 'P2', value: 'P2' },
  { label: 'P3', value: 'P3' },
  { label: 'P4', value: 'P4' },
];

// Issue status dropdown options
const STATUS_OPTIONS: Array<SelectableValue<CatalystIssueStatus>> = [
  { label: 'ACTIVE', value: 'ACTIVE' },
  { label: 'RESOLVED', value: 'RESOLVED' },
  { label: 'IGNORED', value: 'IGNORED' },
];

// AI-driven dropdown options
const AI_OPTIONS: Array<SelectableValue<string>> = [
  { label: 'Any', value: '' },
  { label: 'True', value: 'true' },
  { label: 'False', value: 'false' },
];

// QueryEditor main component
const QueryEditor: React.FC<Props> = ({ query, onChange, onRunQuery }) => {
  // Local state mirrors query fields for controlled inputs
  const [siteId, setSiteId] = useState<string>(query.siteId ?? '');
  const [deviceId, setDeviceId] = useState<string>(query.deviceId ?? '');
  const [macAddress, setMacAddress] = useState<string>(query.macAddress ?? '');
  const [priority, setPriority] = useState<CatalystPriority[]>(query.priority ?? []);
  const [issueStatus, setIssueStatus] = useState<CatalystIssueStatus | ''>(query.issueStatus ?? '');
  const [aiDriven, setAiDriven] = useState<string>(query.aiDriven ? 'true' : query.aiDriven === 'false' ? 'false' : '');
  const [limit, setLimit] = useState<number>(query.limit ?? 25);

  // Debounced versions of each field to avoid excessive backend requests
  const dSiteId = useDebounced(siteId, 400);
  const dDeviceId = useDebounced(deviceId, 400);
  const dMac = useDebounced(macAddress, 400);
  const dPriority = useDebounced(priority, 400);
  const dIssueStatus = useDebounced(issueStatus, 400);
  const dAiDriven = useDebounced(aiDriven, 400);
  const dLimit = useDebounced(limit, 400);

  // Effect: Synchronize debounced state with parent query object
  // Prevents infinite loops by tracking last signature
  const lastSig = useRef<string>('');
  useEffect(() => {
    const sig = [
      dSiteId || '',
      dDeviceId || '',
      dMac || '',
      (dPriority || []).join(','),
      dIssueStatus || '',
      dAiDriven || '',
      String(dLimit ?? 0),
    ].join('|');

    if (sig === lastSig.current) {
      return;
    }
    lastSig.current = sig;

    // Build next query object for backend
    const next: CatalystQuery = {
      ...query,
      queryType: 'alerts',
      siteId: dSiteId || undefined,
      deviceId: dDeviceId || undefined,
      macAddress: dMac || undefined,
      priority: (dPriority && dPriority.length > 0) ? dPriority : undefined,
      issueStatus: dIssueStatus === '' ? undefined : dIssueStatus,
      aiDriven: dAiDriven === '' ? undefined : dAiDriven,
      limit: dLimit ?? 25,
      // clear deprecated aliases (backend already handles)
      severity: undefined,
      status: undefined,
    };

    onChange(next);
    onRunQuery();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [dSiteId, dDeviceId, dMac, dPriority, dIssueStatus, dAiDriven, dLimit]);

  // Handler: Toggle enrichment (fetches extra device/MAC details)
  const onEnrichChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, enrich: e.currentTarget.checked });
    onRunQuery();
  };

  // Render: Query configuration form
  return (
    <div className="gf-form-group">
      <div className="gf-form">
        {/* Site ID field. Supports template variables. */}
        <Field label="Site ID">
          <Input
            value={siteId}
            onChange={(e) => setSiteId(e.currentTarget.value)}
            placeholder="e.g. 7f6c0f40-...  (supports $__value from variables)"
            width={30}
          />
        </Field>
        {/* Device ID field */}
        <Field label="Device ID">
          <Input
            value={deviceId}
            onChange={(e) => setDeviceId(e.currentTarget.value)}
            placeholder="e.g. 9b2d3a10-..."
            width={30}
          />
        </Field>
        {/* MAC address field */}
        <Field label="MAC">
          <Input
            value={macAddress}
            onChange={(e) => setMacAddress(e.currentTarget.value)}
            placeholder="00:11:22:33:44:55"
            width={30}
          />
        </Field>
      </div>

      <div className="gf-form">
        {/* Priority multi-select */}
        <Field label="Priority">
          <Select
            options={PRIORITY_OPTIONS}
            value={PRIORITY_OPTIONS.filter(o => priority.includes(o.value!))}
            onChange={(v) => setPriority(v.map((opt: SelectableValue<CatalystPriority>) => opt.value!))}
            isClearable
            isMulti
            width={20}
          />
        </Field>

        {/* Issue status dropdown */}
        <Field label="Status">
          <Select
            options={STATUS_OPTIONS}
            value={STATUS_OPTIONS.find((o) => o.value === (issueStatus || undefined)) ?? null}
            onChange={(v) => setIssueStatus(v?.value ?? '')}
            isClearable
            width={20}
          />
        </Field>

        {/* AI-driven dropdown */}
        <Field label="AI-driven">
          <Select
            options={AI_OPTIONS}
            value={AI_OPTIONS.find((o) => o.value === (aiDriven || '')) ?? AI_OPTIONS[0]}
            onChange={(v) => setAiDriven(v?.value ?? '')}
            width={12}
          />
        </Field>

        {/* Limit field (max rows) */}
        <Field label="Limit">
          <Input
            type="number"
            value={limit}
            onChange={(e) => setLimit(Number(e.currentTarget.value) || 100)}
            placeholder="Max rows"
            width={12}
          />
        </Field>

        {/* Enrichment toggle */}
        <Field label="Fetch full details" description="Enriches issues with device/MAC details. Slower query time.">
          <Switch value={!!query.enrich} onChange={onEnrichChange} />
        </Field>
      </div>
    </div>
  );
};

// Default export for plugin registration
export default QueryEditor;
