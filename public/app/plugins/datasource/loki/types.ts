import { DataQuery, DataSourceJsonData } from '@grafana/ui';
import { Labels } from '@grafana/data';

export interface LokiQuery extends DataQuery {
  expr: string;
  liveStreaming?: boolean;
  query?: string;
  regexp?: string;
}

export interface LokiOptions extends DataSourceJsonData {
  maxLines?: string;
  derivedFields?: DerivedFieldConfig[];
}

export interface LokiResponse {
  streams: LokiLogsStream[];
}

export interface LokiLogsStream {
  labels: string;
  entries: LokiLogsStreamEntry[];
  search?: string;
  parsedLabels?: Labels;
  uniqueLabels?: Labels;
}

export interface LokiLogsStreamEntry {
  line: string;
  ts: string;
  // Legacy, was renamed to ts
  timestamp?: string;
}

export interface LokiExpression {
  regexp: string;
  query: string;
}

export type DerivedFieldConfig = {
  matcherRegex: string;
  template: string;
  label: string;
};

export type LokiRow = {
  id: string;
  line: string;
  ts: string;
};
