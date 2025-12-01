import React, { useState, useEffect } from 'react';
import { Button, Card, Form, Table, Message, Segment, Grid, Statistic, Pagination, Label } from 'semantic-ui-react';
import { API, showError, showSuccess, timestamp2string } from '../../helpers';

const QueryToken = () => {
  const [tokenKey, setTokenKey] = useState('');
  const [tokenData, setTokenData] = useState(null);
  const [loading, setLoading] = useState(false);
  const [activePage, setActivePage] = useState(1);
  const [pageSize] = useState(10);

  useEffect(() => {
    const prevTitle = document.title;
    document.title = 'query';
    return () => {
      document.title = prevTitle;
    };
  }, []);

  const handleQuery = async () => {
    const trimmedKey = tokenKey.trim();
    if (!trimmedKey) {
      showError('请输入API Key');
      return;
    }

    setLoading(true);
    try {
      const res = await API.get(`/api/query_token?key=${encodeURIComponent(trimmedKey)}&p=${activePage - 1}`);
      const { success, message, data } = res.data;
      if (success) {
        setTokenData(data);
        showSuccess('查询成功');
      } else {
        showError(message || '查询失败');
        setTokenData(null);
      }
    } catch (error) {
      showError('查询失败：' + error.message);
      setTokenData(null);
    } finally {
      setLoading(false);
    }
  };

  const handlePageChange = async (e, { activePage: newPage }) => {
    setActivePage(newPage);
    const trimmedKey = tokenKey.trim();
    if (!trimmedKey) return;

    setLoading(true);
    try {
      const res = await API.get(`/api/query_token?key=${encodeURIComponent(trimmedKey)}&p=${newPage - 1}`);
      const { success, data } = res.data;
      if (success) {
        setTokenData(data);
      }
    } catch (error) {
      showError('查询失败：' + error.message);
    } finally {
      setLoading(false);
    }
  };

  const renderQuota = (quota) => {
    if (quota === undefined || quota === null) return '未知';
    return '$' + (quota / 500000).toFixed(4);
  };

  const renderStatus = (status) => {
    const statusMap = {
      1: { text: '启用', color: 'green' },
      2: { text: '禁用', color: 'red' },
      3: { text: '已过期', color: 'orange' },
      4: { text: '额度已用尽', color: 'yellow' }
    };
    const statusInfo = statusMap[status] || { text: '未知', color: 'grey' };
    return <Label color={statusInfo.color}>{statusInfo.text}</Label>;
  };

  return (
    <div style={{ padding: '20px', maxWidth: '1200px', margin: '0 auto' }}>
      {/* Logo 区域 */}
      <div style={{ textAlign: 'center', marginBottom: '30px', marginTop: '20px' }}>
        <h1 style={{ 
          fontSize: '32px', 
          fontWeight: 'bold', 
          color: '#2185d0',
          marginBottom: '10px',
          letterSpacing: '1px'
        }}>
          硅基之梦
        </h1>
        <p style={{ 
          fontSize: '14px', 
          color: '#666',
          margin: 0
        }}>
          Silicon Dream
        </p>
      </div>

      <Card fluid>
        <Card.Content>
          <Card.Header style={{ fontSize: '24px', marginBottom: '20px' }}>
            API Key 查询
          </Card.Header>
          <Card.Description>
            <Form>
              <Form.Field>
                <label>请输入您的 API Key</label>
                <Form.Input
                  placeholder='sk-xxxxxxxxxx'
                  value={tokenKey}
                  onChange={(e) => setTokenKey(e.target.value)}
                  onKeyPress={(e) => {
                    if (e.key === 'Enter') {
                      handleQuery();
                    }
                  }}
                  action={{
                    content: '查询',
                    onClick: handleQuery,
                    loading: loading,
                    primary: true
                  }}
                />
              </Form.Field>
            </Form>
          </Card.Description>
        </Card.Content>
      </Card>

      {tokenData && (
        <>
          <Segment style={{ marginTop: '20px' }}>
            <Grid columns={2} stackable>
              <Grid.Row>
                <Grid.Column>
                  <Statistic.Group widths='1'>
                    <Statistic>
                      <Statistic.Value>{renderQuota(tokenData.remain_quota)}</Statistic.Value>
                      <Statistic.Label>剩余额度</Statistic.Label>
                    </Statistic>
                  </Statistic.Group>
                </Grid.Column>
                <Grid.Column>
                  <Statistic.Group widths='1'>
                    <Statistic>
                      <Statistic.Value>{renderQuota(tokenData.used_quota)}</Statistic.Value>
                      <Statistic.Label>已使用额度</Statistic.Label>
                    </Statistic>
                  </Statistic.Group>
                </Grid.Column>
              </Grid.Row>
            </Grid>
          </Segment>

          <Card fluid style={{ marginTop: '20px' }}>
            <Card.Content>
              <Card.Header>详情信息</Card.Header>
              <Card.Description>
                <Table basic='very'>
                  <Table.Body>
                    <Table.Row>
                      <Table.Cell><strong>名称</strong></Table.Cell>
                      <Table.Cell>{tokenData.name}</Table.Cell>
                    </Table.Row>
                    <Table.Row>
                      <Table.Cell><strong>状态</strong></Table.Cell>
                      <Table.Cell>{renderStatus(tokenData.status)}</Table.Cell>
                    </Table.Row>
                    <Table.Row>
                      <Table.Cell><strong>剩余额度</strong></Table.Cell>
                      <Table.Cell>
                        {tokenData.unlimited_quota ? (
                          <Label color='blue'>无限额度</Label>
                        ) : (
                          renderQuota(tokenData.remain_quota)
                        )}
                      </Table.Cell>
                    </Table.Row>
                    <Table.Row>
                      <Table.Cell><strong>已使用额度</strong></Table.Cell>
                      <Table.Cell>{renderQuota(tokenData.used_quota)}</Table.Cell>
                    </Table.Row>
                    <Table.Row>
                      <Table.Cell><strong>过期时间</strong></Table.Cell>
                      <Table.Cell>
                        {tokenData.expired_time === 0 || tokenData.expired_time === -1
                          ? '永不过期'
                          : timestamp2string(tokenData.expired_time)}
                      </Table.Cell>
                    </Table.Row>
                    <Table.Row>
                      <Table.Cell><strong>创建时间</strong></Table.Cell>
                      <Table.Cell>{timestamp2string(tokenData.created_time)}</Table.Cell>
                    </Table.Row>
                  </Table.Body>
                </Table>
              </Card.Description>
            </Card.Content>
          </Card>

          <Card fluid style={{ marginTop: '20px' }}>
            <Card.Content>
              <Card.Header>使用记录 {tokenData.time_range && <Label color='blue' size='small'>{tokenData.time_range}</Label>}</Card.Header>
              <Card.Description>
                {tokenData.logs && tokenData.logs.length > 0 ? (
                  <>
                    <Table celled striped>
                      <Table.Header>
                        <Table.Row>
                          <Table.HeaderCell>时间</Table.HeaderCell>
                          <Table.HeaderCell>模型</Table.HeaderCell>
                          <Table.HeaderCell>提示词Token</Table.HeaderCell>
                          <Table.HeaderCell>补全Token</Table.HeaderCell>
                          <Table.HeaderCell>消耗额度</Table.HeaderCell>
                          <Table.HeaderCell>请求耗时</Table.HeaderCell>
                        </Table.Row>
                      </Table.Header>
                      <Table.Body>
                        {tokenData.logs.map((log, index) => (
                          <Table.Row key={index}>
                            <Table.Cell>{timestamp2string(log.created_at)}</Table.Cell>
                            <Table.Cell>{log.model_name || '-'}</Table.Cell>
                            <Table.Cell>{log.prompt_tokens || 0}</Table.Cell>
                            <Table.Cell>{log.completion_tokens || 0}</Table.Cell>
                            <Table.Cell>{renderQuota(log.quota)}</Table.Cell>
                            <Table.Cell>{log.elapsed_time ? `${log.elapsed_time}ms` : '-'}</Table.Cell>
                          </Table.Row>
                        ))}
                      </Table.Body>
                    </Table>
                    {tokenData.logs.length === pageSize && (
                      <div style={{ textAlign: 'center', marginTop: '20px' }}>
                        <Pagination
                          activePage={activePage}
                          onPageChange={handlePageChange}
                          totalPages={Math.ceil(tokenData.logs.length / pageSize) + 1}
                          siblingRange={1}
                          firstItem={null}
                          lastItem={null}
                          prevItem={{ content: '上一页' }}
                          nextItem={{ content: '下一页' }}
                        />
                      </div>
                    )}
                  </>
                ) : (
                  <Message info>
                    <Message.Header>暂无使用记录</Message.Header>
                    <p>该令牌还没有使用记录。</p>
                  </Message>
                )}
              </Card.Description>
            </Card.Content>
          </Card>
        </>
      )}

      {!tokenData && !loading && (
        <Message info style={{ marginTop: '20px' }}>
          <Message.Header>使用说明</Message.Header>
          <Message.List>
            <Message.Item>请在上方输入框中输入您的 API Key</Message.Item>
            <Message.Item>点击"查询"按钮即可查看剩余额度和最近24小时的使用情况</Message.Item>
            <Message.Item>此页面无需登录</Message.Item>
            <Message.Item>请妥善保管您的 API Key，不要泄露给他人</Message.Item>
          </Message.List>
        </Message>
      )}
    </div>
  );
};

export default QueryToken;
