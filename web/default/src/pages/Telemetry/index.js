import React, { useEffect, useState, useCallback } from 'react';
import {
  Grid,
  Header,
  Segment,
  Statistic,
  Table,
  Loader,
  Dimmer,
  Label,
  Dropdown,
  Button,
} from 'semantic-ui-react';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts';
import { API, showError, isAdmin } from '../../helpers';

const Telemetry = () => {
  const [loading, setLoading] = useState(true);
  const [overview, setOverview] = useState(null);
  const [chatStats, setChatStats] = useState([]);
  const [toolStats, setToolStats] = useState([]);
  const [dauStats, setDauStats] = useState([]);
  const [textActionStats, setTextActionStats] = useState([]);
  const [toolApprovalStats, setToolApprovalStats] = useState([]);
  const [dauDays, setDauDays] = useState(30);
  const [timelineStats, setTimelineStats] = useState([]);
  const [timelineHours, setTimelineHours] = useState(24);
  const [timelineLoading, setTimelineLoading] = useState(false);
  const [lastTimelineUpdate, setLastTimelineUpdate] = useState(null);

  const loadAllStats = async () => {
    if (!isAdmin()) {
      showError('只有管理员可以查看遥测统计');
      return;
    }
    setLoading(true);
    try {
      const res = await API.get('/api/telemetry/stats/all', {
        params: { days: dauDays }
      });
      const { success, message, data } = res.data;
      if (success) {
        setOverview(data.overview || {});
        setChatStats(data.chat || []);
        setToolStats(data.tools || []);
        setDauStats(data.dau || []);
        setTextActionStats(data.textActions || []);
        setToolApprovalStats(data.toolApprovals || []);
      } else {
        showError(message);
      }
    } catch (error) {
      showError(error.message);
    }
    setLoading(false);
  };

  const loadTimelineStats = useCallback(async () => {
    if (!isAdmin()) return;
    setTimelineLoading(true);
    try {
      const res = await API.get('/api/telemetry/stats/timeline', {
        params: { hours: timelineHours }
      });
      const { success, message, data } = res.data;
      if (success) {
        setTimelineStats(data || []);
        setLastTimelineUpdate(new Date());
      } else {
        showError(message);
      }
    } catch (error) {
      showError(error.message);
    }
    setTimelineLoading(false);
  }, [timelineHours]);

  useEffect(() => {
    loadAllStats();
  }, [dauDays]);

  useEffect(() => {
    loadTimelineStats();
    // 每30分钟自动刷新时间轴数据
    const interval = setInterval(loadTimelineStats, 30 * 60 * 1000);
    return () => clearInterval(interval);
  }, [loadTimelineStats]);

  const dayOptions = [
    { key: 7, text: '7 天', value: 7 },
    { key: 30, text: '30 天', value: 30 },
    { key: 90, text: '90 天', value: 90 },
  ];

  const hoursOptions = [
    { key: 6, text: '6 小时', value: 6 },
    { key: 12, text: '12 小时', value: 12 },
    { key: 24, text: '24 小时', value: 24 },
    { key: 48, text: '48 小时', value: 48 },
    { key: 168, text: '7 天', value: 168 },
  ];

  // 计算时间轴统计摘要
  const timelineSummary = {
    totalUsers: timelineStats.reduce((sum, d) => sum + (d.active_users || 0), 0),
    totalEvents: timelineStats.reduce((sum, d) => sum + (d.event_count || 0), 0),
    totalChats: timelineStats.reduce((sum, d) => sum + (d.chat_count || 0), 0),
    totalTools: timelineStats.reduce((sum, d) => sum + (d.tool_success_count || 0) + (d.tool_failed_count || 0), 0),
  };

  // 处理时间轴数据用于图表
  const chartData = timelineStats.map(d => ({
    time: d.time_slot.split(' ')[1] || d.time_slot,
    fullTime: d.time_slot,
    activeUsers: d.active_users || 0,
    eventCount: d.event_count || 0,
    chatCount: d.chat_count || 0,
    toolSuccess: d.tool_success_count || 0,
    toolFailed: d.tool_failed_count || 0,
    avgUsageScore: Math.round((d.avg_usage_score || 0) * 100) / 100,
  }));

  // 各指标的图表配置
  const singleChartConfigs = [
    { dataKey: 'activeUsers', name: '活跃用户', color: '#2185d0', total: timelineSummary.totalUsers },
    { dataKey: 'eventCount', name: '事件数', color: '#21ba45', total: timelineSummary.totalEvents },
    { dataKey: 'chatCount', name: '聊天次数', color: '#f2711c', total: timelineSummary.totalChats },
    { dataKey: 'toolSuccess', name: '工具成功', color: '#00b5ad', total: timelineStats.reduce((sum, d) => sum + (d.tool_success_count || 0), 0) },
    { dataKey: 'toolFailed', name: '工具失败', color: '#db2828', total: timelineStats.reduce((sum, d) => sum + (d.tool_failed_count || 0), 0) },
    { dataKey: 'avgUsageScore', name: '平均使用度', color: '#a333c8', total: (timelineStats.reduce((sum, d) => sum + (d.avg_usage_score || 0), 0) / (timelineStats.length || 1)).toFixed(2), isAvg: true },
  ];

  if (!isAdmin()) {
    return (
      <Segment placeholder>
        <Header icon>
          <i className="lock icon"></i>
          只有管理员可以查看遥测统计
        </Header>
      </Segment>
    );
  }

  return (
    <>
      <Segment>
        <Header as="h3">
          <i className="chart bar icon"></i>
          遥测统计 (PacificOceanAI)
        </Header>
      </Segment>

      {loading ? (
        <Segment>
          <Dimmer active inverted>
            <Loader>加载中...</Loader>
          </Dimmer>
          <div style={{ height: 200 }}></div>
        </Segment>
      ) : (
        <>
          {/* 概览统计 */}
          <Segment>
            <Header as="h4">概览</Header>
            <Statistic.Group widths="four">
              <Statistic>
                <Statistic.Value>{overview?.total_users || 0}</Statistic.Value>
                <Statistic.Label>总用户数</Statistic.Label>
              </Statistic>
              <Statistic color="green">
                <Statistic.Value>{overview?.today_users || 0}</Statistic.Value>
                <Statistic.Label>今日活跃</Statistic.Label>
              </Statistic>
              <Statistic color="blue">
                <Statistic.Value>{overview?.total_sessions || 0}</Statistic.Value>
                <Statistic.Label>总会话数</Statistic.Label>
              </Statistic>
              <Statistic color="orange">
                <Statistic.Value>{overview?.total_events || 0}</Statistic.Value>
                <Statistic.Label>总事件数</Statistic.Label>
              </Statistic>
            </Statistic.Group>
          </Segment>

          {/* 实时时间轴统计 */}
          <Segment loading={timelineLoading}>
            <Header as="h4">
              <i className="clock icon"></i>
              实时数据监控（每30分钟更新）
              <Dropdown
                selection
                compact
                options={hoursOptions}
                value={timelineHours}
                onChange={(e, { value }) => setTimelineHours(value)}
                style={{ marginLeft: 20 }}
              />
              <Button
                icon="refresh"
                size="tiny"
                onClick={loadTimelineStats}
                loading={timelineLoading}
                style={{ marginLeft: 10 }}
              />
              {lastTimelineUpdate && (
                <span style={{ fontSize: '12px', color: '#999', marginLeft: 10 }}>
                  上次更新: {lastTimelineUpdate.toLocaleTimeString()}
                </span>
              )}
            </Header>

            {/* 时间段快速统计摘要 */}
            <Statistic.Group widths="four" size="tiny" style={{ marginBottom: 20 }}>
              <Statistic>
                <Statistic.Value>{timelineSummary.totalUsers}</Statistic.Value>
                <Statistic.Label>用户活跃(累计)</Statistic.Label>
              </Statistic>
              <Statistic color="blue">
                <Statistic.Value>{timelineSummary.totalEvents}</Statistic.Value>
                <Statistic.Label>事件数(累计)</Statistic.Label>
              </Statistic>
              <Statistic color="green">
                <Statistic.Value>{timelineSummary.totalChats}</Statistic.Value>
                <Statistic.Label>聊天次数(累计)</Statistic.Label>
              </Statistic>
              <Statistic color="teal">
                <Statistic.Value>{timelineSummary.totalTools}</Statistic.Value>
                <Statistic.Label>工具调用(累计)</Statistic.Label>
              </Statistic>
            </Statistic.Group>

            {chartData.length > 0 ? (
              <Grid columns={3} stackable doubling>
                {singleChartConfigs.map((config, index) => (
                  <Grid.Column key={index}>
                    <div style={{ 
                      border: '1px solid #e0e0e0', 
                      borderRadius: 8, 
                      padding: '12px',
                      backgroundColor: '#fafafa',
                      marginBottom: 8
                    }}>
                      <div style={{ 
                        display: 'flex', 
                        justifyContent: 'space-between', 
                        alignItems: 'center',
                        marginBottom: 8
                      }}>
                        <span style={{ fontWeight: 600, color: config.color }}>{config.name}</span>
                        <span style={{ 
                          backgroundColor: config.color + '20',
                          color: config.color,
                          padding: '2px 8px',
                          borderRadius: 4,
                          fontSize: 12,
                          fontWeight: 600
                        }}>
                          {config.isAvg ? `均值: ${config.total}` : `总计: ${config.total}`}
                        </span>
                      </div>
                      <ResponsiveContainer width="100%" height={150}>
                        <LineChart data={chartData} margin={{ top: 5, right: 5, left: -20, bottom: 5 }}>
                          <CartesianGrid strokeDasharray="3 3" stroke="#eee" />
                          <XAxis 
                            dataKey="time" 
                            tick={{ fontSize: 9 }} 
                            interval={Math.floor(chartData.length / 4)}
                          />
                          <YAxis tick={{ fontSize: 9 }} />
                          <Tooltip
                            labelFormatter={(label, payload) => {
                              if (payload && payload.length > 0) {
                                return payload[0].payload.fullTime;
                              }
                              return label;
                            }}
                            contentStyle={{
                              backgroundColor: 'rgba(255, 255, 255, 0.95)',
                              border: '1px solid #ddd',
                              borderRadius: 4,
                              fontSize: 12
                            }}
                          />
                          <Line
                            type="monotone"
                            dataKey={config.dataKey}
                            name={config.name}
                            stroke={config.color}
                            strokeWidth={2}
                            dot={false}
                            activeDot={{ r: 3 }}
                          />
                        </LineChart>
                      </ResponsiveContainer>
                    </div>
                  </Grid.Column>
                ))}
              </Grid>
            ) : (
              <p style={{ color: '#999', textAlign: 'center', padding: '40px 0' }}>暂无时间轴数据</p>
            )}
          </Segment>

          {/* DAU 统计 */}
          <Segment>
            <Header as="h4">
              日活跃用户 (DAU)
              <Dropdown
                selection
                compact
                options={dayOptions}
                value={dauDays}
                onChange={(e, { value }) => setDauDays(value)}
                style={{ marginLeft: 20 }}
              />
            </Header>
            {dauStats.length > 0 ? (
              <Table basic="very" celled>
                <Table.Header>
                  <Table.Row>
                    <Table.HeaderCell>日期</Table.HeaderCell>
                    <Table.HeaderCell>活跃用户数</Table.HeaderCell>
                  </Table.Row>
                </Table.Header>
                <Table.Body>
                  {dauStats.slice(-10).map((row, index) => (
                    <Table.Row key={index}>
                      <Table.Cell>{row.date}</Table.Cell>
                      <Table.Cell>
                        <Label color="blue">{row.user_count}</Label>
                      </Table.Cell>
                    </Table.Row>
                  ))}
                </Table.Body>
              </Table>
            ) : (
              <p style={{ color: '#999' }}>暂无数据</p>
            )}
          </Segment>

          <Grid columns={2} stackable>
            {/* 聊天统计 */}
            <Grid.Column>
              <Segment>
                <Header as="h4">聊天统计</Header>
                {chatStats.length > 0 ? (
                  <Table compact>
                    <Table.Header>
                      <Table.Row>
                        <Table.HeaderCell>模型</Table.HeaderCell>
                        <Table.HeaderCell>模式</Table.HeaderCell>
                        <Table.HeaderCell>次数</Table.HeaderCell>
                      </Table.Row>
                    </Table.Header>
                    <Table.Body>
                      {chatStats.map((row, index) => (
                        <Table.Row key={index}>
                          <Table.Cell>
                            <Label>{row.model_id}</Label>
                          </Table.Cell>
                          <Table.Cell>
                            <Label color={row.mode === 'agent' ? 'blue' : row.mode === 'chat' ? 'green' : 'grey'}>
                              {row.mode}
                            </Label>
                          </Table.Cell>
                          <Table.Cell>{row.total_count}</Table.Cell>
                        </Table.Row>
                      ))}
                    </Table.Body>
                  </Table>
                ) : (
                  <p style={{ color: '#999' }}>暂无数据</p>
                )}
              </Segment>
            </Grid.Column>

            {/* 工具统计 */}
            <Grid.Column>
              <Segment>
                <Header as="h4">工具调用统计</Header>
                {toolStats.length > 0 ? (
                  <Table compact>
                    <Table.Header>
                      <Table.Row>
                        <Table.HeaderCell>工具名称</Table.HeaderCell>
                        <Table.HeaderCell>成功</Table.HeaderCell>
                        <Table.HeaderCell>失败</Table.HeaderCell>
                      </Table.Row>
                    </Table.Header>
                    <Table.Body>
                      {toolStats.map((row, index) => (
                        <Table.Row key={index}>
                          <Table.Cell>
                            <Label>{row.tool_name}</Label>
                          </Table.Cell>
                          <Table.Cell>
                            <Label color="green">{row.success_count}</Label>
                          </Table.Cell>
                          <Table.Cell>
                            <Label color="red">{row.failed_count}</Label>
                          </Table.Cell>
                        </Table.Row>
                      ))}
                    </Table.Body>
                  </Table>
                ) : (
                  <p style={{ color: '#999' }}>暂无数据</p>
                )}
              </Segment>
            </Grid.Column>
          </Grid>

          <Grid columns={2} stackable>
            {/* 文本操作统计 */}
            <Grid.Column>
              <Segment>
                <Header as="h4">文本操作统计</Header>
                {textActionStats.length > 0 ? (
                  <Table compact>
                    <Table.Header>
                      <Table.Row>
                        <Table.HeaderCell>操作</Table.HeaderCell>
                        <Table.HeaderCell>触发</Table.HeaderCell>
                        <Table.HeaderCell>接受</Table.HeaderCell>
                        <Table.HeaderCell>拒绝</Table.HeaderCell>
                        <Table.HeaderCell>接受率</Table.HeaderCell>
                      </Table.Row>
                    </Table.Header>
                    <Table.Body>
                      {textActionStats.map((row, index) => {
                        const total = parseInt(row.accepted_count) + parseInt(row.rejected_count);
                        const rate = total > 0 ? ((parseInt(row.accepted_count) / total) * 100).toFixed(1) : '-';
                        return (
                          <Table.Row key={index}>
                            <Table.Cell>
                              <Label color="purple">{row.action}</Label>
                            </Table.Cell>
                            <Table.Cell>{row.used_count}</Table.Cell>
                            <Table.Cell>{row.accepted_count}</Table.Cell>
                            <Table.Cell>{row.rejected_count}</Table.Cell>
                            <Table.Cell>
                              {rate !== '-' ? (
                                <Label color={parseFloat(rate) > 70 ? 'green' : parseFloat(rate) > 40 ? 'yellow' : 'red'}>
                                  {rate}%
                                </Label>
                              ) : '-'}
                            </Table.Cell>
                          </Table.Row>
                        );
                      })}
                    </Table.Body>
                  </Table>
                ) : (
                  <p style={{ color: '#999' }}>暂无数据</p>
                )}
              </Segment>
            </Grid.Column>

            {/* 工具审批统计 */}
            <Grid.Column>
              <Segment>
                <Header as="h4">工具审批统计</Header>
                {toolApprovalStats.length > 0 ? (
                  <Table compact>
                    <Table.Header>
                      <Table.Row>
                        <Table.HeaderCell>工具</Table.HeaderCell>
                        <Table.HeaderCell>批准</Table.HeaderCell>
                        <Table.HeaderCell>拒绝</Table.HeaderCell>
                        <Table.HeaderCell>批准率</Table.HeaderCell>
                      </Table.Row>
                    </Table.Header>
                    <Table.Body>
                      {toolApprovalStats.map((row, index) => {
                        const total = parseInt(row.approved_count) + parseInt(row.rejected_count);
                        const rate = total > 0 ? ((parseInt(row.approved_count) / total) * 100).toFixed(1) : '-';
                        return (
                          <Table.Row key={index}>
                            <Table.Cell>
                              <Label color="teal">{row.tool_name}</Label>
                            </Table.Cell>
                            <Table.Cell>{row.approved_count}</Table.Cell>
                            <Table.Cell>{row.rejected_count}</Table.Cell>
                            <Table.Cell>
                              {rate !== '-' ? (
                                <Label color={parseFloat(rate) > 70 ? 'green' : parseFloat(rate) > 40 ? 'yellow' : 'red'}>
                                  {rate}%
                                </Label>
                              ) : '-'}
                            </Table.Cell>
                          </Table.Row>
                        );
                      })}
                    </Table.Body>
                  </Table>
                ) : (
                  <p style={{ color: '#999' }}>暂无数据</p>
                )}
              </Segment>
            </Grid.Column>
          </Grid>
        </>
      )}
    </>
  );
};

export default Telemetry;

