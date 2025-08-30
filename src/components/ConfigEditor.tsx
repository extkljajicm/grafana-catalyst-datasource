import React, { ChangeEvent } from 'react';
import type { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { Field, Input, SecretInput, Switch } from '@grafana/ui';
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
      secureJsonFields: { ...(options.secureJsonFields ?? {}), password: false },
      secureJsonData: { ...(options.secureJsonData ?? {}), password: '' },
    });

  const onToken = (e: ChangeEvent<HTMLInputElement>) => setSecure({ apiToken: e.currentTarget.value });
  const onResetToken = () =>
    onOptionsChange({
      ...options,
      secureJsonFields: { ...(options.secureJsonFields ?? {}), apiToken: false },
      secureJsonData: { ...(options.secureJsonData ?? {}), apiToken: '' },
    });

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
      <Field label="Catalyst Base URL" description="Can be https://<host> or https://<host>/dna/intent/api/v1. Proxy prefixes before /dna are preserved.">
        <Input
          value={jsonData?.baseUrl ?? ''}
          onChange={onBaseUrl}
          placeholder="https://<host>  or  https://<host>/dna/intent/api/v1"
          width={60}
        />
      </Field>

      <Field label="Skip TLS verification">
        <Switch
          value={!!jsonData?.insecureSkipVerify}
          onChange={(e) => setJson({ insecureSkipVerify: e.currentTarget.checked })}
        />
      </Field>

      <Field label="Username">
        <Input
          value={secureJsonData?.username ?? ''}
          onChange={onUser}
          placeholder="dnac-api-user"
          width={40}
        />
      </Field>

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

export default ConfigEditor;
