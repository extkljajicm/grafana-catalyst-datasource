import React, { useEffect, useRef, useState } from 'react';
import { Field, Input, Select, Switch } from '@grafana/ui';
import type { QueryEditorProps, SelectableValue } from '@grafana/data';
import { DataSource } from '../datasource';
import { CatalystQuery, CatalystJsonData, CatalystPriority, CatalystIssueStatus } from '../types';

type Props = QueryEditorProps<DataSource, CatalystQuery, CatalystJsonData>;

function useDebounced<T>(value: T, delayMs: number): T {
  const [debounced, setDebounced] = useState(value);
  useEffect(() => {
    const h = setTimeout(() => setDebounced(value), delayMs);
    return () => clearTimeout(h);
  }, [value, delayMs]);
  return debounced;
}

const PRIORITY_OPTIONS: Array<SelectableValue<CatalystPriority>> = [
  { label: 'P1', value: 'P1' },
  { label: 'P2', value: 'P2' },
  { label: 'P3', value: 'P3' },
  { label: 'P4', value: 'P4' },
];

const STATUS_OPTIONS: Array<SelectableValue<CatalystIssueStatus>> = [
  { label: 'ACTIVE', value: 'ACTIVE' },
  { label: 'RESOLVED', value: 'RESOLVED' },
  { label: 'IGNORED', value: 'IGNORED' },
];

const AI_OPTIONS: Array<SelectableValue<string>> = [
  { label: 'Any', value: '' },
  { label: 'True', value: 'true' },
  { label: 'False', value: 'false' },
];

const QueryEditor: React.FC<Props> = ({ query, onChange, onRunQuery }) => {
  // Local state mirrors query fields
  const [siteId, setSiteId] = useState<string>(query.siteId ?? '');
  const [deviceId, setDeviceId] = useState<string>(query.deviceId ?? '');
  const [macAddress, setMacAddress] = useState<string>(query.macAddress ?? '');
  const [priority, setPriority] = useState<CatalystPriority[]>(query.priority ?? []);
  const [issueStatus, setIssueStatus] = useState<CatalystIssueStatus | ''>(query.issueStatus ?? '');
  const [aiDriven, setAiDriven] = useState<string>(query.aiDriven ? 'true' : query.aiDriven === 'false' ? 'false' : '');
  const [limit, setLimit] = useState<number>(query.limit ?? 25);

  const dSiteId = useDebounced(siteId, 400);
  const dDeviceId = useDebounced(deviceId, 400);
  const dMac = useDebounced(macAddress, 400);
  const dPriority = useDebounced(priority, 400);
  const dIssueStatus = useDebounced(issueStatus, 400);
  const dAiDriven = useDebounced(aiDriven, 400);
  const dLimit = useDebounced(limit, 400);

  // Prevent loops: only push changes when the signature changes
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

  const onEnrichChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, enrich: e.currentTarget.checked });
    onRunQuery();
  };

  return (
    <div className="gf-form-group">
      <div className="gf-form">
        <Field label="Site ID">
          <Input
            value={siteId}
            onChange={(e) => setSiteId(e.currentTarget.value)}
            placeholder="e.g. 7f6c0f40-...  (supports $__value from variables)"
            width={30}
          />
        </Field>
        <Field label="Device ID">
          <Input
            value={deviceId}
            onChange={(e) => setDeviceId(e.currentTarget.value)}
            placeholder="e.g. 9b2d3a10-..."
            width={30}
          />
        </Field>
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

        <Field label="Status">
          <Select
            options={STATUS_OPTIONS}
            value={STATUS_OPTIONS.find((o) => o.value === (issueStatus || undefined)) ?? null}
            onChange={(v) => setIssueStatus(v?.value ?? '')}
            isClearable
            width={20}
          />
        </Field>

        <Field label="AI-driven">
          <Select
            options={AI_OPTIONS}
            value={AI_OPTIONS.find((o) => o.value === (aiDriven || '')) ?? AI_OPTIONS[0]}
            onChange={(v) => setAiDriven(v?.value ?? '')}
            width={12}
          />
        </Field>

        <Field label="Limit">
          <Input
            type="number"
            value={limit}
            onChange={(e) => setLimit(Number(e.currentTarget.value) || 100)}
            placeholder="Max rows"
            width={12}
          />
        </Field>

        <Field label="Fetch full details" description="Enriches issues with device/MAC details. Slower query time.">
          <Switch value={!!query.enrich} onChange={onEnrichChange} />
        </Field>
      </div>
    </div>
  );
};

export default QueryEditor;
