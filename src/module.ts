import { DataSourcePlugin } from '@grafana/data';
import { DataSource } from './datasource';
import { ConfigEditor } from './components/ConfigEditor';
import QueryEditor from './components/QueryEditor';
import VariableQueryEditor from './components/VariableQueryEditor';
import type { CatalystQuery, CatalystJsonData } from './types';

// Wire up datasource + editors.
// Generics: <DataSourceClass, QueryModel, JsonData>
export const plugin = new DataSourcePlugin<DataSource, CatalystQuery, CatalystJsonData>(DataSource)
  .setConfigEditor(ConfigEditor)
  .setQueryEditor(QueryEditor)
  .setVariableQueryEditor(VariableQueryEditor);
