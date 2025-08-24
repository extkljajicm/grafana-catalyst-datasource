import React, { ChangeEvent, useCallback, useEffect, useState } from 'react';
import { InlineField, Input, Stack } from '@grafana/ui';
import { QueryEditorProps } from '@grafana/data';
import { DataSource } from '../datasource';
import { CatalystQuery, CatalystJsonData } from '../types';

type Props = QueryEditorProps<DataSource, CatalystQuery, CatalystJsonData>;

// Simple debounce hook
function useDebounced<T>(value: T, delayMs: number): T {
  const [debounced, setDebounced] = useState(value);
  useEffect(() => {
    const handle = setTimeout(() => setDebounced(value), delayMs);
    return () => clearTimeout(handle);
  }, [value, delayMs]);
  return debounced;
}

export function QueryEditor({ query, onChange, onRunQuery }: Props) {
  // Local UI state for debouncing
  const [severity, setSeverity] = useState(query.severity ?? '');
  const [status, setStatus] = useState(query.status ?? '');
  const [text, setText] = useState(query.text ?? '');
  const [limit, setLimit] = useState(query.limit ?? 100);

  const debouncedSeverity = useDebounced(severity, 400);
  const debouncedStatus = useDebounced(status, 400);
  const debouncedText = useDebounced(text, 400);
  const debouncedLimit = useDebounced(limit, 400);

  // Push debounced values into the query + trigger run
  useEffect(() => {
    onChange({ ...query, severity: debouncedSeverity, status: debouncedStatus, text: debouncedText, limit: debouncedLimit });
    onRunQuery();
  }, [debouncedSeverity, debouncedStatus, debouncedText, debouncedLimit]);

  const handleChange = useCallback(
    (setter: React.Dispatch<React.SetStateAction<string | number>>) => (e: ChangeEvent<HTMLInputElement>) => {
      setter(e.currentTarget.value);
    },
    []
  );

  return (
    <Stack gap={1}>
      <InlineField label="Severity" labelWidth={14} tooltip="Comma-separated severities (e.g. P1,P2,P3)">
        <Input
          value={severity}
          onChange={handleChange(setSeverity)}
          placeholder="P1,P2,P3"
          width={20}
        />
      </InlineField>
      <InlineField label="Status" labelWidth={14} tooltip="Comma-separated statuses (e.g. ACTIVE,CLEARED)">
        <Input
          value={status}
          onChange={handleChange(setStatus)}
          placeholder="ACTIVE,CLEARED"
          width={20}
        />
      </InlineField>
      <InlineField label="Text" labelWidth={14} tooltip="Free-text search filter">
        <Input
          value={text}
          onChange={handleChange(setText)}
          placeholder="search text…"
          width={40}
        />
      </InlineField>
      <InlineField label="Limit" labelWidth={14} tooltip="Maximum alerts to fetch">
        <Input
          type="number"
          value={limit}
          onChange={(e) => setLimit(Number(e.currentTarget.value) || 100)}
          width={10}
        />
      </InlineField>
    </Stack>
  );
}
