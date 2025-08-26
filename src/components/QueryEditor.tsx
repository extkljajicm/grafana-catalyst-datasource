import React, { ChangeEvent, useEffect, useState } from 'react';
import { InlineField, Input, Select, Stack } from '@grafana/ui';
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
    <Stack gap={1}>
      <InlineField label="Site ID" labelWidth={16} tooltip="DNAC siteId (UUID). Supports variables.">
        <Input value={siteId} onChange={onText(setSiteId)} placeholder="e.g. 7f6c0f40-...." width={40} />
      </InlineField>

      <InlineField label="Device ID" labelWidth={16} tooltip="DNAC deviceId (UUID). Supports variables.">
        <Input value={deviceId} onChange={onText(setDeviceId)} placeholder="e.g. 9b2d3a10-...." width={40} />
      </InlineField>

      <InlineField label="MAC Address" labelWidth={16} tooltip="Client MAC (xx:xx:xx:xx:xx:xx). Supports variables.">
        <Input value={macAddress} onChange={onText(setMacAddress)} placeholder="aa:bb:cc:dd:ee:ff" width={24} />
      </InlineField>

      <InlineField
        label="Priority"
        labelWidth={16}
        tooltip="DNAC issue priority. CSV allowed (e.g. P1,P2). Supports variables."
      >
        <Input value={priority} onChange={onText(setPriority)} placeholder="P1,P2,P3,P4" width={24} />
      </InlineField>

      <InlineField
        label="Issue Status"
        labelWidth={16}
        tooltip="DNAC issueStatus. CSV allowed (e.g. ACTIVE,IGNORED). Supports variables."
      >
        <Input value={issueStatus} onChange={onText(setIssueStatus)} placeholder="ACTIVE,IGNORED,RESOLVED" width={28} />
      </InlineField>

      <InlineField label="AI Driven" labelWidth={16} tooltip="YES / NO (leave blank for any).">
        <Select
          options={aiOptions}
          value={aiOptions.find((o) => o.value === (aiDriven || '')) || aiOptions[0]}
          onChange={(v) => setAiDriven(v.value ?? '')}
          width={20}
        />
      </InlineField>

      <InlineField label="Limit" labelWidth={16} tooltip="Max issues to return (panel-side cap).">
        <Input
          type="number"
          value={limit}
          onChange={(e) => setLimit(Number(e.currentTarget.value) || 100)}
          width={12}
        />
      </InlineField>
    </Stack>
  );
}
