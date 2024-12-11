import React, { useMemo } from 'react';
import { Select, Tooltip } from 'antd';
import { GlobalOutlined, LockOutlined } from '@ant-design/icons';
import { SpaceWithPermission, User } from '@/store/types';

interface SpaceListProps {
  spaces: SpaceWithPermission[];
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
    const publicSpaces: SpaceWithPermission[] = [];
    const ownedSpaces: SpaceWithPermission[] = [];
    const sharedSpaces: SpaceWithPermission[] = [];

    spaces.forEach(space => {
      if (space.type === 'public') {
        publicSpaces.push(space);
      } else if (space.ownerId === currentUser?.id) {
        ownedSpaces.push(space);
      } else if (space.permission === 'edit' || space.permission === 'view') {
        sharedSpaces.push(space);
      }
    });

    // 修改排序逻辑，让 public-space 置顶
    const sortByName = (a: SpaceWithPermission, b: SpaceWithPermission) => {
      if (a.id === 'public-space') return -1;
      if (b.id === 'public-space') return 1;
      return a.name.localeCompare(b.name);
    };

    return {
      publicSpaces: publicSpaces.sort(sortByName),
      ownedSpaces: ownedSpaces.sort(sortByName),
      sharedSpaces: sharedSpaces.sort(sortByName),
    };
  }, [spaces, currentUser]);

  const renderSpaceIcon = (space: SpaceWithPermission) => {
    return space.type === 'public' ? 
      <GlobalOutlined style={{ marginRight: 8 }} /> : 
      <LockOutlined style={{ marginRight: 8 }} />;
  };

  const renderSpaceName = (space: SpaceWithPermission) => {
    if (space.id === 'public-space') {
      return `${space.name}（默认）`;
    }
    return space.name;
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
              <Tooltip title={`公共空间 - ${renderSpaceName(space)}`}>
                {renderSpaceIcon(space)}
                {renderSpaceName(space)}
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