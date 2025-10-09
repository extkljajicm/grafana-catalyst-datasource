// QueryEditor: Main query configuration UI for Catalyst datasource in Grafana.
// Allows users to filter alerts/issues by site, device, MAC, priority, status, AI-driven, and more.
// Uses debounced local state to avoid excessive backend requests.
import React, { useEffect, useRef, useState } from 'react';
import { Field, Input, InlineField, MultiSelect, Switch } from '@grafana/ui';
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

const siteHealthMetricOptions: Array<SelectableValue<string>> = [
  { label: 'Client Count', value: 'clientCount' },
  { label: 'Total Connected Wired Clients', value: 'totalNumberOfConnectedWiredClients' },
  { label: 'Total Active Wireless Clients', value: 'totalNumberOfActiveWirelessClients' },
  { label: 'Healthy Network Device %', value: 'healthyNetworkDevicePercentage' },
  { label: 'Healthy Clients %', value: 'healthyClientsPercentage' },
  { label: 'Wired Client Health', value: 'clientHealthWired' },
  { label: 'Wireless Client Health', value: 'clientHealthWireless' },
  { label: 'Number of Network Devices', value: 'numberOfNetworkDevice' },
  { label: 'Network Health Average', value: 'networkHealthAverage' },
  { label: 'Network Health (Access)', value: 'networkHealthAccess' },
  { label: 'Network Health (Core)', value: 'networkHealthCore' },
  { label: 'Network Health (AP)', value: 'networkHealthAP' },
  { label: 'Network Health (WLC)', value: 'networkHealthWLC' },
  { label: 'Network Health (Switch)', value: 'networkHealthSwitch' },
  { label: 'Access Devices (Total)', value: 'accessTotalCount' },
  { label: 'Access Devices (Good)', value: 'accessGoodCount' },
  { label: 'AP Devices (Good)', value: 'apDeviceGoodCount' },
  { label: 'AP Devices (Total)', value: 'apDeviceTotalCount' },
  { label: 'Switch Devices (Good)', value: 'switchDeviceGoodCount' },
  { label: 'Switch Devices (Total)', value: 'switchDeviceTotalCount' },
];

// Define a type for the filter state
type Filters = Omit<Partial<CatalystQuery>, 'refId' | 'queryType' | 'endpoint'>;

const QueryEditor: React.FC<Props> = ({ query, onChange, onRunQuery, range }) => {
  // Get endpoint from config
  const endpoint = query.endpoint ?? 'alerts';

  // Unified state for all filters
  const [filters, setFilters] = useState<Filters>({
    siteId: query.siteId ?? DEFAULT_QUERY.siteId,
    networkDeviceId: query.networkDeviceId ?? DEFAULT_QUERY.networkDeviceId,
    macAddress: query.macAddress ?? DEFAULT_QUERY.macAddress,
    priority: query.priority ?? DEFAULT_QUERY.priority,
    status: query.status ?? DEFAULT_QUERY.status,
    limit: query.limit ?? DEFAULT_QUERY.limit,
    metrics: query.metrics ?? DEFAULT_QUERY.metrics,
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

  // Render common and endpoint-specific filters
  return (
    <div className="gf-form-group">
      {endpoint === 'siteHealth' ? (
        <>
          <InlineField label="Site Type" labelWidth={14}>
            <Input
              width={40}
              value={filters.siteType}
              onChange={(e) => setFilters({ ...filters, siteType: e.currentTarget.value })}
              placeholder="e.g., BUILDING, AREA"
            />
          </InlineField>
          <InlineField label="Parent Site" labelWidth={14}>
            <Input
              width={40}
              value={filters.parentSiteName}
              onChange={(e) => setFilters({ ...filters, parentSiteName: e.currentTarget.value })}
              placeholder="Filter by parent site name"
            />
          </InlineField>
          <InlineField label="Site Name" labelWidth={14}>
            <Input
              width={40}
              value={filters.siteName}
              onChange={(e) => setFilters({ ...filters, siteName: e.currentTarget.value })}
              placeholder="Filter by site name"
            />
          </InlineField>
          <Field label="Metrics">
            <MultiSelect
              options={siteHealthMetricOptions}
              value={filters.metrics}
              onChange={(v) => setFilters({ ...filters, metrics: v.map((item) => item.value!) })}
            />
          </Field>
        </>
      ) : (
        <>
          <InlineField label="Site Name" labelWidth={14}>
            <Input
              width={40}
              value={filters.siteName}
              onChange={(e) => setFilters({ ...filters, siteName: e.currentTarget.value, siteId: '' })}
              placeholder="Enter site name (will resolve to ID)"
            />
          </InlineField>
          <InlineField label="Device ID" labelWidth={14}>
            <Input
              width={40}
              value={filters.networkDeviceId}
              onChange={(e) => setFilters({ ...filters, networkDeviceId: e.currentTarget.value })}
              placeholder="Enter device UUID"
            />
          </InlineField>
          <InlineField label="MAC Address" labelWidth={14}>
            <Input
              width={40}
              value={filters.macAddress}
              onChange={(e) => setFilters({ ...filters, macAddress: e.currentTarget.value })}
              placeholder="Enter MAC address"
            />
          </InlineField>
          <Field label="Priority">
            <MultiSelect
              options={PRIORITY_OPTIONS}
              value={filters.priority}
              onChange={(v) => setFilters({ ...filters, priority: v.map((item) => item.value!) })}
            />
          </Field>
          <Field label="Status">
            <MultiSelect
              options={STATUS_OPTIONS}
              value={filters.status}
              onChange={(v) => setFilters({ ...filters, status: v.map((item) => item.value!) })}
            />
          </Field>
          <InlineField label="Limit" labelWidth={14}>
            <Input
              width={20}
              type="number"
              value={filters.limit}
              onChange={(e) => setFilters({ ...filters, limit: parseInt(e.currentTarget.value, 10) })}
              placeholder="100"
            />
          </InlineField>
          <Field label="Enrich with Site Names">
            <Switch
              value={filters.enrich}
              onChange={(e) => setFilters({ ...filters, enrich: e.currentTarget.checked })}
            />
          </Field>
        </>
      )}
    </div>
  );
};

export default QueryEditor;
