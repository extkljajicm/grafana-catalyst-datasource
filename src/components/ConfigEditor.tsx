// ConfigEditor: Grafana plugin configuration UI for Catalyst datasource.
// Allows users to set connection details, credentials, and security options.
// All logic is handled via controlled components and Grafana's plugin API.
import React, { ChangeEvent } from 'react';
import type { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { Field, Input, SecretInput, Switch } from '@grafana/ui';
import { Select } from '@grafana/ui';
import type { CatalystJsonData } from '../types';

// SecureShape: Structure for secure fields (not stored in plain config)
type SecureShape = {
  username?: string;
  password?: string;
  apiToken?: string;
};

// Props: Grafana passes plugin config and change handler
type Props = DataSourcePluginOptionsEditorProps<CatalystJsonData, SecureShape>;

// ConfigEditor main component
export const ConfigEditor: React.FC<Props> = ({ options, onOptionsChange }) => {
  // Destructure config objects for clarity
  const { jsonData, secureJsonData, secureJsonFields } = options;

  // setJson: Update non-secure config fields
  const setJson = (patch: Partial<CatalystJsonData>) =>
    onOptionsChange({ ...options, jsonData: { ...(jsonData ?? {}), ...patch } });

  // Endpoint options for selection (Grafana Select format)
  const endpointOptions = [
    { label: 'Issues/Alerts', value: 'alerts' },
    { label: 'Site Health', value: 'siteHealth' },
  ];

  // Handler: Update endpoint selection for Grafana Select
  const onEndpointChange = (v: any) => setJson({ endpoint: v?.value ?? 'alerts' });

  // setSecure: Update secure config fields (username, password, token)
  const setSecure = (patch: Partial<SecureShape>) =>
    onOptionsChange({ ...options, secureJsonData: { ...(secureJsonData ?? {}), ...patch } });

  // Handler: Update base URL field
  const onBaseUrl = (e: ChangeEvent<HTMLInputElement>) => setJson({ baseUrl: e.currentTarget.value });
  // Handler: Update username field
  const onUser = (e: ChangeEvent<HTMLInputElement>) => setSecure({ username: e.currentTarget.value });

  // Handler: Update password field
  const onPass = (e: ChangeEvent<HTMLInputElement>) => setSecure({ password: e.currentTarget.value });
  // Handler: Reset password (marks as not configured)
  const onResetPass = () =>
    onOptionsChange({
      ...options,
      secureJsonFields: { ...(options.secureJsonFields ?? {}), password: false },
      secureJsonData: { ...(options.secureJsonData ?? {}), password: '' },
    });

  // Handler: Update API token field
  const onToken = (e: ChangeEvent<HTMLInputElement>) => setSecure({ apiToken: e.currentTarget.value });
  // Handler: Reset API token (marks as not configured)
  const onResetToken = () =>
    onOptionsChange({
      ...options,
      secureJsonFields: { ...(options.secureJsonFields ?? {}), apiToken: false },
      secureJsonData: { ...(options.secureJsonData ?? {}), apiToken: '' },
    });

  // Render: Form fields for all config options
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
      {/* Endpoint selection dropdown (Grafana Select) */}
      <Field label="API Endpoint" description="Choose which Catalyst Center API endpoint to query.">
        <Select
          options={endpointOptions}
          value={endpointOptions.find(opt => opt.value === (jsonData?.endpoint ?? 'alerts'))}
          onChange={onEndpointChange}
          width={30}
        />
      </Field>

      {/* Catalyst Base URL field. Proxy prefixes before /dna are preserved. */}
      <Field label="Catalyst Base URL" description="https://<host> . Proxy prefixes before /dna are preserved.">
        <Input
          value={jsonData?.baseUrl ?? ''}
          onChange={onBaseUrl}
          placeholder="https://<host>"
          width={60}
        />
      </Field>

      {/* TLS verification toggle */}
      <Field label="Skip TLS verification">
        <Switch
          value={!!jsonData?.insecureSkipVerify}
          onChange={(e) => setJson({ insecureSkipVerify: e.currentTarget.checked })}
        />
      </Field>

      {/* Username field (secure) */}
      <Field label="Username">
        <Input
          value={secureJsonData?.username ?? ''}
          onChange={onUser}
          placeholder="dnac-api-user"
          width={40}
        />
      </Field>

      {/* Password field (secure, resettable) */}
      <Field label="Password">
        <SecretInput
          isConfigured={!!secureJsonFields?.password}
          value={secureJsonData?.password}
          onChange={onPass}
          onReset={onResetPass}
          placeholder="••••••••"
          width={40}
        />
      </Field>

      {/* API Token field (secure, optional, overrides password) */}
      <Field label="API Token (override)">
        <SecretInput
          isConfigured={!!secureJsonFields?.apiToken}
          value={secureJsonData?.apiToken}
          onChange={onToken}
          onReset={onResetToken}
          placeholder="(optional) Paste X-Auth-Token"
          width={40}   // narrower, matches username/password
        />
      </Field>
    </div>
  );
};

// Default export for plugin registration
export default ConfigEditor;
