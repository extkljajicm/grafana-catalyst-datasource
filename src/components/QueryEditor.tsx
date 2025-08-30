import React, { ChangeEvent, useEffect, useState } from 'react';
import { Field, Input, Select } from '@grafana/ui';
import { QueryEditorProps, SelectableValue } from '@grafana/data';
import { DataSource } from '../datasource';
import { CatalystQuery, CatalystJsonData } from '../types';

type Props = QueryEditorProps<DataSource, CatalystQuery, CatalystJsonData>;

function useDebounced<T>(value: T, delayMs: number): T {
  const [debounced, setDebounced] = useState(value);
  useEffect(() => {
    const h = setTimeout(() => setDebounced(value), delayMs);
    return () => clearTimeout(h);
  }, [value, delayMs]);
  return debounced;
}

const aiOptions: Array<SelectableValue<string>> = [
  { label: '— Any —', value: '' },
  { label: 'YES', value: 'YES' },
  { label: 'NO', value: 'NO' },
];

export function QueryEditor({ query, onChange, onRunQuery }: Props) {
  // Local state mirrors query fields; debounced push -> onChange + onRunQuery
  const [siteId, setSiteId] = useState(query.siteId ?? '');
  const [deviceId, setDeviceId] = useState(query.deviceId ?? '');
  const [macAddress, setMacAddress] = useState(query.macAddress ?? '');
  const [priority, setPriority] = useState(query.priority ?? query.severity ?? ''); // compat
  const [issueStatus, setIssueStatus] = useState(query.issueStatus ?? query.status ?? ''); // compat
  const [aiDriven, setAiDriven] = useState(query.aiDriven ?? '');
  const [limit, setLimit] = useState<number>(query.limit ?? 100);

  const dSiteId = useDebounced(siteId, 300);
  const dDeviceId = useDebounced(deviceId, 300);
  const dMac = useDebounced(macAddress, 300);
  const dPriority = useDebounced(priority, 300);
  const dIssueStatus = useDebounced(issueStatus, 300);
  const dAiDriven = useDebounced(aiDriven, 300);
  const dLimit = useDebounced(limit, 300);

  // Push debounced values into the query model and run
  useEffect(() => {
    onChange({
      ...query,
      siteId: dSiteId || undefined,
      deviceId: dDeviceId || undefined,
      macAddress: dMac || undefined,
      priority: dPriority || undefined,
      issueStatus: dIssueStatus || undefined,
      aiDriven: dAiDriven || undefined,
      limit: dLimit,
      // keep legacy fields synced for compatibility (optional)
      severity: undefined,
      status: undefined,
    });
    onRunQuery();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [dSiteId, dDeviceId, dMac, dPriority, dIssueStatus, dAiDriven, dLimit]);

  const onText =
    (setter: React.Dispatch<React.SetStateAction<string>>) =>
    (e: ChangeEvent<HTMLInputElement>) =>
      setter(e.currentTarget.value);

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
      <Field label="Site ID">
        <Input
          value={siteId}
          onChange={onText(setSiteId)}
          placeholder="e.g. 7f6c0f40-....  (supports variables)"
          width={40}
        />
      </Field>

      <Field label="Device ID">
        <Input
          value={deviceId}
          onChange={onText(setDeviceId)}
          placeholder="e.g. 9b2d3a10-....  (supports variables)"
          width={40}
        />
      </Field>

      <Field label="MAC Address">
        <Input
          value={macAddress}
          onChange={onText(setMacAddress)}
          placeholder="aa:bb:cc:dd:ee:ff  (supports variables)"
          width={24}
        />
      </Field>

      <Field label="Priority">
        <Input
          value={priority}
          onChange={onText(setPriority)}
          placeholder="P1,P2,P3,P4  (CSV; supports variables)"
          width={24}
        />
      </Field>

      <Field label="Issue Status">
        <Input
          value={issueStatus}
          onChange={onText(setIssueStatus)}
          placeholder="ACTIVE,IGNORED,RESOLVED  (CSV; supports variables)"
          width={28}
        />
      </Field>

      <Field label="AI Driven">
        <Select
          options={aiOptions}
          value={aiOptions.find((o) => o.value === (aiDriven || '')) || aiOptions[0]}
          onChange={(v) => setAiDriven(v.value ?? '')}
          width={20}
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
    </div>
  );
}

export default QueryEditor;
