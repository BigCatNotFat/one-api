import React, { useEffect, useState } from 'react';
import {
  Card,
  Grid,
  Header,
  Segment,
  Statistic,
  Table,
  Loader,
  Dimmer,
  Label,
  Dropdown,
} from 'semantic-ui-react';
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

  useEffect(() => {
    loadAllStats();
  }, [dauDays]);

  const dayOptions = [
    { key: 7, text: '7 天', value: 7 },
    { key: 30, text: '30 天', value: 30 },
    { key: 90, text: '90 天', value: 90 },
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

