import React, { ChangeEvent } from 'react';
import type { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { Field, Input, SecretInput, Stack, InlineField, InlineFieldRow, Alert, Switch } from '@grafana/ui';
import type { CatalystJsonData } from '../types';

type SecureShape = {
  username?: string;
  password?: string;
  apiToken?: string;
};

type Props = DataSourcePluginOptionsEditorProps<CatalystJsonData, SecureShape>;

export const ConfigEditor: React.FC<Props> = ({ options, onOptionsChange }) => {
  const { jsonData, secureJsonData, secureJsonFields } = options;

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

  const onToggleInsecure = (v: boolean) => setJson({ insecureSkipVerify: v });

  return (
    <Stack gap={2}>
      <Field label="Catalyst Base URL" description="Example: https://dnac.example.com/dna/intent/api/v1">
        <Input
          value={jsonData?.baseUrl ?? ''}
          onChange={onBaseUrl}
          placeholder="https://<host>/dna/intent/api/v1"
          width={60}
        />
      </Field>

      <InlineFieldRow>
        <InlineField label="Skip TLS verification" tooltip="Disable TLS cert verification (use only for self‑signed certs in lab)">
          <Switch value={!!jsonData?.insecureSkipVerify} onChange={(e) => onToggleInsecure(e.currentTarget.checked)} />
        </InlineField>
      </InlineFieldRow>

      <Alert title="Auth model" severity="info">
        Backend logs in with username/password to fetch a short‑lived X‑Auth‑Token. You can also paste a token manually (override).
      </Alert>

      <InlineFieldRow>
        <InlineField label="Username" grow>
          <Input value={secureJsonData?.username ?? ''} onChange={onUser} placeholder="dnac-api-user" width={40} />
        </InlineField>
      </InlineFieldRow>

      <InlineFieldRow>
        <InlineField label="Password" grow>
          <SecretInput
            isConfigured={!!secureJsonFields?.password}
            value={secureJsonData?.password}
            onChange={onPass}
            onReset={onResetPass}
            placeholder="••••••••"
            width={40}
          />
        </InlineField>
      </InlineFieldRow>

      <InlineFieldRow>
        <InlineField label="API Token (override)" grow>
          <SecretInput
            isConfigured={!!secureJsonFields?.apiToken}
            value={secureJsonData?.apiToken}
            onChange={onToken}
            onReset={onResetToken}
            placeholder="Optional: paste existing X‑Auth‑Token"
            width={60}
          />
        </InlineField>
      </InlineFieldRow>
    </Stack>
  );
};

export default ConfigEditor;
