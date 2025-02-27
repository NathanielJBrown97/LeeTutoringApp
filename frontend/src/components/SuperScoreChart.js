import React from 'react';
import { Line } from 'react-chartjs-2';
import { useTheme, useMediaQuery } from '@mui/material';
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
} from 'chart.js';

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend);

// Define your brand colors.
const brandBlue = '#0e1027';       // Composite line color.
const brandGold = '#b29600';
const brandGoldLight = '#d4a100';
const brandGoldLighter = '#f5dd5c';
const brandBlueLight = '#2a2f45';

const SuperScoreChart = ({ testData, filter }) => {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));

  // Filter tests based on selected filter.
  let filteredTests = [];
  if (filter === 'ACT') {
    filteredTests = testData.filter(test => (test.test || '').toUpperCase().includes('ACT'));
  } else {
    filteredTests = testData.filter(test => {
      const upperTest = (test.test || '').toUpperCase();
      return upperTest.includes('SAT') || upperTest.includes('PSAT');
    });
  }
  filteredTests.sort((a, b) => new Date(a.date) - new Date(b.date));

  // Build labels from test dates.
  const labels = filteredTests.map(test => {
    const d = new Date(test.date);
    return isNaN(d.getTime()) ? '' : d.toLocaleDateString();
  });

  let compositeArr = [];
  let additionalDatasets = [];

  if (filter === 'ACT') {
    // For ACT tests: expect arrays like [English, Math, Reading, Science, ACT_Total]
    let runningComposite = 0,
      runningEnglish = 0,
      runningMath = 0,
      runningReading = 0,
      runningScience = 0;
    const compositeData = [],
      englishData = [],
      mathData = [],
      readingData = [],
      scienceData = [];
    filteredTests.forEach(test => {
      let scores = null;
      if (Array.isArray(test.ACT_Scores) && test.ACT_Scores.length >= 5) {
        scores = test.ACT_Scores;
      } else if (Array.isArray(test.ACT) && test.ACT.length >= 5) {
        scores = test.ACT;
      }
      const english = scores ? parseFloat(scores[0]) || 0 : 0;
      const math = scores ? parseFloat(scores[1]) || 0 : 0;
      const reading = scores ? parseFloat(scores[2]) || 0 : 0;
      const science = scores ? parseFloat(scores[3]) || 0 : 0;
      const composite = scores ? parseFloat(scores[4]) || 0 : 0;
      
      runningComposite = Math.max(runningComposite, composite);
      runningEnglish = Math.max(runningEnglish, english);
      runningMath = Math.max(runningMath, math);
      runningReading = Math.max(runningReading, reading);
      runningScience = Math.max(runningScience, science);
      
      compositeData.push(runningComposite);
      englishData.push(runningEnglish);
      mathData.push(runningMath);
      readingData.push(runningReading);
      scienceData.push(runningScience);
    });
    compositeArr = compositeData;
    additionalDatasets = [
      { label: "English", data: englishData, fill: false, borderColor: brandGold, borderWidth: 2, pointRadius: 3, hidden: true, tension: 0.1 },
      { label: "Math", data: mathData, fill: false, borderColor: brandGoldLight, borderWidth: 2, pointRadius: 3, hidden: true, tension: 0.1 },
      { label: "Reading", data: readingData, fill: false, borderColor: brandGoldLighter, borderWidth: 2, pointRadius: 3, hidden: true, tension: 0.1 },
      { label: "Science", data: scienceData, fill: false, borderColor: brandBlueLight, borderWidth: 2, pointRadius: 3, hidden: true, tension: 0.1 },
    ];
  } else {
    // For SAT/PSAT tests: expect arrays like [EBRW, Math, Reading, Writing, SAT_Total]
    let runningComposite = 0, runningEBRW = 0, runningMath = 0;
    const compositeData = [], ebrwData = [], mathData = [];
    filteredTests.forEach(test => {
      let scores = null;
      if (Array.isArray(test.SAT_Scores) && test.SAT_Scores.length >= 5) {
        scores = test.SAT_Scores;
      } else if (Array.isArray(test.SAT) && test.SAT.length >= 5) {
        scores = test.SAT;
      }
      const ebrw = scores ? parseFloat(scores[0]) || 0 : 0;
      const math = scores ? parseFloat(scores[1]) || 0 : 0;
      const composite = scores ? parseFloat(scores[4]) || 0 : 0;
      
      runningComposite = Math.max(runningComposite, composite);
      runningEBRW = Math.max(runningEBRW, ebrw);
      runningMath = Math.max(runningMath, math);
      
      compositeData.push(runningComposite);
      ebrwData.push(runningEBRW);
      mathData.push(runningMath);
    });
    compositeArr = compositeData;
    additionalDatasets = [
      { label: "EBRW", data: ebrwData, fill: false, borderColor: brandGold, borderWidth: 2, pointRadius: 3, hidden: true, tension: 0.1 },
      { label: "Math", data: mathData, fill: false, borderColor: brandGoldLight, borderWidth: 2, pointRadius: 3, hidden: true, tension: 0.1 },
    ];
  }

  // For mobile ACT only: compute a forced y-axis tick configuration.
  let yTicksOptions;
  if (filter === 'ACT' && isMobile && compositeArr.length > 0) {
    const actualMin = Math.floor(Math.min(...compositeArr));
    const actualMax = Math.ceil(Math.max(...compositeArr));
    let minVal = actualMin;
    let maxVal = actualMax;
    let range = maxVal - minVal;
    // Expand the range if it’s too narrow (fewer than 4 units).
    if (range < 4) {
      const diff = 4 - range;
      minVal -= Math.floor(diff / 2);
      maxVal += Math.ceil(diff / 2);
    }
    // Force every integer value between min and max to be a tick.
    yTicksOptions = {
      autoSkip: false,
      stepSize: 1,
      min: minVal,
      max: maxVal,
      callback: value => value,
    };
  } else {
    yTicksOptions = { callback: value => value };
  }

  const options = {
    responsive: true,
    maintainAspectRatio: (isMobile && filter === 'ACT') ? false : true,
    plugins: {
      legend: { display: true },
      title: { display: true, text: `${filter} Super Score Over Time` },
    },
    scales: {
      x: {
        display: true,
        ticks: {
          autoSkip: true,
          maxTicksLimit: isMobile ? 5 : undefined,
        },
      },
      y: {
        beginAtZero: false,
        ticks: yTicksOptions,
      },
    },
  };

  const datasets = [
    {
      label: "Composite",
      data: compositeArr,
      fill: false,
      borderColor: brandBlue,
      borderWidth: 3,
      pointRadius: 4,
      tension: 0.1,
    },
    ...additionalDatasets,
  ];

  const data = { labels, datasets };

  // Only for mobile ACT, wrap the chart in a container with a fixed height.
  const containerStyle = (isMobile && filter === 'ACT') ? { height: 300 } : {};

  return <div style={containerStyle}><Line data={data} options={options} /></div>;
};

export default SuperScoreChart;
