import React, { useState, useEffect } from 'react';
import { Form, Input, Button, DatePicker, Space, Table, message, Modal, InputNumber, Switch } from 'antd';
import { LockOutlined, KeyOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';
import { createToken, listTokens, deleteToken, getToken } from '@/api/tokens';
import { Token } from '@/store/types';
import { changePassword } from '@/api/auth';
import { useDispatch, useSelector } from 'react-redux';
import { clearAuth } from '@/store/slices/authSlice';
import { useNavigate } from 'react-router-dom';
import Sidebar from '@/pages/sidebar/Sidebar';
import { RootState } from '@/store';
import { copyToClipboard } from '@/utils/clipboard';


const Settings: React.FC = () => {
  const currentUser = useSelector((state: RootState) => state.auth.user);
  const token = useSelector((state: RootState) => state.auth.token);
  const [tokens, setTokens] = useState<Token[]>([]);
  const [maxTokens, setMaxTokens] = useState(0);
  const [loading, setLoading] = useState(false);
  const [form] = Form.useForm();
  const dispatch = useDispatch();
  const navigate = useNavigate();
  const [isPasswordModalVisible, setIsPasswordModalVisible] = useState(false);
  const [isTokenModalVisible, setIsTokenModalVisible] = useState(false);
  const [expireDate, setExpireDate] = useState<dayjs.Dayjs | null>(null);
  const [expireDays, setExpireDays] = useState<number | null>(null);
  const [isNeverExpire, setIsNeverExpire] = useState(false);

  useEffect(() => {
    const fetchTokens = async () => {
      setLoading(true);
      try {
        const res = await listTokens();
        setTokens(res.tokens || []);
        setMaxTokens(res.maxItems || 0);
      } catch (error) {
        message.error('获取Token列表失败');
      } finally {
        setLoading(false);
      }
    };
    fetchTokens();
  }, []);

  const handleChangePassword = async (values: any) => {
    try {
      setLoading(true);
      await changePassword({
        oldPassword: values.currentPassword,
        newPassword: values.newPassword
      });
      
      message.success('密码修改成功，请重新登录');
      setTimeout(() => {
        dispatch(clearAuth());
        navigate('/login', { replace: true });
      }, 1000);
      
    } catch (error: any) {
      message.error(error.message || '修改密码失败');
    } finally {
      setLoading(false);
    }
  };

  const handleCreateToken = async (values: any) => {
    if (tokens.length >= maxTokens) {
      message.error(`已达到最大Token数量限制（${maxTokens}个）`);
      return;
    }
    try {
      setLoading(true);
      const res = await createToken({
        description: values.description,
        expiryDays: values.isNeverExpire ? null : values.expireDays
      });
      setTokens([...tokens, {
        id: res.tokenInfo.id,
        token: res.token,
        description: res.tokenInfo.description,
        createdAt: res.tokenInfo.createdAt,
        expiresAt: res.tokenInfo.expiresAt,
        lastUsedAt: res.tokenInfo.lastUsedAt
      }]);
      message.success('Token创建成功');
      form.resetFields();
      setIsTokenModalVisible(false);
    } catch (error) {
      message.error('创建Token失败');
    } finally {
      setLoading(false);
    }
  };

  const handleDeleteToken = async (tokenId: string) => {
    try {
      await deleteToken(tokenId);
      setTokens(tokens.filter(token => token.id !== tokenId));
      message.success('Token删除成功');
    } catch (error) {
      message.error('删除Token失败');
    }
  };

  const handleCopyToken = async (tokenId: string) => {
    try {
      const res = await getToken(tokenId);
      await copyToClipboard(
        res.token,
        () => message.success('Token已复制到剪贴板')
      );
    } catch (err) {
      message.error('复制失败');
    }
  };

  const columns = [
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
    },
    {
      title: 'Token',
      dataIndex: 'token',
      key: 'token',
      render: (text: string, record: Token) => {
        // const visibleLength = 4;
        // const masked = text.slice(0, visibleLength) + '*'.repeat(text.length - visibleLength * 2) + text.slice(-visibleLength);
        return (
          <Space>
            <span>{text}</span>
            <Button type="link" onClick={() => handleCopyToken(record.id)}>
              复制
            </Button>
          </Space>
        );
      },
    },
    {
      title: '创建时间',
      dataIndex: 'createdAt',
      key: 'createdAt',
      render: (text: string) => dayjs(text).format('YYYY-MM-DD HH:mm'),
    },
    {
      title: '过期时间',
      dataIndex: 'expiresAt',
      key: 'expiresAt',
      render: (text: string | null) => text ? dayjs(text).format('YYYY-MM-DD HH:mm') : '永不过期',
    },
    {
      title: '上次使用时间',
      dataIndex: 'lastUsedAt',
      key: 'lastUsedAt',
      render: (text: string | null) => text ? dayjs(text).format('YYYY-MM-DD HH:mm') : '从未使用',
    },
    {
      title: '操作',
      key: 'action',
      render: (_: string, record: Token) => (
        <Button type="link" danger onClick={() => handleDeleteToken(record.id)}>
          删除
        </Button>
      ),
    },
  ];

  const showPasswordModal = () => {
    setIsPasswordModalVisible(true);
  };

  const showTokenModal = () => {
    setIsTokenModalVisible(true);
  };

  const handleLogout = () => {
    dispatch(clearAuth());
    navigate('/clips/public-space', { replace: true });
  };

  const handleDateChange = (date: dayjs.Dayjs | null) => {
    if (date) {
      const newDate = date.startOf('day');
      if (!newDate.isSame(expireDate, 'day')) {
        setExpireDate(newDate);
      }
      if (!isNeverExpire) {
        const days = newDate.diff(dayjs().startOf('day'), 'day');
        if (days != expireDays) {
          setExpireDays(days);
        }
      }
    }
  };

  const handleDaysChange = (value: number | null) => {
    if (value && value != expireDays) {
      setExpireDays(value);
      if (!isNeverExpire) {
        const newDate = dayjs().startOf('day').add(value, 'day');
        if (!newDate.isSame(expireDate, 'day')) {
          setExpireDate(newDate);
        }
      }
    }
  };

  useEffect(() => {
    form.setFieldsValue({
      expireDays,
      expireAt: expireDate
    });
  }, [expireDays, expireDate, form]);

  return (
    <div className="tw-flex tw-flex-row tw-w-full tw-transition-all tw-mx-auto 
      tw-min-h-screen tw-justify-center tw-items-start tw-pl-56 tw-bg-gray-50">
      <Sidebar 
        navigate={navigate}
        username={currentUser?.username}
        token={token || ''}
        handleLogout={handleLogout}
      />
      <main className="tw-w-full tw-h-auto tw-flex-grow tw-shrink tw-flex 
        tw-flex-col tw-justify-start tw-items-center">
          <div className="tw-container tw-w-full tw-max-w-5xl tw-min-h-full 
            tw-flex tw-flex-col tw-justify-start tw-items-center tw-px-4 tw-pt-3 tw-pb-8"
          >
            {/* 合并后的设置区域 */}
            <div className="tw-w-full tw-bg-white tw-rounded-lg tw-shadow-sm tw-p-6">
              {/* 个人信息区域 */}
              <div className="tw-mb-8">
                <div className="tw-flex tw-justify-between tw-items-center tw-mb-4">
                  <h2 className="tw-text-xl tw-font-semibold">个人信息</h2>
                </div>
                <div className="tw-flex tw-items-center tw-gap-4">
                  <span className="tw-text-gray-700">用户名：{currentUser?.username}</span>
                  <Button 
                    type="primary" 
                    onClick={showPasswordModal}
                    icon={<LockOutlined />}
                  >
                    修改密码
                  </Button>
                </div>
              </div>

              {/* Access Tokens 区域 */}
              <div>
                <div className="tw-flex tw-justify-between tw-items-center tw-mb-4">
                  <h2 className="tw-text-xl tw-font-semibold">Access Tokens</h2>
                  <Button 
                    type="primary" 
                    onClick={showTokenModal}
                    icon={<KeyOutlined />}
                  >
                    创建Token
                  </Button>
                </div>
                <div className="tw-mb-4">
                  当前Token数量：{tokens.length}/{maxTokens}
                </div>
                <Table
                  columns={columns}
                  dataSource={tokens}
                  loading={loading}
                  rowKey="id"
                  pagination={false}
                />
              </div>
            </div>
          </div>
        </main>

      {/* 修改密码弹窗 */}
      <Modal
        title="修改密码"
        open={isPasswordModalVisible}
        onCancel={() => setIsPasswordModalVisible(false)}
        footer={null}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleChangePassword}
        >
          <Form.Item
            label="当前密码"
            name="currentPassword"
            rules={[
              { required: true, message: '请输入当前密码' },
              { min: 6, message: '密码长度不能小于6位' }
            ]}
          >
          </Form.Item>
          <Form.Item
            label="新密码"
            name="newPassword"
            rules={[
              { required: true, message: '请输入新密码' },
              { min: 8, message: '密码长度不能小于8位' }
            ]}
          >
          </Form.Item>
          <Form.Item
            label="确认新密码"
            name="confirmPassword"
            dependencies={['newPassword']}
            rules={[
              { required: true, message: '请确认新密码' },
              ({ getFieldValue }) => ({
                validator(_, value) {
                  if (!value || getFieldValue('newPassword') === value) {
                    return Promise.resolve();
                  }
                  return Promise.reject(new Error('两次输入的密码不一致'));
                },
              }),
            ]}
          >
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading}>
              修改密码
            </Button>
          </Form.Item>
        </Form>
      </Modal>

      {/* 创建Token弹窗 */}
      <Modal
        title="创建Token"
        open={isTokenModalVisible}
        onCancel={() => setIsTokenModalVisible(false)}
        footer={null}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleCreateToken}
          initialValues={{
            expireDays: null,
            expireAt: null,
            description: ''
          }}
        >
          <Form.Item
            name="description"
            label="描述"
            rules={[{ required: true, message: '请输入描述' }]}
          >
            <Input />
          </Form.Item>

          <Form.Item label="永不过期">
            <Switch 
              checked={isNeverExpire}
              onChange={(checked) => {
                setIsNeverExpire(checked);
              }}
            />
          </Form.Item>

          {!isNeverExpire && (
            <Form.Item label="过期设置">
              <Space align="start" className="tw-w-full">
                <Form.Item
                  name="expireAt"
                  rules={[{ required: true, message: '请选择过期时间' }]}
                  className="tw-flex-[2] tw-mb-0"
                >
                  <DatePicker
                    value={expireDate}
                    onChange={handleDateChange}
                    disabledDate={(current) => current && current < dayjs().startOf('day')}
                    className="tw-w-full"
                  />
                </Form.Item>
                <Form.Item
                  name="expireDays"
                  rules={[{ required: true, message: '请输入有效天数' }]}
                  className="tw-flex-1 tw-mb-0"
                >
                  <InputNumber
                    min={1}
                    value={expireDays}
                    onChange={handleDaysChange}
                    className="tw-w-full"
                  />
                </Form.Item>
                <span className="tw-mx-2 tw-leading-8">天</span>
              </Space>
            </Form.Item>
          )}

          <Form.Item>
            <Button type="primary" htmlType="submit">
              创建Token
            </Button>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default Settings; 