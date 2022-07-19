import React from 'react';
import { useSelector } from 'react-redux';
import { selectDiskUsage, selectIsDarkTheme } from '../AppSlice';
import { replaceWithNbsp } from '../util/fmt';
import prettyBytes from 'pretty-bytes';

function DiskUsageBar() {
  const isDarkTheme = useSelector(selectIsDarkTheme);
  const diskUsage = useSelector(selectDiskUsage);

  if (diskUsage === null || diskUsage.Total === 0) {
    return null;
  }

  const percent = Math.round(diskUsage.Percent);
  const ratio = prettyBytes(diskUsage.Used) + ' / ' + prettyBytes(diskUsage.Total);
  const width = `${percent > 10 ? percent : 10}%`;

  let progressClassName = 'ui progress disk-usage-progress';
  if (isDarkTheme) {
    progressClassName += ' inverted';
  }

  return (
    <div className={progressClassName} data-percent={percent}>
      <div className="bar" style={{ width }} />
      <div className="label">
        {replaceWithNbsp(`Disk Usage (${percent} %)      `)}
        <strong>{replaceWithNbsp(ratio)}</strong>
      </div>
    </div>
  );
}

export default DiskUsageBar;
