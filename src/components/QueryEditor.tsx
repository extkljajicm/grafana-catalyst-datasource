// QueryEditor: Main query configuration UI for Catalyst datasource in Grafana.
// Allows users to filter alerts/issues by site, device, MAC, priority, status, AI-driven, and more.
// Uses debounced local state to avoid excessive backend requests.
import React, { useEffect, useRef, useState } from 'react';
import { Field, Input, InlineField, MultiSelect, Switch, Select } from '@grafana/ui';
import type { QueryEditorProps, SelectableValue } from '@grafana/data';
import { DataSource } from '../datasource';
import {
  CatalystQuery,
  CatalystJsonData,
  CatalystPriority,
  CatalystIssueStatus,
  DEFAULT_QUERY,
  QueryType,
} from '../types';

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

const METRIC_OPTIONS: Array<SelectableValue<string>> = [
  { label: 'Client Count', value: 'clientCount' },
  { label: 'Health Score', value: 'healthScore' },
  { label: 'Wired Client Count', value: 'wiredClientCount' },
  { label: 'Wireless Client Count', value: 'wirelessClientCount' },
];

// Define a type for the filter state
type Filters = Omit<Partial<CatalystQuery>, 'refId' | 'queryType' | 'endpoint'>;

const QueryEditor: React.FC<Props> = ({ query, onChange, onRunQuery, range }) => {
  // Get endpoint from config
  const endpoint = query.endpoint ?? 'alerts';

  // Unified state for all filters
  const [filters, setFilters] = useState<Filters>({
    siteId: query.siteId ?? DEFAULT_QUERY.siteId,
    deviceId: query.deviceId ?? DEFAULT_QUERY.deviceId,
    macAddress: query.macAddress ?? DEFAULT_QUERY.macAddress,
    priority: query.priority ?? DEFAULT_QUERY.priority,
    issueStatus: query.issueStatus ?? DEFAULT_QUERY.issueStatus,
    aiDriven: query.aiDriven ?? DEFAULT_QUERY.aiDriven,
    limit: query.limit ?? DEFAULT_QUERY.limit,
    metric: query.metric ?? DEFAULT_QUERY.metric,
    parentSiteName: query.parentSiteName ?? DEFAULT_QUERY.parentSiteName,
    siteName: query.siteName ?? DEFAULT_QUERY.siteName,
    enrich: query.enrich ?? DEFAULT_QUERY.enrich,
  });

  // Debounced version of the filters
  const debouncedFilters = useDebounced(filters, 400);

  // Effect: Synchronize debounced state with parent query object
  // Prevents infinite loops by tracking last signature
  const lastSig = useRef<string>('');
  useEffect(() => {
    const next: CatalystQuery = {
      ...query,
      queryType: endpoint as QueryType,
      ...debouncedFilters,
    };
    const sig = JSON.stringify(next);

    if (sig !== lastSig.current) {
      lastSig.current = sig;
      onChange(next);
      onRunQuery();
    }
  }, [debouncedFilters, endpoint, onChange, onRunQuery, query]);

  // Handler for filter changes
  const onFilterChange = (patch: Partial<Filters>) => {
    setFilters((prev) => ({ ...prev, ...patch }));
  };

  // Render common and endpoint-specific filters
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
      {endpoint === 'alerts' && (
        <>
          <div style={{ display: 'flex', gap: 8 }}>
            <Field label="Site ID" description="Filter by Catalyst site ID (UUID)">
              <Input
                value={filters.siteId}
                onChange={(e) => onFilterChange({ siteId: e.currentTarget.value })}
                placeholder="All sites"
                width={30}
              />
            </Field>
            <Field label="Device ID" description="Filter by device IP address">
              <Input
                value={filters.deviceId}
                onChange={(e) => onFilterChange({ deviceId: e.currentTarget.value })}
                placeholder="All devices"
                width={30}
              />
            </Field>
          </div>
          <div style={{ display: 'flex', gap: 8 }}>
            <Field label="MAC Address" description="Filter by client MAC address">
              <Input
                value={filters.macAddress}
                onChange={(e) => onFilterChange({ macAddress: e.currentTarget.value })}
                placeholder="All clients"
                width={30}
              />
            </Field>
            <Field label="Priority" description="Select one or more priorities">
              <MultiSelect
                options={PRIORITY_OPTIONS}
                value={filters.priority}
                onChange={(v) => onFilterChange({ priority: v.map((item) => item.value!) })}
                width={30}
              />
            </Field>
          </div>
          <div style={{ display: 'flex', gap: 8 }}>
            <Field label="Status" description="Filter by issue status">
              <Select
                options={STATUS_OPTIONS}
                value={filters.issueStatus}
                onChange={(v) => onFilterChange({ issueStatus: v?.value as CatalystIssueStatus })}
                width={30}
                isClearable
              />
            </Field>
            <Field label="AI-Driven" description="Filter by AI-driven issues">
              <Select
                options={AI_OPTIONS}
                value={filters.aiDriven}
                onChange={(v) => onFilterChange({ aiDriven: v?.value ?? '' })}
                width={30}
              />
            </Field>
          </div>
          <div style={{ display: 'flex', gap: 8 }}>
            <Field label="Enrich" description="Resolve site IDs to names (slower)">
              <Switch
                value={filters.enrich}
                onChange={(e) => onFilterChange({ enrich: e.currentTarget.checked })}
              />
            </Field>
          </div>
        </>
      )}
      {endpoint === 'siteHealth' && (
        <>
          <div style={{ display: 'flex', gap: 8 }}>
            <InlineField label="Parent Site Name" labelWidth={20}>
              <Input
                width={40}
                value={filters.parentSiteName}
                onChange={(e) => onFilterChange({ parentSiteName: e.currentTarget.value })}
                placeholder="Filter by parent site name"
              />
            </InlineField>
            <InlineField label="Site Name" labelWidth={20}>
              <Input
                width={40}
                value={filters.siteName}
                onChange={(e) => onFilterChange({ siteName: e.currentTarget.value })}
                placeholder="Filter by site name"
              />
            </InlineField>
          </div>
          <div className="gf-form">
            <InlineField label="Parent Site ID" labelWidth={20}>
              <Input
                width={40}
                value={filters.parentSiteId}
                onChange={(e) => onFilterChange({ parentSiteId: e.currentTarget.value })}
                placeholder="Filter by parent site ID"
              />
            </InlineField>
            <InlineField label="Site ID" labelWidth={20}>
              <Input
                width={40}
                value={filters.siteId}
                onChange={(e) => onFilterChange({ siteId: e.currentTarget.value })}
                placeholder="Filter by site ID"
              />
            </InlineField>
          </div>
          <div className="gf-form">
            <InlineField label="Metrics" labelWidth={20}>
              <MultiSelect
                width={40}
                options={METRIC_OPTIONS}
                value={filters.metric}
                onChange={(v) => onFilterChange({ metric: v.map((item) => item.value!) })}
              />
            </InlineField>
          </div>
        </>
      )}
      <div style={{ display: 'flex', gap: 8 }}>
        <Field label="Limit" description="Maximum number of issues to return">
          <Input
            type="number"
            value={filters.limit}
            onChange={(e) => onFilterChange({ limit: parseInt(e.currentTarget.value, 10) || 0 })}
            width={15}
          />
        </Field>
      </div>
    </div>
  );
};

export default QueryEditor;
