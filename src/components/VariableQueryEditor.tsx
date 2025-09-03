import React, { useMemo } from 'react';
import { Field, Input, Select } from '@grafana/ui';
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

function varQueryFrom(type: QType, search?: string): CatalystVariableQuery {
  switch (type) {
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
    default:
      return { type: 'priorities' };
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
    default:
      return '';
  }
}

const VariableQueryEditor: React.FC<VarQEProps> = ({ query, onChange }) => {
  const type: QType = query?.type ?? 'priorities';
  const search = query && 'search' in query ? (query.search ?? '') : '';

  const definition = useMemo(() => {
    const q = varQueryFrom(type, search);
    return definitionFor(q);
  }, [type, search]);

  const updateType = (t?: QType) => {
    const q = varQueryFrom(t ?? 'priorities', search);
    onChange(q, definitionFor(q));
  };

  const updateSearch = (s: string) => {
    const q = varQueryFrom(type, s);
    onChange(q, definitionFor(q));
  };

  return (
    <div className="gf-form-group">
      <Field label="Type">
        <Select
          options={TYPE_OPTIONS}
          value={TYPE_OPTIONS.find((o) => o.value === type) ?? TYPE_OPTIONS[0]}
          onChange={(v) => updateType(v?.value as QType)}
        />
      </Field>

      {type !== 'priorities' && type !== 'issueStatuses' && (
        <Field label="Search (optional)">
          <Input
            value={search}
            onChange={(e) => updateSearch(e.currentTarget.value)}
            placeholder="e.g. branch-a, 00:11:22  (optional; supports variables)"
          />
        </Field>
      )}

      <Field label="Definition">
        <Input value={definition} readOnly />
      </Field>
    </div>
  );
};

export default VariableQueryEditor;
