import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Result, Button, message, Spin } from 'antd';
import { verifyInviteToken, acceptInvite } from '@/api/spaces';
import { useSelector } from 'react-redux';
import { RootState } from '@/store';
import { VerifyInviteTokenResponse } from '@/store/types';

const InviteConfirmation: React.FC = () => {
  const { token } = useParams<{ token: string }>();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [spaceInfo, setSpaceInfo] = useState<VerifyInviteTokenResponse | null>(null);
  const [error, setError] = useState<string>('');
  const isAuthenticated = useSelector((state: RootState) => !!state.auth.token);

  useEffect(() => {
    if (token) {
      verifyToken();
    }
  }, [token]);

  const verifyToken = async () => {
    try {
      const info = await verifyInviteToken(token!);
      
      // 如果用户已经是该空间的协作者，直接跳转到空间页面
      if (info.isCollaborator) {
        message.info('您已经是该空间的成员');
        navigate(`/clips/${info.spaceId}`);
        return;
      }
      
      setSpaceInfo(info);
    } catch (err: any) {
      setError(err.message || '无效的邀请链接');
    } finally {
      setLoading(false);
    }
  };

  const handleAccept = async () => {
    if (!spaceInfo) {
      message.error('空间信息无效');
      return;
    }
    try {
      await acceptInvite(token!);
      message.success('已成功加入空间');
      navigate(`/clips/${spaceInfo.spaceId}`);
    } catch (err: any) {
      message.error(err.message || '加入空间失败');
    }
  };

  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}>
        <Spin size="large" />
      </div>
    );
  }

  if (error) {
    return (
      <Result
        status="error"
        title="邀请确认失败"
        subTitle={error}
        extra={[
          <Button key="back" onClick={() => navigate('/')}>
            返回首页
          </Button>
        ]}
      />
    );
  }

  return (
    <Result
      status="info"
      title="空间协作邀请"
      subTitle={
        <>
          <p>您已被邀请加入空间：{spaceInfo?.spaceName}</p>
          <p>邀请人：{spaceInfo?.inviterName}</p>
          <p>权限级别：{spaceInfo?.permission === 'edit' ? '可编辑' : '可查看'}</p>
        </>
      }
      extra={[
        <Button
          type="primary"
          key="accept"
          onClick={handleAccept}
          disabled={!isAuthenticated}
        >
          接受邀请
        </Button>,
        !isAuthenticated && (
          <Button key="login" onClick={() => navigate(`/login?redirect=${encodeURIComponent(window.location.pathname)}`)}>
            请先登录
          </Button>
        )
      ].filter(Boolean)}
    />
  );
};

export default InviteConfirmation;
