import React, { useMemo } from 'react';
import { InlineField, InlineFieldRow, Input, Select, TextArea } from '@grafana/ui';
import type { SelectableValue, VariableQueryEditorProps } from '@grafana/data';

/**
 * Shape we store inside the variable's "Query options" UI.
 * You can tweak/extend this to match your datasource's needs.
 */
export type CatalystVariableQuery = {
  mode?: 'raw' | 'labelValues';
  // raw
  query?: string;
  // labelValues helper
  metric?: string;
  label?: string;
};

const MODE_OPTIONS: Array<SelectableValue<CatalystVariableQuery['mode']>> = [
  { label: 'Raw (passed to metricFindQuery)', value: 'raw' },
  { label: 'Label values helper', value: 'labelValues' },
];

/**
 * Variable Query Editor
 * - "raw": whatever you type is sent as-is to metricFindQuery(query)
 * - "labelValues": builds a definition like `label_values(<metric>, <label>)`
 *
 * Make sure this editor is wired in module.ts:
 *   plugin.setVariableQueryEditor(VariableQueryEditor)
 */
export function VariableQueryEditor(
  props: VariableQueryEditorProps<any, CatalystVariableQuery>
): JSX.Element {
  const { query, onChange } = props;
  const value: CatalystVariableQuery = query ?? { mode: 'raw', query: '' };
  const mode: CatalystVariableQuery['mode'] = value.mode ?? 'raw';

  const definition = useMemo(() => {
    if (mode === 'labelValues') {
      const m = value.metric?.trim() ?? '';
      const l = value.label?.trim() ?? '';
      if (m && l) {
        return `label_values(${m}, ${l})`;
      }
      if (m) {
        return `label_values(${m}, <label>)`;
      }
      return 'label_values(<metric>, <label>)';
    }
    // raw
    return (value.query ?? '').trim();
  }, [mode, value.metric, value.label, value.query]);

  const update = (patch: Partial<CatalystVariableQuery>) => {
    const next = { ...value, ...patch };
    // Recompute a human-friendly definition string Grafana displays under the variable.
    const def =
      next.mode === 'labelValues'
        ? (() => {
            const m = next.metric?.trim() ?? '';
            const l = next.label?.trim() ?? '';
            if (m && l) {
              return `label_values(${m}, ${l})`;
            }
            if (m) {
              return `label_values(${m}, <label>)`;
            }
            return 'label_values(<metric>, <label>)';
          })()
        : (next.query ?? '').trim();

    onChange(next, def);
  };

  return (
    <div className="gf-form-group">
      <InlineFieldRow>
        <InlineField label="Mode" grow={true}>
          <Select
            options={MODE_OPTIONS}
            value={MODE_OPTIONS.find((o) => o.value === mode) ?? MODE_OPTIONS[0]}
            onChange={(v) => update({ mode: v.value ?? 'raw' })}
            aria-label="Variable query mode"
          />
        </InlineField>
      </InlineFieldRow>

      {mode === 'raw' && (
        <>
          <InlineFieldRow>
            <InlineField label="Query" grow={true} tooltip="This value is passed directly to metricFindQuery(query)">
              <TextArea
                value={value.query ?? ''}
                onChange={(e) => update({ query: e.currentTarget.value })}
                placeholder="e.g. services | jsonpath '$[*].name'"
                rows={4}
              />
            </InlineField>
          </InlineFieldRow>
        </>
      )}

      {mode === 'labelValues' && (
        <>
          <InlineFieldRow>
            <InlineField label="Metric" grow={true} tooltip="The metric/series to inspect for label values">
              <Input
                value={value.metric ?? ''}
                onChange={(e) => update({ metric: e.currentTarget.value })}
                placeholder="e.g. http_requests_total"
              />
            </InlineField>
          </InlineFieldRow>
          <InlineFieldRow>
            <InlineField label="Label" grow={true} tooltip="The label key to return values for">
              <Input
                value={value.label ?? ''}
                onChange={(e) => update({ label: e.currentTarget.value })}
                placeholder="e.g. instance"
              />
            </InlineField>
          </InlineFieldRow>
        </>
      )}

      <InlineFieldRow>
        <InlineField label="Definition" grow={true} tooltip="Preview of the definition Grafana shows for this variable">
          <Input value={definition} readOnly />
        </InlineField>
      </InlineFieldRow>
    </div>
  );
}

export default VariableQueryEditor;
