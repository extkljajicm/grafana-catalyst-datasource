// VariableQueryEditor: UI for configuring template variable queries in Grafana.
// Lets users select query type (priorities, statuses, sites, devices, MACs) and optional search.
// Generates query definition for display and backend use.
import React, { useMemo } from 'react';
import { Field, Input, Select } from '@grafana/ui';
import type { SelectableValue } from '@grafana/data';
import type { CatalystVariableQuery } from '../types';

// Props: query object and change handler from Grafana
type VarQEProps = {
  query: CatalystVariableQuery | undefined;
  onChange: (query: CatalystVariableQuery, definition?: string) => void;
};

// QType: Supported variable query types
type QType = CatalystVariableQuery['type'];

// Dropdown options for variable query types
const TYPE_OPTIONS: Array<SelectableValue<QType>> = [
  { label: 'Priorities (P1..P4)', value: 'priorities' },
  { label: 'Issue Statuses', value: 'issueStatuses' },
  { label: 'Sites', value: 'sites' },
  { label: 'Devices', value: 'devices' },
  { label: 'MACs', value: 'macs' },
];

// Helper: Build CatalystVariableQuery object from type and search
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

// Helper: Generate query definition string for display and backend
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

// VariableQueryEditor main component
const VariableQueryEditor: React.FC<VarQEProps> = ({ query, onChange }) => {
  // Extract type and search from query object
  const type: QType = query?.type ?? 'priorities';
  const search = query && 'search' in query ? (query.search ?? '') : '';

  // Memoized query definition string
  const definition = useMemo(() => {
    const q = varQueryFrom(type, search);
    return definitionFor(q);
  }, [type, search]);

  // Handler: Update query type
  const updateType = (t?: QType) => {
    const q = varQueryFrom(t ?? 'priorities', search);
    onChange(q, definitionFor(q));
  };

  // Handler: Update search string
  const updateSearch = (s: string) => {
    const q = varQueryFrom(type, s);
    onChange(q, definitionFor(q));
  };

  // Render: Variable query configuration form
  return (
    <div className="gf-form-group">
      {/* Query type dropdown */}
      <Field label="Type">
        <Select
          options={TYPE_OPTIONS}
          value={TYPE_OPTIONS.find((o) => o.value === type) ?? TYPE_OPTIONS[0]}
          onChange={(v) => updateType(v?.value as QType)}
        />
      </Field>

      {/* Search field for types that support it */}
      {type !== 'priorities' && type !== 'issueStatuses' && (
        <Field label="Search (optional)">
          <Input
            value={search}
            onChange={(e) => updateSearch(e.currentTarget.value)}
            placeholder="e.g. branch-a, 00:11:22  (optional; supports variables)"
          />
        </Field>
      )}

      {/* Query definition display */}
      <Field label="Definition">
        <Input value={definition} readOnly />
      </Field>
    </div>
  );
};

// Default export for plugin registration
export default VariableQueryEditor;
