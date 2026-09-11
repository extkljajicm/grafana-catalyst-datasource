// module.ts: Entry point for Grafana plugin registration.
// Wires up the DataSource class and all editor components for the plugin UI.
import { DataSourcePlugin } from '@grafana/data';
import { DataSource } from './datasource';
import { ConfigEditor } from './components/ConfigEditor';
import QueryEditor from './components/QueryEditor';
import type { CatalystQuery, CatalystJsonData } from './types';

// Register the plugin with Grafana, specifying:
// - DataSource: main backend logic
// - ConfigEditor: UI for datasource config
// - QueryEditor: UI for query building
// - VariableQueryEditor: UI for template variable queries
// Uses generics for type safety: <DataSourceClass, QueryModel, JsonData>
export const plugin = new DataSourcePlugin<DataSource, CatalystQuery, CatalystJsonData>(DataSource)
  .setConfigEditor(ConfigEditor)
  .setQueryEditor(QueryEditor);
