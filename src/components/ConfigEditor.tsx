import React, { ChangeEvent, useState } from 'react';
import type { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { Stack, Field, Input, SecretInput, Switch, Alert } from '@grafana/ui';
import type { CatalystJsonData } from '../types';

type SecureShape = {
  username?: string;
  password?: string;
  apiToken?: string;
};

type Props = DataSourcePluginOptionsEditorProps<CatalystJsonData, SecureShape>;

export const ConfigEditor: React.FC<Props> = ({ options, onOptionsChange }) => {
  const { jsonData, secureJsonData, secureJsonFields } = options;

  const [authOpen, setAuthOpen] = useState(true);

  const setJson = (patch: Partial<CatalystJsonData>) =>
    onOptionsChange({ ...options, jsonData: { ...(jsonData ?? {}), ...patch } });

  const setSecure = (patch: Partial<SecureShape>) =>
    onOptionsChange({ ...options, secureJsonData: { ...(secureJsonData ?? {}), ...patch } });

  const onBaseUrl = (e: ChangeEvent<HTMLInputElement>) => setJson({ baseUrl: e.currentTarget.value });
  const onUser = (e: ChangeEvent<HTMLInputElement>) => setSecure({ username: e.currentTarget.value });

  const onPass = (e: ChangeEvent<HTMLInputElement>) => setSecure({ password: e.currentTarget.value });
  const onResetPass = () =>
    onOptionsChange({
      ...options,
      secureJsonFields: { ...(secureJsonFields ?? {}), password: false },
      secureJsonData: { ...(secureJsonData ?? {}), password: '' },
    });

  const onToken = (e: ChangeEvent<HTMLInputElement>) => setSecure({ apiToken: e.currentTarget.value });
  const onResetToken = () =>
    onOptionsChange({
      ...options,
      secureJsonFields: { ...(secureJsonFields ?? {}), apiToken: false },
      secureJsonData: { ...(secureJsonData ?? {}), apiToken: '' },
    });

  return (
    <Stack gap={3}>
      {/* -------- Server -------- */}
      <div className="gf-form-group">
        <h3 className="page-heading">Server</h3>
        <Field label="Catalyst Base URL" description="Example: https://dnac.example.com/dna/intent/api/v1">
          <Input
            value={jsonData?.baseUrl ?? ''}
            onChange={onBaseUrl}
            placeholder="https://<host>/dna/intent/api/v1"
            width={60}
          />
        </Field>
      </div>

      {/* -------- Authentication (collapsible) -------- */}
      <div className="gf-form-group">
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <h3 className="page-heading" style={{ margin: 0 }}>Authentication</h3>
          <button
            type="button"
            className="btn btn-secondary"
            onClick={() => setAuthOpen((v) => !v)}
            aria-expanded={authOpen}
            aria-controls="auth-section"
          >
            {authOpen ? 'Hide' : 'Show'}
          </button>
        </div>

        {authOpen && (
          <div id="auth-section" style={{ marginTop: 8 }}>
            <Alert title="Auth model" severity="info">
              Backend logs in with username/password to fetch a short-lived X-Auth-Token.
              You can also paste a token manually (override).
            </Alert>

            <Field label="Username" description="Catalyst Center API user">
              <Input value={secureJsonData?.username ?? ''} onChange={onUser} placeholder="dnac-api-user" width={40} />
            </Field>

            <Field label="Password" description="Stored securely by Grafana">
              <SecretInput
                isConfigured={!!secureJsonFields?.password}
                value={secureJsonData?.password}
                onChange={onPass}
                onReset={onResetPass}
                placeholder="••••••••"
                width={40}
              />
            </Field>

            <Field
              label="API Token (override)"
              description="Paste an existing X-Auth-Token to bypass username/password login (optional)."
            >
              <SecretInput
                isConfigured={!!secureJsonFields?.apiToken}
                value={secureJsonData?.apiToken}
                onChange={onToken}
                onReset={onResetToken}
                placeholder="Optional: paste existing X-Auth-Token"
                width={60}
              />
            </Field>
          </div>
        )}
      </div>

      {/* -------- Security -------- */}
      <div className="gf-form-group">
        <h3 className="page-heading">Security</h3>
        <Field
          label="Skip TLS verification"
          description="Disable TLS certificate verification (use only with self-signed certs in lab/test)."
        >
          <Switch
            value={!!jsonData?.insecureSkipVerify}
            onChange={(e) => setJson({ insecureSkipVerify: e.currentTarget.checked })}
          />
        </Field>
      </div>
    </Stack>
  );
};

export default ConfigEditor;
