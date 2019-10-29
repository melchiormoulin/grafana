import React, { Component } from 'react';
import { DerivedLogField, LinkModel, LogRowModel, TimeZone } from '@grafana/data';
import { cx } from 'emotion';
import { DataQueryResponse } from '../../index';
import {
  LogRowContextRows,
  LogRowContextQueryErrors,
  HasMoreContextRows,
  LogRowContextProvider,
} from './LogRowContextProvider';
import { LogLabels } from './LogLabels';
import { Themeable } from '../../types/theme';
import { withTheme } from '../../themes/index';
import { getLogRowStyles } from './getLogRowStyles';
import { LogRowMessage } from './LogRowMessage';

interface Props extends Themeable {
  highlighterExpressions?: string[];
  row: LogRowModel;
  showDuplicates: boolean;
  showLabels: boolean;
  showTime: boolean;
  timeZone: TimeZone;
  getRows: () => LogRowModel[];
  onClickLabel?: (label: string, value: string) => void;
  onContextClick?: () => void;
  getRowContext: (row: LogRowModel, options?: any) => Promise<DataQueryResponse>;
  getDerivedFields: (row: LogRowModel) => DerivedLogField[];
}

interface State {
  showContext: boolean;
  showDetails: boolean;
  loadingDerivedFields: boolean;
  linkModels: Array<LinkModel<any>> | null;
}

/**
 * Renders a log line.
 *
 * When user hovers over it for a certain time, it lazily parses the log line.
 * Once a parser is found, it will determine fields, that will be highlighted.
 * When the user requests stats for a field, they will be calculated and rendered below the row.
 */
class UnThemedLogRow extends Component<Props, State> {
  mounted = true;

  state: State = {
    showContext: false,
    showDetails: false,
    loadingDerivedFields: false,
    linkModels: null,
  };

  componentWillUnmount(): void {
    this.mounted = false;
  }

  toggleContext = () => {
    this.setState(state => {
      return {
        showContext: !state.showContext,
      };
    });
  };

  renderLogRow(
    context?: LogRowContextRows,
    errors?: LogRowContextQueryErrors,
    hasMoreContextRows?: HasMoreContextRows,
    updateLimit?: () => void
  ) {
    const {
      getRows,
      highlighterExpressions,
      onClickLabel,
      row,
      showDuplicates,
      showLabels,
      timeZone,
      showTime,
      theme,
      getDerivedFields,
    } = this.props;
    const { showContext, showDetails } = this.state;
    const style = getLogRowStyles(theme, row.logLevel);
    const showUtc = timeZone === 'utc';

    return (
      <div>
        <div
          className={cx([style.logsRow])}
          onClick={async () => {
            this.setState({
              showDetails: !showDetails,
            });
          }}
        >
          {showDuplicates && (
            <div className={cx([style.logsRowDuplicates])}>
              {row.duplicates && row.duplicates > 0 ? `${row.duplicates + 1}x` : null}
            </div>
          )}
          <div className={cx([style.logsRowLevel])} />
          {showTime && showUtc && (
            <div className={cx([style.logsRowLocalTime])} title={`Local: ${row.timeLocal} (${row.timeFromNow})`}>
              {row.timeUtc}
            </div>
          )}
          {showTime && !showUtc && (
            <div className={cx([style.logsRowLocalTime])} title={`${row.timeUtc} (${row.timeFromNow})`}>
              {row.timeLocal}
            </div>
          )}
          {showLabels && (
            <div className={cx([style.logsRowLabels])}>
              <LogLabels
                getRows={getRows}
                labels={row.uniqueLabels ? row.uniqueLabels : {}}
                onClickLabel={onClickLabel}
              />
            </div>
          )}
          <LogRowMessage
            highlighterExpressions={highlighterExpressions}
            row={row}
            getRows={getRows}
            errors={errors}
            hasMoreContextRows={hasMoreContextRows}
            updateLimit={updateLimit}
            context={context}
            showContext={showContext}
            onToggleContext={this.toggleContext}
          />
        </div>
        {showDetails && (
          <div
            style={{
              padding: 20,
            }}
          >
            {getDerivedFields(row).map(f => {
              return (
                <div>
                  {f.field} = <a href={f.href}>{f.value}</a>
                </div>
              );
            })}
          </div>
        )}
      </div>
    );
  }

  render() {
    const { showContext } = this.state;

    if (showContext) {
      return (
        <>
          <LogRowContextProvider row={this.props.row} getRowContext={this.props.getRowContext}>
            {({ result, errors, hasMoreContextRows, updateLimit }) => {
              return <>{this.renderLogRow(result, errors, hasMoreContextRows, updateLimit)}</>;
            }}
          </LogRowContextProvider>
        </>
      );
    }

    return this.renderLogRow();
  }
}

export const LogRow = withTheme(UnThemedLogRow);
LogRow.displayName = 'LogRow';
