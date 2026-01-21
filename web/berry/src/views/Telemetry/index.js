import { useEffect, useState, useCallback } from 'react';
import {
  Grid,
  Typography,
  Box,
  Card,
  CardContent,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  LinearProgress,
  Chip,
  Stack,
  ToggleButton,
  ToggleButtonGroup,
  Tooltip
} from '@mui/material';
import { useTheme, styled } from '@mui/material/styles';
import Chart from 'react-apexcharts';
import { gridSpacing } from 'store/constant';
import MainCard from 'ui-component/cards/MainCard';
import { API } from 'utils/api';
import { showError } from 'utils/common';
import {
  IconUsers,
  IconMessages,
  IconCalendarStats,
  IconActivity,
  IconTool,
  IconFileText,
  IconGitBranch,
  IconTimeline,
  IconRefresh
} from '@tabler/icons-react';

// 统计卡片样式
const StatsCard = styled(Card)(({ theme, bgcolor }) => ({
  background: bgcolor || theme.palette.primary.main,
  color: '#fff',
  overflow: 'hidden',
  position: 'relative',
  '&:after': {
    content: '""',
    position: 'absolute',
    width: 210,
    height: 210,
    background: 'rgba(255,255,255,0.1)',
    borderRadius: '50%',
    top: -85,
    right: -95
  },
  '&:before': {
    content: '""',
    position: 'absolute',
    width: 210,
    height: 210,
    background: 'rgba(255,255,255,0.1)',
    borderRadius: '50%',
    top: -125,
    right: -15,
    opacity: 0.5
  }
}));

