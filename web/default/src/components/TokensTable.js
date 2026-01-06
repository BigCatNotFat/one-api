import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Button,
  Dropdown,
  Form,
  Label,
  Pagination,
  Popup,
  Table,
  Checkbox,
} from 'semantic-ui-react';
import { Link } from 'react-router-dom';
import {
  API,
  copy,
  showError,
  showSuccess,
  showWarning,
  timestamp2string,
} from '../helpers';

import { ITEMS_PER_PAGE } from '../constants';
import { renderQuota } from '../helpers/render';

function renderTimestamp(timestamp) {
  return <>{timestamp2string(timestamp)}</>;
}

function renderStatus(status, t) {
  switch (status) {
    case 1:
      return (
        <Label basic color='green'>
          {t('token.table.status_enabled')}
        </Label>
      );
    case 2:
      return (
        <Label basic color='red'>
          {t('token.table.status_disabled')}
        </Label>
      );
    case 3:
      return (
        <Label basic color='yellow'>
          {t('token.table.status_expired')}
        </Label>
      );
    case 4:
      return (
        <Label basic color='grey'>
          {t('token.table.status_depleted')}
        </Label>
      );
    default:
      return (
        <Label basic color='black'>
          {t('token.table.status_unknown')}
        </Label>
      );
  }
}

