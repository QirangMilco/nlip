import React, { useMemo } from 'react';
import { Select, Tooltip } from 'antd';
import { GlobalOutlined, LockOutlined } from '@ant-design/icons';
import { Space, User } from '@/store/types';
import { checkSpaceAccess } from '@/utils/permission';

interface SpaceListProps {
  spaces: Space[];
  currentUser: User | null;
  value?: string;
  loading?: boolean;
  onChange: (spaceId: string) => void;
}

const SpaceList: React.FC<SpaceListProps> = ({
  spaces,
  currentUser,
  value,
  loading,
  onChange,
}) => {
  // 对空间进行分类和排序
  const categorizedSpaces = useMemo(() => {
    const publicSpaces: Space[] = [];
    const ownedSpaces: Space[] = [];
    const sharedSpaces: Space[] = [];

    spaces.forEach(space => {
      if (space.type === 'public') {
        publicSpaces.push(space);
      } else if (space.ownerId === currentUser?.id) {
        ownedSpaces.push(space);
      } else if (checkSpaceAccess(space, currentUser)) {
        sharedSpaces.push(space);
      }
    });

    // 按名称排序
    const sortByName = (a: Space, b: Space) => a.name.localeCompare(b.name);
    return {
      publicSpaces: publicSpaces.sort(sortByName),
      ownedSpaces: ownedSpaces.sort(sortByName),
      sharedSpaces: sharedSpaces.sort(sortByName),
    };
  }, [spaces, currentUser]);

  const renderSpaceIcon = (space: Space) => {
    return space.type === 'public' ? 
      <GlobalOutlined style={{ marginRight: 8 }} /> : 
      <LockOutlined style={{ marginRight: 8 }} />;
  };

  return (
    <Select
      value={value}
      onChange={onChange}
      style={{ width: '100%' }}
      loading={loading}
      placeholder="选择空间"
      optionFilterProp="children"
      showSearch
    >
      {categorizedSpaces.publicSpaces.length > 0 && (
        <Select.OptGroup label="公共空间">
          {categorizedSpaces.publicSpaces.map(space => (
            <Select.Option key={space.id} value={space.id}>
              <Tooltip title={`公共空间 - ${space.name}`}>
                {renderSpaceIcon(space)}
                {space.name}
              </Tooltip>
            </Select.Option>
          ))}
        </Select.OptGroup>
      )}

      {categorizedSpaces.ownedSpaces.length > 0 && (
        <Select.OptGroup label="我的空间">
          {categorizedSpaces.ownedSpaces.map(space => (
            <Select.Option key={space.id} value={space.id}>
              <Tooltip title={`我的空间 - ${space.name}`}>
                {renderSpaceIcon(space)}
                {space.name}
              </Tooltip>
            </Select.Option>
          ))}
        </Select.OptGroup>
      )}

      {categorizedSpaces.sharedSpaces.length > 0 && (
        <Select.OptGroup label="共享空间">
          {categorizedSpaces.sharedSpaces.map(space => (
            <Select.Option key={space.id} value={space.id}>
              <Tooltip title={`共享空间 - ${space.name}`}>
                {renderSpaceIcon(space)}
                {space.name}
              </Tooltip>
            </Select.Option>
          ))}
        </Select.OptGroup>
      )}
    </Select>
  );
};

export default SpaceList; 