import React from 'react';
import { useNavigate } from 'react-router-dom';
import { useSelector } from 'react-redux';
import { Pie } from '@nivo/pie';
import {
  selectTotalSizeBuildCache,
  selectTotalSizeContainers,
  selectTotalSizeImages,
  selectTotalSizeLogs,
  selectTotalSizeVolumes,
  selectCountBuildCache,
  selectCountContainers,
  selectCountImages,
  selectCountLogs,
  selectCountVolumes,
  selectIsDarkTheme,
} from '../AppSlice';
import { prettyCount, replaceWithNbsp } from '../util/fmt';
import prettyBytes from 'pretty-bytes';
import { Icon } from 'semantic-ui-react';

function PieChart() {
  const navigate = useNavigate();

  const isDarkTheme = useSelector(selectIsDarkTheme);
  const totalSizeImages = useSelector(selectTotalSizeImages);
  const totalSizeContainers = useSelector(selectTotalSizeContainers);
  const totalSizeVolumes = useSelector(selectTotalSizeVolumes);
  const totalSizeLogs = useSelector(selectTotalSizeLogs);
  const totalSizeBuildCache = useSelector(selectTotalSizeBuildCache);

  const countImages = useSelector(selectCountImages);
  const countContainers = useSelector(selectCountContainers);
  const countVolumes = useSelector(selectCountVolumes);
  const countLogs = useSelector(selectCountLogs);
  const countBuildCache = useSelector(selectCountBuildCache);

  const pieData = [
    {
      id: 'Images',
      label: `Images ${prettyCount(countImages)}`,
      to: '/images',
      value: totalSizeImages,
      color: '#cce2ff',
    },
    {
      id: 'Containers',
      label: `Containers ${prettyCount(countContainers)}`,
      to: '/containers',
      value: totalSizeContainers,
      color: '#f47560',
    },
    {
      id: 'Volumes',
      label: `Volumes ${prettyCount(countVolumes)}`,
      to: '/volumes',
      value: totalSizeVolumes,
      color: '#b2df8a',
    },
    {
      id: 'Logs',
      label: `Logs ${prettyCount(countLogs)}`,
      to: '/logs',
      value: totalSizeLogs,
      color: '#f1e15b',
    },
    {
      id: 'Build Cache',
      label: `Build Cache ${prettyCount(countBuildCache)}`,
      to: '/build-cache',
      value: totalSizeBuildCache,
      color: '#e8c1a0',
    },
  ];

  const textColor = isDarkTheme ? '#ffffff' : 'rgba(0,0,0,.87)';
  const backgroundColor = isDarkTheme ? '#2b2b2b' : '#ffffff';
  const borderColor = isDarkTheme ? '#444444' : '#dddddd';

  const tooltip = (d) => {
    const styles = {
      border: `1px solid ${borderColor}`,
      backgroundColor: backgroundColor,
      color: textColor,
      padding: '9px 12px',
    };
    return (
      <div style={styles}>
        <div>
          <Icon name="square" style={{ color: d.datum.color }} />
          {d.datum.id}: {d.datum.formattedValue}
        </div>
      </div>
    );
  };

  // https://nivo.rocks/storybook/?path=/story/pie--formatted-values
  // noinspection RequiredAttributes
  return (
    <Pie
      colors={{ datum: 'data.color' }}
      data={pieData}
      width={450}
      height={450}
      margin={{ top: 40, right: 40, bottom: 40, left: 40 }}
      innerRadius={0.7}
      padAngle={3} // Padding between each pie slice.
      cornerRadius={3} // Rounded slices.
      activeOuterRadiusOffset={4} // Extends active slice outer radius.
      borderWidth={2}
      borderColor={{
        from: 'color',
        modifiers: [['darker', 0.2]],
      }}
      arcLinkLabelsSkipAngle={10} // Skip label if corresponding slice's angle is lower than provided value.
      arcLinkLabelsTextColor={textColor}
      arcLinkLabelsThickness={2}
      arcLinkLabelsColor={{ from: 'color' }}
      arcLabelsSkipAngle={10} // Skip label if corresponding arc's angle is lower than provided value.
      arcLabelsTextColor={{
        from: 'color',
        modifiers: [['darker', 2]],
      }}
      valueFormat={(size) => replaceWithNbsp(prettyBytes(size))}
      tooltip={tooltip}
      legends={[
        {
          onClick: (d) => navigate(d.data.to),
          anchor: 'center',
          direction: 'column',
          justify: false,
          translateX: 0,
          translateY: 0,
          itemsSpacing: 0,
          itemWidth: 110,
          itemHeight: 24,
          itemTextColor: textColor,
          itemDirection: 'left-to-right',
          itemOpacity: 1,
          symbolSize: 20,
          symbolShape: 'circle',
          effects: [
            {
              on: 'hover',
              style: {
                itemTextColor: textColor,
              },
            },
          ],
        },
      ]}
    />
  );
}

export default PieChart;
