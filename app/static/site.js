/**
 * Menu activator
 *
 * - Add active class (uk-active) to current menu item based on the current path
 */
function setActiveMenuItem() {
  // Normalize current path (remove trailing slash)
  const currentPath = window.location.pathname.replace(/\/$/, '');

  // Remove active class from all menu items
  document.querySelectorAll('.uk-navbar-nav li').forEach(item => {
      item.classList.remove('uk-active');
  });

  // Find and activate current menu item
  document.querySelectorAll('.uk-navbar-nav li a').forEach(link => {
      const href = link.getAttribute('href').replace(/\/$/, '');
      if (href === currentPath) {
          link.parentElement.classList.add('uk-active');
      }
  });
};

// Run when DOM is loaded
document.addEventListener('DOMContentLoaded', setActiveMenuItem);


/**
 * Theme switcher
 *
 * - Switch between light and dark mode
 * - Save theme to local storage
 */
document.addEventListener('DOMContentLoaded', function() {
  const themeSwitcher = document.getElementById('theme-switcher');
  const html = document.documentElement;

  const darkTextColor = '#fff';
  const lightTextColor = '#333';
  let duChart = null;
  let overlay2Chart = null;

  const savedTheme = localStorage.getItem('theme');
  if (savedTheme === 'dark') {
    themeSwitcher.querySelector('#theme-switcher>i').setAttribute('class', 'bi-sun');
    duChart = initializeChart('du-chart', darkTextColor);
    overlay2Chart = initializeChart('overlay2-chart', darkTextColor);
  } else {
    duChart = initializeChart('du-chart', lightTextColor);
    overlay2Chart = initializeChart('overlay2-chart', lightTextColor);
  }

  themeSwitcher.addEventListener('click', function(e) {
      e.preventDefault();
      if (html.classList.contains('dark')) {
          disableDarkMode();
      } else {
          enableDarkMode();
      }
  });

  function changeChartTextColor(chart, textColor) {
    chart.setOption({
      legend: {
        textStyle: {
          color: textColor
        }
      },
      series: [{
        label: {
          color: textColor
        }
      }]
    });
  }

  function enableDarkMode() {
      html.classList.add('uk-light', 'dark');
      localStorage.setItem('theme', 'dark');
      themeSwitcher.querySelector('#theme-switcher>i').setAttribute('class', 'bi-sun');
      changeChartTextColor(duChart, darkTextColor);
      changeChartTextColor(overlay2Chart, darkTextColor);
  }

  function disableDarkMode() {
      html.classList.remove('uk-light', 'dark');
      localStorage.setItem('theme', 'light');
      themeSwitcher.querySelector('#theme-switcher>i').setAttribute('class', 'bi-moon');
      changeChartTextColor(duChart, lightTextColor);
      changeChartTextColor(overlay2Chart, lightTextColor);
  }
});

(function() {
  const savedTheme = localStorage.getItem('theme');
  if (savedTheme === 'dark') {
    document.documentElement.classList.add('uk-light', 'dark');
  }
})();


/**
 * Chart initializer
 *
 * @param {string} textColor - Text color for the chart labels
 */
function initializeChart(elemId, textColor) {
  // Initialize the echarts instance based on the prepared dom
  const elem = document.getElementById(elemId);
  if (!elem) {
    return null;
  }

  const data = generateChartData(elemId);

  let chart = echarts.init(elem);
  // Specify the configuration items and data for the chart
  option = {
    tooltip: {
      trigger: 'item',
      formatter: function(params) {
        const size = humanSize(params.value, si, 2);
        return `${params.name}: ${size} | ${params.percent}%`;
      }
    },
    legend: {
      top: '5%',
      left: 'center',
      textStyle: {
        color: textColor
      },
      formatter: function(name) {
        return `${name} (${data[name].num})`;
      }
    },
    series: [
      {
        type: 'pie',
        radius: ['50%', '75%'],
        center: ['50%', '60%'],
        // adjust the start and end angle
        startAngle: 180,
        endAngle: 360,
        itemStyle: {
          borderRadius: 4,
          borderColor: '#fff',
          borderWidth: 2
        },
        data: Object.values(data),
        label: {
          color: textColor
        }
      }
    ]
  };

  // Display the chart using the configuration items and data just specified.
  chart.setOption(option);
  return chart;
}


