// ConfigEditor: Grafana plugin configuration UI for Catalyst datasource.
// Allows users to set connection details, credentials, and security options.
// All logic is handled via controlled components and Grafana's plugin API.
import React, { ChangeEvent, useState } from 'react';
import type { DataSourcePluginOptionsEditorProps, SelectableValue } from '@grafana/data';
import { Field, Input, SecretInput, Switch, Select } from '@grafana/ui';
import type { CatalystJsonData } from '../types';
import { HelpTooltip } from './HelpTooltip';
import { validateConfig } from '../validation';

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
  
  // State for validation errors
  const [validationError, setValidationError] = useState<string>();

  // setJson: Update non-secure config fields
  const setJson = (patch: Partial<CatalystJsonData>) => {
    const updated = { ...(jsonData ?? {}), ...patch };
    onOptionsChange({ ...options, jsonData: updated });
    
    // Validate on change
    const validation = validateConfig(updated);
    if (!validation.valid) {
      setValidationError(validation.errors.join(', '));
    } else {
      setValidationError(undefined);
    }
  };

  // Endpoint options for selection (Grafana Select format)
  const endpointOptions: Array<SelectableValue<string>> = [
    { label: 'Issues/Alerts', value: 'alerts' },
    { label: 'Site Health', value: 'siteHealth' },
  ];

  // Handler: Update endpoint selection for Grafana Select
  const onEndpointChange = (v: SelectableValue<string>) => setJson({ endpoint: v.value });

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
      {/* Display validation errors if any */}
      {validationError && (
        <div
          style={{
            padding: '8px 12px',
            backgroundColor: '#f44336',
            color: 'white',
            borderRadius: '4px',
            marginBottom: '8px',
          }}
        >
          ⚠️ Configuration Error: {validationError}
        </div>
      )}

      {/* Endpoint selection dropdown (Grafana Select) */}
      <Field
        label={
          <span>
            API Endpoint
            <HelpTooltip content="Select which Catalyst Center API endpoint to query. Alerts for issues/alerts, Site Health for site metrics." />
          </span>
        }
        description="Choose which Catalyst Center API endpoint to query."
      >
        <Select
          options={endpointOptions}
          value={endpointOptions.find((opt) => opt.value === (jsonData?.endpoint ?? 'alerts'))}
          onChange={onEndpointChange}
          width={30}
        />
      </Field>

      {/* Catalyst Base URL field. Proxy prefixes before /dna are preserved. */}
      <Field
        label={
          <span>
            Catalyst Base URL
            <HelpTooltip
              content="The base URL of your Cisco Catalyst Center instance. Include any proxy prefixes. Example: https://catalyst.example.com"
              link="https://github.com/extkljajicm/grafana-catalyst-datasource#configuration"
            />
          </span>
        }
        description="https://<host> . Proxy prefixes before /dna are preserved."
        invalid={validationError?.includes('Base URL')}
        error={validationError?.includes('Base URL') ? 'Invalid Base URL format' : undefined}
      >
        <Input
          value={jsonData?.baseUrl ?? ''}
          onChange={onBaseUrl}
          placeholder="https://catalyst.example.com"
          width={60}
          invalid={validationError?.includes('Base URL')}
        />
      </Field>

      {/* TLS verification toggle */}
      <Field
        label={
          <span>
            Skip TLS verification
            <HelpTooltip content="Only enable this for development/lab environments with self-signed certificates. Not recommended for production." />
          </span>
        }
        description="Disable TLS certificate verification (use with caution)"
      >
        <Switch
          value={!!jsonData?.insecureSkipVerify}
          onChange={(e) => setJson({ insecureSkipVerify: e.currentTarget.checked })}
        />
      </Field>

      {/* Username field (secure) */}
      <Field
        label={
          <span>
            Username
            <HelpTooltip content="Username for Catalyst Center API authentication. Used to obtain an X-Auth-Token." />
          </span>
        }
        description="Username for API authentication"
      >
        <Input
          value={secureJsonData?.username ?? ''}
          onChange={onUser}
          placeholder="dnac-api-user"
          width={40}
        />
      </Field>

      {/* Password field (secure, resettable) */}
      <Field
        label={
          <span>
            Password
            <HelpTooltip content="Password for Catalyst Center API authentication. Stored securely by Grafana." />
          </span>
        }
        description="Password for API authentication"
      >
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
      <Field
        label={
          <span>
            API Token (override)
            <HelpTooltip content="Optional pre-issued X-Auth-Token. If provided, bypasses username/password authentication. Leave empty to use username/password." />
          </span>
        }
        description="Optional: paste a pre-issued X-Auth-Token to bypass username/password"
      >
        <SecretInput
          isConfigured={!!secureJsonFields?.apiToken}
          value={secureJsonData?.apiToken}
          onChange={onToken}
          onReset={onResetToken}
          placeholder="(optional) Paste X-Auth-Token"
          width={40}
        />
      </Field>
    </div>
  );
};
