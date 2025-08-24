import { DataSourcePlugin } from '@grafana/data';
import { DataSource } from './datasource';
import { ConfigEditor } from './components/ConfigEditor';
import { QueryEditor } from './components/QueryEditor';
import { CatalystQuery, CatalystJsonData } from './types';

export const plugin = new DataSourcePlugin<DataSource, CatalystQuery, CatalystJsonData>(DataSource)
  .setConfigEditor(ConfigEditor)
  .setQueryEditor(QueryEditor);
  .setVariableQueryEditor(VariableQueryEditor);