/**
 * Table initializer
 *
 * @param {number|number[]} sizeCol - Column index or array of indices with size values (used for formatting)
 * @param {boolean} si - True to use metric (SI) units, aka powers of 1000. False to use binary (IEC), aka powers of 1024.
 * @param {string} tableId - Table ID
 * @param {array} nonSortableColumns - Array of column indexes to exclude from sorting
 * @param {array} nonSearchableColumns - Array of column indexes to exclude from searching
 *
 * @returns {object} - DataTable instance
 */
function initializeDataTable(config = {}) {
  let {
    sizeCol,
    si,
    nonSortableColumns = [],
    nonSearchableColumns = [],
    tableId = 'datatable'
  } = config;

  // Convert sizeCol to array if it's a single number
  const sizeColumns = Array.isArray(sizeCol) ? sizeCol : [sizeCol];

  const columnDefs = sizeColumns.map(col => ({
    'targets': col,
    'render': function(data, type, row) {
      if (type === 'display') {
        return humanSize(data, si=si, dp=2);
      }
      return data;
    }
  }));

  // Add non-sortable columns
  if (nonSortableColumns.length > 0) {
    columnDefs.push({
      'orderable': false,
      'targets': nonSortableColumns,
    });
  }

  // Add non-searchable columns
  if (nonSearchableColumns.length > 0) {
    // Add all size columns to non-searchable columns
    nonSearchableColumns.push(...sizeColumns);
    columnDefs.push({
      'searchable': false,
      'targets': nonSearchableColumns,
    });
  }

  let table = new DataTable('#' + tableId,
    {
      'paging': true,
      'pageLength': 50,
      'lengthChange': false,
      'searching': true,
      'ordering': true,
      'info': false,
      'autoWidth': true,
      'responsive': true,
      'search': {
        'caseInsensitive': true,
      },
      'language': {
        'search': '',
        'searchPlaceholder': 'Filtering...',
      },
      'layout': {
        'topStart': 'search',
        'topEnd': 'paging',
        'bottomStart': '',
        'bottomEnd': '',
      },
      'columnDefs': columnDefs,
      'order': [[sizeColumns[0], 'desc']],  // Sort by first size column
    }
  );
  return table;
};


/**
 * Format bytes as human-readable text.
 *
 * @param bytes Number of bytes.
 * @param si True to use metric (SI) units, aka powers of 1000. False to use binary (IEC), aka powers of 1024.
 * @param dp Number of decimal places to display.
 *
 * @return Formatted string.
 */
function humanSize(bytes, si=false, dp=2) {
  const denom = si ? 1000 : 1024;

  if (Math.abs(bytes) < denom) {
    return bytes + ' B';
  }

  const units = si
    ? ['kB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB']
    : ['KiB', 'MiB', 'GiB', 'TiB', 'PiB', 'EiB', 'ZiB', 'YiB'];
  let u = -1;
  const r = 10**dp;

  do {
    bytes /= denom;
    ++u;
  } while (Math.round(Math.abs(bytes) * r) / r >= denom && u < units.length - 1);


  return bytes.toFixed(dp) + '&nbsp;' + units[u];
}


/**
 * Generate chart data from elements with a specific prefix
 *
 * @param {string} prefix - The prefix of the HTML elements to look for (e.g., 'du-chart')
 * @returns {Array} Array of data objects with value and name properties
 */
function generateChartData(prefix) {
  const chartData = {};
  const elements = document.querySelectorAll(`[id^="${prefix}"]`);

  elements.forEach(element => {
    // Skip the main chart element if it exists
    if (element.id === prefix) return;

    const key = $(element).data('name');

    chartData[key] = {
      name: key,
      value: $(element).data('value'),
      num: $(element).data('num'),
    };
  });

  return chartData;
}
