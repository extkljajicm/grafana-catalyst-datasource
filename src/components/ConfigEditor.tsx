import React, { ChangeEvent } from 'react';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { Field, Input, Button } from '@grafana/ui';
import { CatalystJsonData, CatalystSecureJsonData } from '../types';

type Props = DataSourcePluginOptionsEditorProps<CatalystJsonData, CatalystSecureJsonData>;

export const ConfigEditor: React.FC<Props> = ({ options, onOptionsChange }) => {
  const jsonData = options.jsonData || {};
  const secureJsonData = options.secureJsonData || {};
  const secureJsonFields = options.secureJsonFields || {};

  const onBaseUrlChange = (e: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      jsonData: { ...jsonData, baseUrl: e.currentTarget.value },
    });
  };

  const onApiTokenChange = (e: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      secureJsonData: { ...secureJsonData, apiToken: e.currentTarget.value },
    });
  };

  const onResetToken = () => {
    onOptionsChange({
      ...options,
      secureJsonFields: { ...secureJsonFields, apiToken: false },
      secureJsonData: { ...secureJsonData, apiToken: '' },
    });
  };

  return (
    <>
      <Field label="Catalyst Base URL" description="Example: https://dnac.example.com/dna/intent/api/v1">
        <Input
          value={jsonData.baseUrl ?? ''}
          onChange={onBaseUrlChange}
          placeholder="https://dnac.example.com/dna/intent/api/v1"
          width={60}
        />
      </Field>

      <Field
        label="API Token"
        description="Catalyst Center X-Auth-Token; stored securely."
      >
        {secureJsonFields.apiToken ? (
          <div className="gf-form">
            <Button variant="secondary" onClick={onResetToken}>
              Reset saved token
            </Button>
          </div>
        ) : (
          <Input
            value={secureJsonData.apiToken ?? ''}
            onChange={onApiTokenChange}
            type="password"
            placeholder="Paste token"
            width={60}
          />
        )}
      </Field>
    </>
  );
};
