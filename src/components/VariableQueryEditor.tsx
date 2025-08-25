import React, { useMemo } from 'react';
import { InlineField, InlineFieldRow, Input, Select } from '@grafana/ui';
import type { SelectableValue } from '@grafana/data';
import type { CatalystVariableQuery } from '../types';

type VarQEProps = {
  query: CatalystVariableQuery | undefined;
  onChange: (query: CatalystVariableQuery, definition?: string) => void;
};

type QType = CatalystVariableQuery['type'];

const TYPE_OPTIONS: Array<SelectableValue<QType>> = [
  { label: 'Priorities (P1..P4)', value: 'priorities' },
  { label: 'Issue Statuses', value: 'issueStatuses' },
  { label: 'Sites', value: 'sites' },
  { label: 'Devices', value: 'devices' },
  { label: 'MACs', value: 'macs' },
];

function getType(q?: CatalystVariableQuery): QType {
  if (!q) return 'priorities';
  return q.type;
}

function getSearch(q?: CatalystVariableQuery): string {
  if (!q) return '';
  switch (q.type) {
    case 'sites':
    case 'devices':
    case 'macs':
      return q.search ?? '';
    default:
      return '';
  }
}

function buildQuery(t: QType, search?: string): CatalystVariableQuery {
  switch (t) {
    case 'priorities':
      return { type: 'priorities' };
    case 'issueStatuses':
      return { type: 'issueStatuses' };
    case 'sites':
      return { type: 'sites', search: (search ?? '').trim() || undefined };
    case 'devices':
      return { type: 'devices', search: (search ?? '').trim() || undefined };
    case 'macs':
      return { type: 'macs', search: (search ?? '').trim() || undefined };
  }
}

function definitionFor(q: CatalystVariableQuery): string {
  switch (q.type) {
    case 'priorities':
      return 'priorities()';
    case 'issueStatuses':
      return 'issueStatuses()';
    case 'sites':
      return q.search ? `sites(search:"${q.search}")` : 'sites()';
    case 'devices':
      return q.search ? `devices(search:"${q.search}")` : 'devices()';
    case 'macs':
      return q.search ? `macs(search:"${q.search}")` : 'macs()';
  }
}

export function VariableQueryEditor({ query, onChange }: VarQEProps): JSX.Element {
  const selectedType = getType(query);
  const search = getSearch(query);

  const definition = useMemo(() => {
    return definitionFor(buildQuery(selectedType, search));
  }, [selectedType, search]);

  const updateType = (t: QType) => {
    const next = buildQuery(t, search);
    onChange(next, definitionFor(next));
  };

  const updateSearch = (s: string) => {
    const next = buildQuery(selectedType, s);
    onChange(next, definitionFor(next));
  };

  return (
    <div className="gf-form-group">
      <InlineFieldRow>
        <InlineField label="Type" grow>
          <Select<QType>
            options={TYPE_OPTIONS}
            value={TYPE_OPTIONS.find((o) => o.value === selectedType) ?? TYPE_OPTIONS[0]}
            onChange={(v) => updateType((v.value ?? 'priorities') as QType)}
          />
        </InlineField>
      </InlineFieldRow>

      {(selectedType === 'sites' || selectedType === 'devices' || selectedType === 'macs') && (
        <InlineFieldRow>
          <InlineField label="Search" grow tooltip="Optional contains filter; supports variables.">
            <Input
              value={search}
              onChange={(e) => updateSearch(e.currentTarget.value)}
              placeholder="e.g. branch-a, 00:11:22"
            />
          </InlineField>
        </InlineFieldRow>
      )}

      <InlineFieldRow>
        <InlineField label="Definition" grow tooltip="Preview of the variable definition shown by Grafana">
          <Input value={definition} readOnly />
        </InlineField>
      </InlineFieldRow>
    </div>
  );
}

export default VariableQueryEditor;
