import React, { ChangeEvent } from 'react';
import { InlineField, Input, SecretInput } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { CatalystJsonData, CatalystSecureJsonData } from '../types';

type Props = DataSourcePluginOptionsEditorProps<CatalystJsonData, CatalystSecureJsonData>;

export function ConfigEditor(props: Props) {
  const { onOptionsChange, options } = props;
  const { jsonData, secureJsonFields, secureJsonData } = options;

  const onBaseUrlChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      jsonData: {
        ...jsonData,
        baseUrl: event.target.value,
      },
    });
  };

  // Secure field (only stored server-side)
  const onApiTokenChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      secureJsonData: {
        apiToken: event.target.value,
      },
    });
  };

  const onResetApiToken = () => {
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        apiToken: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        apiToken: '',
      },
    });
  };

  return (
    <>
      <InlineField
        label="Base URL"
        labelWidth={14}
        interactive
        tooltip={
          'Catalyst Center API base URL.\nExample: https://dnac.example.com/dna/intent/api/v1'
        }
      >
        <Input
          id="config-editor-base-url"
          onChange={onBaseUrlChange}
          value={jsonData.baseUrl || ''}
          placeholder="https://dnac.example.com/dna/intent/api/v1"
          width={50}
        />
      </InlineField>

      <InlineField
        label="API Token"
        labelWidth={14}
        interactive
        tooltip={'X-Auth-Token used for Catalyst Center requests (stored securely on the server).'}
      >
        <SecretInput
          required
          id="config-editor-api-token"
          isConfigured={Boolean(secureJsonFields?.apiToken)}
          value={secureJsonData?.apiToken || ''}
          placeholder="Enter your X-Auth-Token"
          width={50}
          onReset={onResetApiToken}
          onChange={onApiTokenChange}
        />
      </InlineField>
    </>
  );
}