const Telemetry = () => {
  const theme = useTheme();
  const [loading, setLoading] = useState(true);
  const [overview, setOverview] = useState(null);
  const [chatStats, setChatStats] = useState([]);
  const [toolStats, setToolStats] = useState([]);
  const [dauStats, setDauStats] = useState([]);
  const [textActionStats, setTextActionStats] = useState([]);
  const [toolApprovalStats, setToolApprovalStats] = useState([]);
  const [dauDays, setDauDays] = useState('30');
  const [timelineStats, setTimelineStats] = useState([]);
  const [timelineHours, setTimelineHours] = useState('24');
  const [timelineLoading, setTimelineLoading] = useState(false);
  const [lastTimelineUpdate, setLastTimelineUpdate] = useState(null);

  const loadAllStats = async () => {
    setLoading(true);
    try {
      const res = await API.get('/api/telemetry/stats/all', {
        params: { days: parseInt(dauDays) }
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
    setTimelineLoading(true);
    try {
      const res = await API.get('/api/telemetry/stats/timeline', {
        params: { hours: parseInt(timelineHours) }
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

  // DAU 图表配置
  const dauChartOptions = {
    chart: {
      type: 'area',
      toolbar: { show: false },
      zoom: { enabled: false }
    },
    dataLabels: { enabled: false },
    stroke: { curve: 'smooth', width: 2 },
    fill: {
      type: 'gradient',
      gradient: {
        shadeIntensity: 1,
        opacityFrom: 0.5,
        opacityTo: 0.1
      }
    },
    xaxis: {
      categories: dauStats.map(d => d.date),
      labels: {
        rotate: -45,
        style: { fontSize: '10px' }
      }
    },
    yaxis: {
      labels: {
        formatter: (val) => Math.round(val)
      }
    },
    tooltip: {
      y: {
        formatter: (val) => `${val} 用户`
      }
    },
    colors: [theme.palette.primary.main]
  };

  const dauChartSeries = [{
    name: '日活用户',
    data: dauStats.map(d => d.user_count)
  }];

  // 模型使用饼图配置
  const modelPieOptions = {
    chart: { type: 'donut' },
    labels: chatStats.map(s => `${s.model_id} (${s.mode})`),
    legend: {
      position: 'bottom',
      fontSize: '12px'
    },
    responsive: [{
      breakpoint: 480,
      options: {
        legend: { position: 'bottom' }
      }
    }],
    colors: [
      '#008FFB', '#00E396', '#FEB019', '#FF4560', '#775DD0',
      '#55efc4', '#81ecec', '#74b9ff', '#a29bfe', '#00b894'
    ]
  };

  const modelPieSeries = chatStats.map(s => parseInt(s.total_count));

  // 工具使用柱状图配置
  const toolBarOptions = {
    chart: {
      type: 'bar',
      toolbar: { show: false }
    },
    plotOptions: {
      bar: {
        horizontal: true,
        dataLabels: { position: 'top' }
      }
    },
    dataLabels: {
      enabled: true,
      offsetX: -6,
      style: { fontSize: '12px', colors: ['#fff'] }
    },
    xaxis: {
      categories: toolStats.map(t => t.tool_name)
    },
    colors: ['#00E396', '#FF4560'],
    legend: { position: 'top' }
  };

  const toolBarSeries = [
    { name: '成功', data: toolStats.map(t => parseInt(t.success_count)) },
    { name: '失败', data: toolStats.map(t => parseInt(t.failed_count)) }
  ];

  // 生成单个指标的图表配置
  const createSingleChartOptions = (title, color, isFloat = false) => ({
    chart: {
      type: 'area',
      toolbar: { show: false },
      zoom: { enabled: false },
      animations: { enabled: true },
      sparkline: { enabled: false }
    },
    dataLabels: { enabled: false },
    stroke: { curve: 'smooth', width: 2 },
    fill: {
      type: 'gradient',
      gradient: {
        shadeIntensity: 1,
        opacityFrom: 0.4,
        opacityTo: 0.1
      }
    },
    xaxis: {
      categories: timelineStats.map(d => d.time_slot.split(' ')[1] || d.time_slot),
      labels: {
        rotate: -45,
        style: { fontSize: '9px' },
        formatter: (value, timestamp, opts) => {
          if (opts && opts.i !== undefined) {
            const interval = timelineStats.length > 48 ? 8 : timelineStats.length > 24 ? 4 : 2;
            return opts.i % interval === 0 ? value : '';
          }
          return value;
        }
      }
    },
    yaxis: {
      labels: { 
        formatter: (val) => isFloat ? val.toFixed(1) : Math.round(val),
        style: { fontSize: '10px' }
      }
    },
    tooltip: {
      x: {
        formatter: (val, opts) => {
          if (opts && opts.dataPointIndex !== undefined && timelineStats[opts.dataPointIndex]) {
            return timelineStats[opts.dataPointIndex].time_slot;
          }
          return val;
        }
      }
    },
    colors: [color],
    title: {
      text: title,
      align: 'left',
      style: { fontSize: '14px', fontWeight: 600 }
    }
  });

  // 各指标的图表配置
  const chartConfigs = [
    {
      title: '活跃用户',
      color: theme.palette.primary.main,
      data: timelineStats.map(d => d.active_users || 0),
      total: timelineStats.reduce((sum, d) => sum + (d.active_users || 0), 0)
    },
    {
      title: '事件数',
      color: theme.palette.success.main,
      data: timelineStats.map(d => d.event_count || 0),
      total: timelineStats.reduce((sum, d) => sum + (d.event_count || 0), 0)
    },
    {
      title: '聊天次数',
      color: theme.palette.warning.main,
      data: timelineStats.map(d => d.chat_count || 0),
      total: timelineStats.reduce((sum, d) => sum + (d.chat_count || 0), 0)
    },
    {
      title: '工具成功',
      color: theme.palette.info.main,
      data: timelineStats.map(d => d.tool_success_count || 0),
      total: timelineStats.reduce((sum, d) => sum + (d.tool_success_count || 0), 0)
    },
    {
      title: '工具失败',
      color: theme.palette.error.main,
      data: timelineStats.map(d => d.tool_failed_count || 0),
      total: timelineStats.reduce((sum, d) => sum + (d.tool_failed_count || 0), 0)
    },
    {
      title: '平均使用度',
      color: theme.palette.secondary.main,
      data: timelineStats.map(d => Math.round((d.avg_usage_score || 0) * 100) / 100),
      total: (timelineStats.reduce((sum, d) => sum + (d.avg_usage_score || 0), 0) / (timelineStats.length || 1)).toFixed(2),
      isFloat: true
    }
  ];

  return (
    <>
      <Stack direction="row" alignItems="center" justifyContent="space-between" mb={2.5}>
        <Typography variant="h4">遥测统计 (PacificOceanAI)</Typography>
      </Stack>

      {loading && <LinearProgress sx={{ mb: 2 }} />}

      <Grid container spacing={gridSpacing}>
        {/* 概览卡片 */}
        <Grid item xs={12}>
          <Grid container spacing={gridSpacing}>
            <Grid item lg={3} sm={6} xs={12}>
              <StatsCard bgcolor={theme.palette.primary.main}>
                <CardContent>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                    <IconUsers size={40} />
                    <Box>
                      <Typography variant="h3" color="inherit">
                        {overview?.total_users || 0}
                      </Typography>
                      <Typography variant="body2" color="inherit" sx={{ opacity: 0.8 }}>
                        总用户数
                      </Typography>
                    </Box>
                  </Box>
                </CardContent>
              </StatsCard>
            </Grid>
            <Grid item lg={3} sm={6} xs={12}>
              <StatsCard bgcolor={theme.palette.success.dark}>
                <CardContent>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                    <IconCalendarStats size={40} />
                    <Box>
                      <Typography variant="h3" color="inherit">
                        {overview?.today_users || 0}
                      </Typography>
                      <Typography variant="body2" color="inherit" sx={{ opacity: 0.8 }}>
                        今日活跃
                      </Typography>
                    </Box>
                  </Box>
                </CardContent>
              </StatsCard>
            </Grid>
            <Grid item lg={3} sm={6} xs={12}>
              <StatsCard bgcolor={theme.palette.warning.dark}>
                <CardContent>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                    <IconMessages size={40} />
                    <Box>
                      <Typography variant="h3" color="inherit">
                        {overview?.total_sessions || 0}
                      </Typography>
                      <Typography variant="body2" color="inherit" sx={{ opacity: 0.8 }}>
                        总会话数
                      </Typography>
                    </Box>
                  </Box>
                </CardContent>
              </StatsCard>
            </Grid>
            <Grid item lg={3} sm={6} xs={12}>
              <StatsCard bgcolor={theme.palette.error.dark}>
                <CardContent>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                    <IconActivity size={40} />
                    <Box>
                      <Typography variant="h3" color="inherit">
                        {overview?.total_events || 0}
                      </Typography>
                      <Typography variant="body2" color="inherit" sx={{ opacity: 0.8 }}>
                        总事件数
                      </Typography>
                    </Box>
                  </Box>
                </CardContent>
              </StatsCard>
            </Grid>
          </Grid>
        </Grid>

        {/* 实时时间轴统计 */}
        <Grid item xs={12}>
          <MainCard>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2, flexWrap: 'wrap', gap: 1 }}>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                <Typography variant="h4">
                  <IconTimeline size={20} style={{ marginRight: 8, verticalAlign: 'middle' }} />
                  实时数据监控（每30分钟更新）
                </Typography>
                {timelineLoading && <LinearProgress sx={{ width: 100, ml: 2 }} />}
              </Box>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                {lastTimelineUpdate && (
                  <Typography variant="caption" color="textSecondary">
                    上次更新: {lastTimelineUpdate.toLocaleTimeString()}
                  </Typography>
                )}
                <Tooltip title="手动刷新">
                  <Chip
                    icon={<IconRefresh size={16} />}
                    label="刷新"
                    size="small"
                    clickable
                    onClick={loadTimelineStats}
                    disabled={timelineLoading}
                  />
                </Tooltip>
                <ToggleButtonGroup
                  value={timelineHours}
                  exclusive
                  onChange={(e, newValue) => newValue && setTimelineHours(newValue)}
                  size="small"
                >
                  <ToggleButton value="6">6小时</ToggleButton>
                  <ToggleButton value="12">12小时</ToggleButton>
                  <ToggleButton value="24">24小时</ToggleButton>
                  <ToggleButton value="48">48小时</ToggleButton>
                  <ToggleButton value="168">7天</ToggleButton>
                </ToggleButtonGroup>
              </Box>
            </Box>
            {timelineStats.length > 0 ? (
              <Grid container spacing={2}>
                {chartConfigs.map((config, index) => (
                  <Grid item xs={12} md={6} lg={4} key={index}>
                    <Box sx={{ 
                      border: '1px solid', 
                      borderColor: 'divider', 
                      borderRadius: 2, 
                      p: 1,
                      backgroundColor: 'background.paper'
                    }}>
                      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 1, px: 1 }}>
                        <Typography variant="subtitle2" color="textSecondary">
                          {config.title}
                        </Typography>
                        <Chip 
                          label={config.isFloat ? `均值: ${config.total}` : `总计: ${config.total}`}
                          size="small"
                          sx={{ 
                            backgroundColor: config.color + '20',
                            color: config.color,
                            fontWeight: 600
                          }}
                        />
                      </Box>
                      <Chart 
                        options={createSingleChartOptions(config.title, config.color, config.isFloat)} 
                        series={[{ name: config.title, data: config.data }]} 
                        type="area" 
                        height={180} 
                      />
                    </Box>
                  </Grid>
                ))}
              </Grid>
            ) : (
              <Box sx={{ height: 350, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                <Typography color="textSecondary">暂无时间轴数据</Typography>
              </Box>
            )}
          </MainCard>
        </Grid>

        {/* DAU 趋势图 */}
        <Grid item xs={12} lg={8}>
          <MainCard>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
              <Typography variant="h4">
                <IconUsers size={20} style={{ marginRight: 8, verticalAlign: 'middle' }} />
                日活跃用户 (DAU)
              </Typography>
              <ToggleButtonGroup
                value={dauDays}
                exclusive
                onChange={(e, newValue) => newValue && setDauDays(newValue)}
                size="small"
              >
                <ToggleButton value="7">7天</ToggleButton>
                <ToggleButton value="30">30天</ToggleButton>
                <ToggleButton value="90">90天</ToggleButton>
              </ToggleButtonGroup>
            </Box>
            {dauStats.length > 0 ? (
              <Chart options={dauChartOptions} series={dauChartSeries} type="area" height={300} />
            ) : (
              <Box sx={{ height: 300, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                <Typography color="textSecondary">暂无数据</Typography>
              </Box>
            )}
          </MainCard>
        </Grid>

        {/* 模型使用分布 */}
        <Grid item xs={12} lg={4}>
          <MainCard>
            <Typography variant="h4" sx={{ mb: 2 }}>
              <IconMessages size={20} style={{ marginRight: 8, verticalAlign: 'middle' }} />
              模型使用分布
            </Typography>
            {chatStats.length > 0 ? (
              <Chart options={modelPieOptions} series={modelPieSeries} type="donut" height={300} />
            ) : (
              <Box sx={{ height: 300, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                <Typography color="textSecondary">暂无数据</Typography>
              </Box>
            )}
          </MainCard>
        </Grid>

        {/* 工具使用统计 */}
        <Grid item xs={12} lg={6}>
          <MainCard>
            <Typography variant="h4" sx={{ mb: 2 }}>
              <IconTool size={20} style={{ marginRight: 8, verticalAlign: 'middle' }} />
              工具调用统计
            </Typography>
            {toolStats.length > 0 ? (
              <Chart options={toolBarOptions} series={toolBarSeries} type="bar" height={Math.max(300, toolStats.length * 40)} />
            ) : (
              <Box sx={{ height: 300, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                <Typography color="textSecondary">暂无数据</Typography>
              </Box>
            )}
          </MainCard>
        </Grid>

        {/* 文本操作统计 */}
        <Grid item xs={12} lg={6}>
          <MainCard>
            <Typography variant="h4" sx={{ mb: 2 }}>
              <IconFileText size={20} style={{ marginRight: 8, verticalAlign: 'middle' }} />
              文本操作统计
            </Typography>
            {textActionStats.length > 0 ? (
              <TableContainer component={Paper} variant="outlined">
                <Table size="small">
                  <TableHead>
                    <TableRow>
                      <TableCell>操作类型</TableCell>
                      <TableCell align="right">触发次数</TableCell>
                      <TableCell align="right">接受次数</TableCell>
                      <TableCell align="right">拒绝次数</TableCell>
                      <TableCell align="right">接受率</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {textActionStats.map((row) => {
                      const total = parseInt(row.accepted_count) + parseInt(row.rejected_count);
                      const acceptRate = total > 0 ? ((parseInt(row.accepted_count) / total) * 100).toFixed(1) : '-';
                      return (
                        <TableRow key={row.action}>
                          <TableCell>
                            <Chip label={row.action} size="small" color="primary" variant="outlined" />
                          </TableCell>
                          <TableCell align="right">{row.used_count}</TableCell>
                          <TableCell align="right">{row.accepted_count}</TableCell>
                          <TableCell align="right">{row.rejected_count}</TableCell>
                          <TableCell align="right">
                            {acceptRate !== '-' ? (
                              <Chip
                                label={`${acceptRate}%`}
                                size="small"
                                color={parseFloat(acceptRate) > 70 ? 'success' : parseFloat(acceptRate) > 40 ? 'warning' : 'error'}
                              />
                            ) : '-'}
                          </TableCell>
                        </TableRow>
                      );
                    })}
                  </TableBody>
                </Table>
              </TableContainer>
            ) : (
              <Box sx={{ height: 200, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                <Typography color="textSecondary">暂无数据</Typography>
              </Box>
            )}
          </MainCard>
        </Grid>

        {/* 工具审批统计 */}
        <Grid item xs={12} lg={6}>
          <MainCard>
            <Typography variant="h4" sx={{ mb: 2 }}>
              <IconGitBranch size={20} style={{ marginRight: 8, verticalAlign: 'middle' }} />
              工具审批统计
            </Typography>
            {toolApprovalStats.length > 0 ? (
              <TableContainer component={Paper} variant="outlined">
                <Table size="small">
                  <TableHead>
                    <TableRow>
                      <TableCell>工具名称</TableCell>
                      <TableCell align="right">批准次数</TableCell>
                      <TableCell align="right">拒绝次数</TableCell>
                      <TableCell align="right">批准率</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {toolApprovalStats.map((row) => {
                      const total = parseInt(row.approved_count) + parseInt(row.rejected_count);
                      const approveRate = total > 0 ? ((parseInt(row.approved_count) / total) * 100).toFixed(1) : '-';
                      return (
                        <TableRow key={row.tool_name}>
                          <TableCell>
                            <Chip label={row.tool_name} size="small" color="secondary" variant="outlined" />
                          </TableCell>
                          <TableCell align="right">{row.approved_count}</TableCell>
                          <TableCell align="right">{row.rejected_count}</TableCell>
                          <TableCell align="right">
                            {approveRate !== '-' ? (
                              <Chip
                                label={`${approveRate}%`}
                                size="small"
                                color={parseFloat(approveRate) > 70 ? 'success' : parseFloat(approveRate) > 40 ? 'warning' : 'error'}
                              />
                            ) : '-'}
                          </TableCell>
                        </TableRow>
                      );
                    })}
                  </TableBody>
                </Table>
              </TableContainer>
            ) : (
              <Box sx={{ height: 200, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                <Typography color="textSecondary">暂无数据</Typography>
              </Box>
            )}
          </MainCard>
        </Grid>

        {/* 聊天详细统计 */}
        <Grid item xs={12} lg={6}>
          <MainCard>
            <Typography variant="h4" sx={{ mb: 2 }}>
              <IconMessages size={20} style={{ marginRight: 8, verticalAlign: 'middle' }} />
              聊天详细统计
            </Typography>
            {chatStats.length > 0 ? (
              <TableContainer component={Paper} variant="outlined">
                <Table size="small">
                  <TableHead>
                    <TableRow>
                      <TableCell>模型</TableCell>
                      <TableCell>模式</TableCell>
                      <TableCell align="right">对话次数</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {chatStats.map((row, index) => (
                      <TableRow key={index}>
                        <TableCell>
                          <Chip label={row.model_id} size="small" variant="outlined" />
                        </TableCell>
                        <TableCell>
                          <Chip
                            label={row.mode}
                            size="small"
                            color={row.mode === 'agent' ? 'primary' : row.mode === 'chat' ? 'success' : 'default'}
                          />
                        </TableCell>
                        <TableCell align="right">{row.total_count}</TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </TableContainer>
            ) : (
              <Box sx={{ height: 200, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                <Typography color="textSecondary">暂无数据</Typography>
              </Box>
            )}
          </MainCard>
        </Grid>
      </Grid>
    </>
  );
};

export default Telemetry;