const TokensTable = () => {
  const { t } = useTranslation();

  const COPY_OPTIONS = [
    { key: 'raw', text: t('token.copy_options.raw'), value: '' },
    { key: 'next', text: t('token.copy_options.next'), value: 'next' },
    { key: 'ama', text: t('token.copy_options.ama'), value: 'ama' },
    { key: 'opencat', text: t('token.copy_options.opencat'), value: 'opencat' },
    { key: 'lobe', text: t('token.copy_options.lobe'), value: 'lobechat' },
  ];

  const OPEN_LINK_OPTIONS = [
    { key: 'next', text: t('token.copy_options.next'), value: 'next' },
    { key: 'ama', text: t('token.copy_options.ama'), value: 'ama' },
    { key: 'opencat', text: t('token.copy_options.opencat'), value: 'opencat' },
    { key: 'lobe', text: t('token.copy_options.lobe'), value: 'lobechat' },
  ];

  const [tokens, setTokens] = useState([]);
  const [loading, setLoading] = useState(true);
  const [activePage, setActivePage] = useState(1);
  const [searchKeyword, setSearchKeyword] = useState('');
  const [searching, setSearching] = useState(false);
  const [showTopUpModal, setShowTopUpModal] = useState(false);
  const [targetTokenIdx, setTargetTokenIdx] = useState(0);
  const [orderBy, setOrderBy] = useState('');
  const [selectedTokens, setSelectedTokens] = useState([]);

  const loadTokens = async (startIdx) => {
    const res = await API.get(`/api/token/?p=${startIdx}&order=${orderBy}`);
    const { success, message, data } = res.data;
    if (success) {
      if (startIdx === 0) {
        setTokens(data);
      } else {
        let newTokens = [...tokens];
        newTokens.splice(startIdx * ITEMS_PER_PAGE, data.length, ...data);
        setTokens(newTokens);
      }
    } else {
      showError(message);
    }
    setLoading(false);
  };

  const onPaginationChange = (e, { activePage }) => {
    (async () => {
      if (activePage === Math.ceil(tokens.length / ITEMS_PER_PAGE) + 1) {
        // In this case we have to load more data and then append them.
        await loadTokens(activePage - 1, orderBy);
      }
      setActivePage(activePage);
    })();
  };

  const refresh = async () => {
    setLoading(true);
    await loadTokens(activePage - 1);
  };

  const onCopy = async (type, key) => {
    let status = localStorage.getItem('status');
    let serverAddress = '';
    if (status) {
      status = JSON.parse(status);
      serverAddress = status.server_address;
    }
    if (serverAddress === '') {
      serverAddress = window.location.origin;
    }
    let encodedServerAddress = encodeURIComponent(serverAddress);
    const nextLink = localStorage.getItem('chat_link');
    let nextUrl;

    if (nextLink) {
      nextUrl =
        nextLink + `/#/?settings={"key":"sk-${key}","url":"${serverAddress}"}`;
    } else {
      nextUrl = `https://app.nextchat.dev/#/?settings={"key":"sk-${key}","url":"${serverAddress}"}`;
    }

    let url;
    switch (type) {
      case 'ama':
        url = `ama://set-api-key?server=${encodedServerAddress}&key=sk-${key}`;
        break;
      case 'opencat':
        url = `opencat://team/join?domain=${encodedServerAddress}&token=sk-${key}`;
        break;
      case 'next':
        url = nextUrl;
        break;
      case 'lobechat':
        url =
          nextLink +
          `/?settings={"keyVaults":{"openai":{"apiKey":"sk-${key}","baseURL":"${serverAddress}/v1"}}}`;
        break;
      default:
        url = `sk-${key}`;
    }
    if (await copy(url)) {
      showSuccess(t('token.messages.copy_success'));
    } else {
      showWarning(t('token.messages.copy_failed'));
      setSearchKeyword(url);
    }
  };

  const onOpenLink = async (type, key) => {
    let status = localStorage.getItem('status');
    let serverAddress = '';
    if (status) {
      status = JSON.parse(status);
      serverAddress = status.server_address;
    }
    if (serverAddress === '') {
      serverAddress = window.location.origin;
    }
    let encodedServerAddress = encodeURIComponent(serverAddress);
    const chatLink = localStorage.getItem('chat_link');
    let defaultUrl;

    if (chatLink) {
      defaultUrl =
        chatLink + `/#/?settings={"key":"sk-${key}","url":"${serverAddress}"}`;
    } else {
      defaultUrl = `https://app.nextchat.dev/#/?settings={"key":"sk-${key}","url":"${serverAddress}"}`;
    }
    let url;
    switch (type) {
      case 'ama':
        url = `ama://set-api-key?server=${encodedServerAddress}&key=sk-${key}`;
        break;

      case 'opencat':
        url = `opencat://team/join?domain=${encodedServerAddress}&token=sk-${key}`;
        break;

      case 'lobechat':
        url =
          chatLink +
          `/?settings={"keyVaults":{"openai":{"apiKey":"sk-${key}","baseURL":"${serverAddress}/v1"}}}`;
        break;

      default:
        url = defaultUrl;
    }

    window.open(url, '_blank');
  };

  useEffect(() => {
    loadTokens(0, orderBy)
      .then()
      .catch((reason) => {
        showError(reason);
      });
  }, [orderBy]);

  const manageToken = async (id, action, idx) => {
    let data = { id };
    let res;
    switch (action) {
      case 'delete':
        res = await API.delete(`/api/token/${id}/`);
        break;
      case 'enable':
        data.status = 1;
        res = await API.put('/api/token/?status_only=true', data);
        break;
      case 'disable':
        data.status = 2;
        res = await API.put('/api/token/?status_only=true', data);
        break;
    }
    const { success, message } = res.data;
    if (success) {
      showSuccess(t('token.messages.operation_success'));
      let token = res.data.data;
      let newTokens = [...tokens];
      let realIdx = (activePage - 1) * ITEMS_PER_PAGE + idx;
      if (action === 'delete') {
        newTokens[realIdx].deleted = true;
      } else {
        newTokens[realIdx].status = token.status;
      }
      setTokens(newTokens);
    } else {
      showError(message);
    }
  };

  const searchTokens = async () => {
    if (searchKeyword === '') {
      // if keyword is blank, load files instead.
      await loadTokens(0);
      setActivePage(1);
      setOrderBy('');
      return;
    }
    setSearching(true);
    const res = await API.get(`/api/token/search?keyword=${searchKeyword}`);
    const { success, message, data } = res.data;
    if (success) {
      setTokens(data);
      setActivePage(1);
    } else {
      showError(message);
    }
    setSearching(false);
  };

  const handleKeywordChange = async (e, { value }) => {
    setSearchKeyword(value.trim());
  };

  const sortToken = (key) => {
    if (tokens.length === 0) return;
    setLoading(true);
    let sortedTokens = [...tokens];
    sortedTokens.sort((a, b) => {
      if (!isNaN(a[key])) {
        // If the value is numeric, subtract to sort
        return a[key] - b[key];
      } else {
        // If the value is not numeric, sort as strings
        return ('' + a[key]).localeCompare(b[key]);
      }
    });
    if (sortedTokens[0].id === tokens[0].id) {
      sortedTokens.reverse();
    }
    setTokens(sortedTokens);
    setLoading(false);
  };

  const handleOrderByChange = (e, { value }) => {
    setOrderBy(value);
    setActivePage(1);
  };

  const handleCheck = (id) => {
    if (selectedTokens.includes(id)) {
      setSelectedTokens(selectedTokens.filter((t) => t !== id));
    } else {
      setSelectedTokens([...selectedTokens, id]);
    }
  };

  const handleCheckAll = () => {
    const currentPageTokens = tokens
      .slice((activePage - 1) * ITEMS_PER_PAGE, activePage * ITEMS_PER_PAGE)
      .filter((t) => !t.deleted);
    const currentPageIds = currentPageTokens.map((t) => t.id);
    const allSelected = currentPageIds.every((id) =>
      selectedTokens.includes(id)
    );

    if (allSelected) {
      setSelectedTokens(
        selectedTokens.filter((id) => !currentPageIds.includes(id))
      );
    } else {
      const newSelected = new Set([...selectedTokens, ...currentPageIds]);
      setSelectedTokens(Array.from(newSelected));
    }
  };

  const bulkDelete = async () => {
    if (selectedTokens.length === 0) return;
    if (!window.confirm(`确定要删除选中的 ${selectedTokens.length} 个令牌吗？`))
      return;
    setLoading(true);
    let successCount = 0;
    for (const id of selectedTokens) {
      try {
        await API.delete(`/api/token/${id}/`);
        successCount++;
      } catch (e) {
        console.error(e);
      }
    }
    showSuccess(`成功删除 ${successCount} 个令牌`);
    setSelectedTokens([]);
    await refresh();
  };

  const bulkExport = async () => {
    setLoading(true);
    try {
      let allData = [];
      let p = 0;
      while (true) {
        const res = await API.get(`/api/token/?p=${p}`);
        const { success, data } = res.data;
        if (success && data && data.length > 0) {
          allData.push(...data);
          if (data.length < ITEMS_PER_PAGE) break;
          p++;
        } else {
          break;
        }
      }

      const jsonString = JSON.stringify(allData, null, 2);
      const blob = new Blob([jsonString], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `tokens_${new Date().toISOString().slice(0, 10)}.json`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      showSuccess('导出成功');
    } catch (e) {
      showError('导出失败: ' + e.message);
    }
    setLoading(false);
  };

  return (
    <>
      <Form onSubmit={searchTokens}>
        <Form.Input
          icon='search'
          fluid
          iconPosition='left'
          placeholder={t('token.search')}
          value={searchKeyword}
          loading={searching}
          onChange={handleKeywordChange}
        />
      </Form>

      <Table basic={'very'} compact size='small'>
        <Table.Header>
          <Table.Row>
            <Table.HeaderCell>
              <Checkbox
                onChange={handleCheckAll}
                checked={
                  tokens
                    .slice(
                      (activePage - 1) * ITEMS_PER_PAGE,
                      activePage * ITEMS_PER_PAGE
                    )
                    .filter((t) => !t.deleted).length > 0 &&
                  tokens
                    .slice(
                      (activePage - 1) * ITEMS_PER_PAGE,
                      activePage * ITEMS_PER_PAGE
                    )
                    .filter((t) => !t.deleted)
                    .every((t) => selectedTokens.includes(t.id))
                }
              />
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortToken('name');
              }}
            >
              {t('token.table.name')}
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortToken('status');
              }}
            >
              {t('token.table.status')}
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortToken('used_quota');
              }}
            >
              {t('token.table.used_quota')}
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortToken('remain_quota');
              }}
            >
              {t('token.table.remain_quota')}
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortToken('created_time');
              }}
            >
              {t('token.table.created_time')}
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortToken('expired_time');
              }}
            >
              {t('token.table.expired_time')}
            </Table.HeaderCell>
            <Table.HeaderCell>{t('token.table.actions')}</Table.HeaderCell>
          </Table.Row>
        </Table.Header>

        <Table.Body>
          {tokens
            .slice(
              (activePage - 1) * ITEMS_PER_PAGE,
              activePage * ITEMS_PER_PAGE
            )
            .map((token, idx) => {
              if (token.deleted) return <></>;

              const copyOptionsWithHandlers = COPY_OPTIONS.map((option) => ({
                ...option,
                onClick: async () => {
                  await onCopy(option.value, token.key);
                },
              }));

              const openLinkOptionsWithHandlers = OPEN_LINK_OPTIONS.map(
                (option) => ({
                  ...option,
                  onClick: async () => {
                    await onOpenLink(option.value, token.key);
                  },
                })
              );

              return (
                <Table.Row key={token.id}>
                  <Table.Cell>
                    <Checkbox
                      onChange={() => handleCheck(token.id)}
                      checked={selectedTokens.includes(token.id)}
                    />
                  </Table.Cell>
                  <Table.Cell>
                    {token.name ? token.name : t('token.table.no_name')}
                  </Table.Cell>
                  <Table.Cell>{renderStatus(token.status, t)}</Table.Cell>
                  <Table.Cell>{renderQuota(token.used_quota, t)}</Table.Cell>
                  <Table.Cell>
                    {token.unlimited_quota
                      ? t('token.table.unlimited')
                      : renderQuota(token.remain_quota, t, 2)}
                  </Table.Cell>
                  <Table.Cell>{renderTimestamp(token.created_time)}</Table.Cell>
                  <Table.Cell>
                    {token.expired_time === -1
                      ? t('token.table.never_expire')
                      : renderTimestamp(token.expired_time)}
                  </Table.Cell>
                  <Table.Cell>
                    <div>
                      <Button.Group color='green' size={'tiny'}>
                        <Button
                          size={'tiny'}
                          positive
                          onClick={async () => await onCopy('', token.key)}
                        >
                          {t('token.buttons.copy')}
                        </Button>
                        <Dropdown
                          className='button icon'
                          floating
                          options={copyOptionsWithHandlers}
                          trigger={<></>}
                        />
                      </Button.Group>{' '}
                      <Button.Group color='olive' size={'tiny'}>
                        <Button
                          size={'tiny'}
                          positive
                          onClick={() => onOpenLink('', token.key)}
                        >
                          {t('token.buttons.chat')}
                        </Button>
                        <Dropdown
                          className='button icon'
                          floating
                          options={openLinkOptionsWithHandlers}
                          trigger={<></>}
                        />
                      </Button.Group>{' '}
                      <Popup
                        trigger={
                          <Button size='mini' negative>
                            {t('token.buttons.delete')}
                          </Button>
                        }
                        on='click'
                        flowing
                        hoverable
                      >
                        <Button
                          size={'tiny'}
                          negative
                          onClick={() => {
                            manageToken(token.id, 'delete', idx);
                          }}
                        >
                          {t('token.buttons.confirm_delete')} {token.name}
                        </Button>
                      </Popup>
                      <Button
                        size={'tiny'}
                        onClick={() => {
                          manageToken(
                            token.id,
                            token.status === 1 ? 'disable' : 'enable',
                            idx
                          );
                        }}
                      >
                        {token.status === 1
                          ? t('token.buttons.disable')
                          : t('token.buttons.enable')}
                      </Button>
                      <Button
                        size={'tiny'}
                        as={Link}
                        to={'/token/edit/' + token.id}
                      >
                        {t('token.buttons.edit')}
                      </Button>
                    </div>
                  </Table.Cell>
                </Table.Row>
              );
            })}
        </Table.Body>

        <Table.Footer>
          <Table.Row>
            <Table.HeaderCell colSpan='8'>
              <Button size='small' as={Link} to='/token/add' loading={loading}>
                {t('token.buttons.add')}
              </Button>
              <Button size='small' onClick={bulkExport} loading={loading}>
                {t('token.buttons.export') || '导出批量令牌'}
              </Button>
              {selectedTokens.length > 0 && (
                <Button size='small' onClick={bulkDelete} loading={loading}>
                  {t('token.buttons.delete_selected') || '删除选中令牌'}
                </Button>
              )}
              <Button size='small' onClick={refresh} loading={loading}>
                {t('token.buttons.refresh')}
              </Button>
              <Dropdown
                placeholder={t('token.sort.placeholder')}
                selection
                options={[
                  { key: '', text: t('token.sort.default'), value: '' },
                  {
                    key: 'remain_quota',
                    text: t('token.sort.by_remain'),
                    value: 'remain_quota',
                  },
                  {
                    key: 'used_quota',
                    text: t('token.sort.by_used'),
                    value: 'used_quota',
                  },
                ]}
                value={orderBy}
                onChange={handleOrderByChange}
                style={{ marginLeft: '10px' }}
              />
              <Pagination
                floated='right'
                activePage={activePage}
                onPageChange={onPaginationChange}
                size='small'
                siblingRange={1}
                totalPages={
                  Math.ceil(tokens.length / ITEMS_PER_PAGE) +
                  (tokens.length % ITEMS_PER_PAGE === 0 ? 1 : 0)
                }
              />
            </Table.HeaderCell>
          </Table.Row>
        </Table.Footer>
      </Table>
    </>
  );
};

export default TokensTable;
