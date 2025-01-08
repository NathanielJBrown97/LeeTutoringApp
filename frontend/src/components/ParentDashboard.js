import React, { useEffect, useState, useContext } from 'react';
import { API_BASE_URL } from '../config';
import { AuthContext } from '../contexts/AuthContext';
import { useNavigate } from 'react-router-dom';

// --- Recharts ---
import {
  ResponsiveContainer,
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
} from 'recharts';

import {
  Container,
  AppBar,
  Toolbar,
  Typography,
  Button,
  Box,
  Grid,
  Select,
  MenuItem,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  CircularProgress,
  Divider,
  Tabs,
  Tab,
  Avatar,
  ListItemText,
  IconButton,
  Slide,
  useMediaQuery,
  useTheme,
  Chip,
} from '@mui/material';
import { styled } from '@mui/system';
import { Collapse } from '@mui/material';

// MUI Icons
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import ExpandLessIcon from '@mui/icons-material/ExpandLess';
import KeyboardArrowUpIcon from '@mui/icons-material/KeyboardArrowUp';
import KeyboardArrowDownIcon from '@mui/icons-material/KeyboardArrowDown';
import ArrowUpwardIcon from '@mui/icons-material/ArrowUpward';
import ArrowDownwardIcon from '@mui/icons-material/ArrowDownward';

// -------------------- Brand Colors --------------------
const brandBlue = '#0e1027';
const brandGold = '#b29600';
const lightBackground = '#fafafa';
const brandGoldLight = '#d4a100';
const brandGoldLighter = '#f5dd5c';
const brandBlueLight = '#2a2f45';


// The tab labels in an array (for desktop tabs or mobile dropdown).
const tabLabels = [
  'Recent Appointments',
  'School Goals',
  'Student Profile',
  'Test Data',
];

// -------------------- Styled Components --------------------
const StyledAppBar = styled(AppBar)(() => ({
  backgroundColor: brandBlue,
}));

const HeroSection = styled(Box)(({ theme }) => ({
  borderRadius: '8px',
  padding: '40px',
  marginTop: '24px',
  marginBottom: '24px',
  color: '#fff',
  background: `linear-gradient(to bottom right, ${brandBlue}, #2a2f45)`,
  boxShadow: '0 4px 20px rgba(0,0,0,0.1)',
  [theme.breakpoints.down('md')]: {
    padding: '24px',
    marginTop: '16px',
    marginBottom: '24px',
  },
  [theme.breakpoints.down('sm')]: {
    padding: '16px',
    marginTop: '12px',
    marginBottom: '16px',
  },
}));

const ContentWrapper = styled(Box)(({ theme }) => ({
  backgroundColor: '#fff',
  borderRadius: '16px',
  padding: '24px',
  marginBottom: '40px',
  boxShadow: '0 4px 20px rgba(0,0,0,0.1)',
  [theme.breakpoints.down('md')]: {
    padding: theme.spacing(2),
    marginBottom: '24px',
  },
  [theme.breakpoints.down('sm')]: {
    padding: theme.spacing(1.5),
    marginBottom: '16px',
  },
}));

const SectionContainer = styled(Paper)(({ theme }) => ({
  padding: '24px',
  borderRadius: '16px',
  backgroundColor: '#fff',
  boxShadow: '0 4px 20px rgba(0,0,0,0.05)',
  [theme.breakpoints.down('md')]: {
    padding: theme.spacing(2),
    borderRadius: '12px',
  },
}));

const SectionTitle = styled(Typography)(({ theme }) => ({
  marginBottom: '16px',
  fontWeight: 600,
  color: brandBlue,
  [theme.breakpoints.down('sm')]: {
    fontSize: '1rem',
    marginBottom: '12px',
  },
}));

const AppointmentCard = styled(Paper)(({ theme }) => ({
  borderRadius: '12px',
  padding: '16px',
  backgroundColor: '#fff',
  boxShadow: '0 2px 12px rgba(0,0,0,0.1)',
  marginBottom: '16px',
  [theme.breakpoints.down('sm')]: {
    padding: theme.spacing(1.5),
    marginBottom: theme.spacing(1),
  },
}));

const RootContainer = styled(Box)(() => ({
  minHeight: '100vh',
  backgroundColor: lightBackground,
}));

function TabPanel(props) {
  const { children, value, index, ...other } = props;
  return (
    <Box
      role="tabpanel"
      hidden={value !== index}
      {...other}
      sx={{ marginTop: '16px' }}
    >
      {value === index && children}
    </Box>
  );
}

// -- Helper function to safely parse integers --
const parseScore = (scoreStr) => {
  // Attempt to parse a number from the string. Returns NaN if invalid.
  const val = parseInt(scoreStr, 10);
  return isNaN(val) ? null : val;
};

// Percentage change helper
function computePercentageChange(oldVal, newVal) {
  // If oldVal is null, zero, or invalid, we can't do a % difference
  if (!oldVal || oldVal <= 0) return null;
  const diff = newVal - oldVal;
  const percent = (diff / oldVal) * 100; // e.g. +25 means +25%
  return percent;
}

// A combined function that returns a Chip showing both the numeric difference
// (+2 or -1) AND the % difference (+10.0% or -5.0%).
function renderChangeChip(oldVal, newVal) {
  if (oldVal === null || newVal === null) return '—'; // baseline or invalid

  const diff = newVal - oldVal; // e.g. +2 or -3
  const pct = computePercentageChange(oldVal, newVal);
  if (pct === null) return '—'; // baseline

  const signDiff = diff > 0 ? `+${diff}` : diff.toString(); // e.g. +2 or -2
  const signPct = pct > 0 ? `+${pct.toFixed(1)}%` : `${pct.toFixed(1)}%`;

  const arrowSymbol = diff > 0 ? '▲' : '▼';
  const combinedLabel = `${signDiff} ${arrowSymbol} ${signPct}`; // e.g. "+2 ▲ +10.0%"

  // Color
  let chipColor = 'default';
  if (diff > 0) chipColor = 'success';
  else if (diff < 0) chipColor = 'error';

  return (
    <Chip
      label={combinedLabel}
      color={chipColor}
      size="small"
      sx={{ fontWeight: 600 }}
    />
  );
}

const ParentDashboard = () => {
  const authState = useContext(AuthContext);
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm')); // phone

  const [associatedStudents, setAssociatedStudents] = useState([]);
  const [selectedStudentID, setSelectedStudentID] = useState(null);
  const [studentData, setStudentData] = useState(null);
  const [parentName, setParentName] = useState('Parent');
  const [parentPicture, setParentPicture] = useState(null);
  const [activeTab, setActiveTab] = useState(0);
  const [loading, setLoading] = useState(true);

  // For the 3-appointment scroller
  const [startIndex, setStartIndex] = useState(0);
  const [scrollDirection, setScrollDirection] = useState('down');

  const navigate = useNavigate();

  // -------------- Data Fetching --------------

  // 1) Check for valid token & fetch parent data
  useEffect(() => {
    const token = localStorage.getItem('authToken');
    if (!token) {
      navigate('/');
      return;
    }
    fetch(`${API_BASE_URL}/api/parent`, {
      method: 'GET',
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
      .then(async (response) => {
        if (!response.ok) {
          throw new Error('Failed to fetch parent data');
        }
        return response.json();
      })
      .then((data) => {
        setParentName(data.name || 'Parent');
        setParentPicture(data.picture || null);
      })
      .catch((error) => {
        console.error('Error fetching parent data:', error);
      });
  }, [navigate]);

  /**
   * 2) [NEW USEEFFECT] Attempt Automatic Association
   *    Then re-fetch associated students to ensure we have updated info.
   */
  useEffect(() => {
    const token = localStorage.getItem('authToken');
    if (!token) return;

    fetch(`${API_BASE_URL}/api/attemptAutomaticAssociation`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
      .then(async (res) => {
        if (!res.ok) {
          throw new Error('Auto-association attempt failed');
        }
        const data = await res.json();
        console.log('Auto-association response:', data);
        // Re-fetch to update local state
        fetchAssociatedStudents(token);
      })
      .catch((err) => console.error('Auto-association error:', err));
  }, []);

  // 3) Fetch associated students
  useEffect(() => {
    const token = localStorage.getItem('authToken');
    fetchAssociatedStudents(token);
  }, []);

  const fetchAssociatedStudents = (token) => {
    fetch(`${API_BASE_URL}/api/associated-students`, {
      method: 'GET',
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
      .then((res) => res.json())
      .then((data) => {
        setAssociatedStudents(data.associatedStudents);
        const params = new URLSearchParams(window.location.search);
        const queriedStudentID = params.get('studentID');
        if (
          queriedStudentID &&
          data.associatedStudents.includes(queriedStudentID)
        ) {
          setSelectedStudentID(queriedStudentID);
        } else if (data.associatedStudents.length > 0) {
          setSelectedStudentID(data.associatedStudents[0]);
        }
      })
      .catch((err) => console.error(err));
  };

  // 4) When a student is selected, fetch their data
  useEffect(() => {
    if (selectedStudentID) {
      const token = localStorage.getItem('authToken');
      setLoading(true);
      fetch(`${API_BASE_URL}/api/students/${selectedStudentID}`, {
        method: 'GET',
        headers: {
          Authorization: `Bearer ${token}`,
        },
      })
        .then(async (res) => {
          if (!res.ok) {
            throw new Error('Failed to fetch student data');
          }
          return res.json();
        })
        .then((data) => {
          setStudentData(data);
          setLoading(false);
          setStartIndex(0);
          setScrollDirection('down');
        })
        .catch((err) => {
          console.error(err);
          setLoading(false);
        });
    }
  }, [selectedStudentID]);

  // -------------- Handlers --------------
  const handleSignOut = () => {
    localStorage.removeItem('authToken');
    authState.updateToken(null);
    navigate('/');
  };

  const handleStudentChange = (event) => {
    const newStudentID = event.target.value;
    setSelectedStudentID(newStudentID);
    navigate(`/parentdashboard?studentID=${newStudentID}`);
  };

  // For desktop tabs
  const handleTabChange = (event, newValue) => {
    setActiveTab(newValue);
  };

  // Book With A Tutor
  const handleNavigateToBooking = () => {
    if (selectedStudentID) {
      navigate(`/booking?studentID=${selectedStudentID}`);
    } else {
      navigate('/booking');
    }
  };

  // 3-appointment scroller
  const handlePrevAppointment = () => {
    if (startIndex > 0) {
      setScrollDirection('up');
      setStartIndex(startIndex - 1);
    }
  };
  const handleNextAppointment = (appointmentsCount) => {
    if (startIndex + 3 < appointmentsCount) {
      setScrollDirection('down');
      setStartIndex(startIndex + 1);
    }
  };

  // -------------- Loading / Early Return --------------
  if (loading || !studentData) {
    return (
      <Box
        display="flex"
        justifyContent="center"
        alignItems="center"
        height="100vh"
        sx={{ backgroundColor: lightBackground }}
      >
        <CircularProgress />
      </Box>
    );
  }

  // -------------- After we have studentData --------------
  // Test data separation
  const testData = studentData.testData || [];
  const satTests = testData.filter((t) => {
    const upperTest = (t.test || '').toUpperCase();
    return upperTest.includes('SAT') || upperTest.includes('PSAT');
  });
  const actTests = testData.filter((t) => {
    const upperTest = (t.test || '').toUpperCase();
    return upperTest.includes('ACT');
  });

  // Ascending arrays for line charts
  const ascendingSatTests = [...satTests].sort(
    (a, b) => new Date(a.date || 0) - new Date(b.date || 0)
  );
  const ascendingActTests = [...actTests].sort(
    (a, b) => new Date(a.date || 0) - new Date(b.date || 0)
  );

  // Build line data for SAT
  const lineDataSAT = ascendingSatTests.map((testDoc, idx) => {
    const { EBRW, Math, Reading, Writing, SAT_Total } = renderSATScores(testDoc);
    return {
      name: testDoc.date || `Test #${idx + 1}`,
      EBRW: parseScore(EBRW) || 0,
      Math: parseScore(Math) || 0,
      Reading: parseScore(Reading) || 0,
      Writing: parseScore(Writing) || 0,
      Total: parseScore(SAT_Total) || 0,
    };
  });

  // Build line data for ACT
  const lineDataACT = ascendingActTests.map((testDoc, idx) => {
    const { English, MathVal, Reading, Science, ACT_Total } = renderACTScores(testDoc);
    return {
      name: testDoc.date || `Test #${idx + 1}`,
      English: parseScore(English) || 0,
      MathVal: parseScore(MathVal) || 0,
      Reading: parseScore(Reading) || 0,
      Science: parseScore(Science) || 0,
      Total: parseScore(ACT_Total) || 0,
    };
  });

  // Sort appointments (descending)
  const sortedAppointments = [...(studentData.homeworkCompletion || [])].sort((a, b) => {
    const dateA = a.date ? new Date(a.date) : new Date(0);
    const dateB = b.date ? new Date(b.date) : new Date(0);
    return dateB - dateA;
  });
  const appointmentsToShow = sortedAppointments.slice(startIndex, startIndex + 3);

  // Filter test dates (only upcoming)
  const testDates = (studentData.testDates || [])
    .filter((test) => {
      const testDateStr = test.test_date;
      if (!testDateStr) return false;
      const testDate = new Date(testDateStr);
      if (isNaN(testDate.getTime())) return false;
      const today = new Date();
      today.setHours(0, 0, 0, 0);
      return testDate >= today;
    })
    .sort((a, b) => new Date(a.test_date) - new Date(b.test_date));

  // Score parsing
  function renderSATScores(testDoc) {
    let EBRW = '';
    let Math = '';
    let Reading = '';
    let Writing = '';
    let SAT_Total = '';
    if (Array.isArray(testDoc.SAT_Scores) && testDoc.SAT_Scores.length > 0) {
      [EBRW, Math, Reading, Writing, SAT_Total] = testDoc.SAT_Scores;
    } else if (Array.isArray(testDoc.SAT) && testDoc.SAT.length > 0) {
      [EBRW, Math, Reading, Writing, SAT_Total] = testDoc.SAT;
    }
    return { EBRW, Math, Reading, Writing, SAT_Total };
  }

  function renderACTScores(testDoc) {
    // We'll store the second index as MathVal for clarity
    let English = '';
    let MathVal = '';
    let Reading = '';
    let Science = '';
    let ACT_Total = '';
    if (Array.isArray(testDoc.ACT_Scores) && testDoc.ACT_Scores.length > 0) {
      [English, MathVal, Reading, Science, ACT_Total] = testDoc.ACT_Scores;
    } else if (Array.isArray(testDoc.ACT) && testDoc.ACT.length > 0) {
      [English, MathVal, Reading, Science, ACT_Total] = testDoc.ACT;
    }
    return { English, MathVal, Reading, Science, ACT_Total };
  }

  // For the mobile cards: we compute changes with correct field names
  function getActPercentChanges(currentDoc, prevDoc) {
    if (!prevDoc) {
      return {
        ePercent: null,
        mPercent: null,
        rPercent: null,
        sPercent: null,
        tPercent: null,
      };
    }

    const currScores = renderACTScores(currentDoc);
    const prevScores = renderACTScores(prevDoc);

    // Use the same "MathVal" naming
    const ePercent = computePercentageChange(
      parseScore(prevScores.English),
      parseScore(currScores.English)
    );
    const mPercent = computePercentageChange(
      parseScore(prevScores.MathVal),
      parseScore(currScores.MathVal)
    );
    const rPercent = computePercentageChange(
      parseScore(prevScores.Reading),
      parseScore(currScores.Reading)
    );
    const sPercent = computePercentageChange(
      parseScore(prevScores.Science),
      parseScore(currScores.Science)
    );
    const tPercent = computePercentageChange(
      parseScore(prevScores.ACT_Total),
      parseScore(currScores.ACT_Total)
    );

    return { ePercent, mPercent, rPercent, sPercent, tPercent };
  }

  function getSatPercentChanges(currentDoc, prevDoc) {
    if (!prevDoc) {
      return {
        ePercent: null,
        mPercent: null,
        rPercent: null,
        wPercent: null,
        tPercent: null,
      };
    }

    const currScores = renderSATScores(currentDoc);
    const prevScores = renderSATScores(prevDoc);

    const ePercent = computePercentageChange(
      parseScore(prevScores.EBRW),
      parseScore(currScores.EBRW)
    );
    const mPercent = computePercentageChange(
      parseScore(prevScores.Math),
      parseScore(currScores.Math)
    );
    const rPercent = computePercentageChange(
      parseScore(prevScores.Reading),
      parseScore(currScores.Reading)
    );
    const wPercent = computePercentageChange(
      parseScore(prevScores.Writing),
      parseScore(currScores.Writing)
    );
    const tPercent = computePercentageChange(
      parseScore(prevScores.SAT_Total),
      parseScore(currScores.SAT_Total)
    );

    return { ePercent, mPercent, rPercent, wPercent, tPercent };
  }

  return (
    <RootContainer>
      {/* ---------- AppBar ---------- */}
      <StyledAppBar position="static" elevation={3}>
        <Toolbar
          disableGutters
          sx={{
            px: 2,
            py: 1,
          }}
        >
          {isMobile ? (
            <Box sx={{ width: '100%' }}>
              {/* Top row: Welcome + Sign Out (single line) */}
              <Box
                sx={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                  width: '100%',
                  mb: 1,
                  whiteSpace: 'nowrap',
                }}
              >
                <Typography
                  variant="h6"
                  sx={{
                    fontWeight: 'bold',
                    whiteSpace: 'nowrap',
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                  }}
                >
                  Welcome, {parentName}!
                </Typography>

                <Button
                  onClick={handleSignOut}
                  variant="contained"
                  sx={{
                    backgroundColor: brandGold,
                    color: '#fff',
                    fontWeight: 'bold',
                    textTransform: 'none',
                    whiteSpace: 'nowrap',
                    mr: 0.5,
                    '&:hover': {
                      backgroundColor: '#d4a100',
                    },
                  }}
                >
                  Sign Out
                </Button>
              </Box>

              {/* Second row: Avatar only */}
              <Box sx={{ display: 'flex', alignItems: 'center', py: 1 }}>
                <Avatar
                  src={parentPicture || undefined}
                  alt={parentName}
                  sx={{
                    bgcolor: parentPicture ? 'transparent' : brandGold,
                    color: '#fff',
                  }}
                >
                  {!parentPicture && parentName.charAt(0).toUpperCase()}
                </Avatar>
              </Box>
            </Box>
          ) : (
            // Desktop Layout
            <Box
              sx={{
                display: 'flex',
                flexDirection: 'row',
                alignItems: 'center',
                justifyContent: 'space-between',
                width: '100%',
                px: 2,
                py: 1,
              }}
            >
              {/* Left side: Welcome + Avatar */}
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                <Typography
                  variant="h5"
                  sx={{ fontWeight: 'bold', whiteSpace: 'nowrap' }}
                >
                  Welcome, {parentName}!
                </Typography>
                <Avatar
                  src={parentPicture || undefined}
                  alt={parentName}
                  sx={{
                    bgcolor: parentPicture ? 'transparent' : brandGold,
                    color: '#fff',
                  }}
                >
                  {!parentPicture && parentName.charAt(0).toUpperCase()}
                </Avatar>
              </Box>

              {/* Right side: Sign Out */}
              <Button
                onClick={handleSignOut}
                variant="contained"
                sx={{
                  backgroundColor: brandGold,
                  color: '#fff',
                  fontWeight: 'bold',
                  textTransform: 'none',
                  whiteSpace: 'nowrap',
                  mr: 0.5,
                  '&:hover': {
                    backgroundColor: '#d4a100',
                  },
                }}
              >
                Sign Out
              </Button>
            </Box>
          )}
        </Toolbar>
      </StyledAppBar>

      {/* ---------- Hero Section ---------- */}
      <Container maxWidth="xl" sx={{ marginTop: '24px' }}>
        <HeroSection>
          {/* Single line: "Dashboard Overview for:" + dropdown */}
          <Box
            display="flex"
            alignItems="center"
            flexWrap="nowrap"
            gap={2}
            sx={{ whiteSpace: 'nowrap' }}
          >
            <Typography
              variant={isMobile ? 'h6' : 'h5'}
              sx={{ fontWeight: 700, m: 0 }}
            >
              Overview for:
            </Typography>

            <Select
              size="small"
              value={selectedStudentID}
              onChange={handleStudentChange}
              variant="outlined"
              sx={{
                height: 40,
                backgroundColor: '#fff',
                borderRadius: '8px',
                fontWeight: 500,
                minWidth: isMobile ? 140 : 180,
              }}
            >
              {associatedStudents.map((student) => (
                <MenuItem key={student} value={student}>
                  {student}
                </MenuItem>
              ))}
            </Select>
          </Box>

          {/* Second row: descriptive text */}
          <Box mt={2}>
            <Typography
              variant={isMobile ? 'body1' : 'h6'}
              sx={{ opacity: 0.9 }}
            >
              View your student's recent appointments, goals, and test data.
            </Typography>
          </Box>
        </HeroSection>
      </Container>

      {/* ------------- Main Content ------------- */}
      <Container maxWidth="xl">
        <ContentWrapper>
          {/* Desktop vs. Mobile tabs */}
          {!isMobile ? (
            // Desktop: normal MUI tabs
            <Box display="flex" alignItems="center" justifyContent="space-between">
              <Tabs
                value={activeTab}
                onChange={handleTabChange}
                textColor="primary"
                indicatorColor="primary"
                variant="scrollable"
                scrollButtons="auto"
                sx={{ marginBottom: '16px' }}
              >
                {tabLabels.map((label, idx) => (
                  <Tab
                    key={label}
                    label={label}
                    sx={{ textTransform: 'none', fontWeight: 'bold' }}
                  />
                ))}
              </Tabs>

              <Button
                onClick={handleNavigateToBooking}
                variant="contained"
                sx={{
                  backgroundColor: brandBlue,
                  color: '#fff',
                  textTransform: 'none',
                  fontWeight: 'bold',
                  '&:hover': {
                    backgroundColor: '#1c2231',
                  },
                }}
              >
                Book With A Tutor
              </Button>
            </Box>
          ) : (
            // Mobile: drop-down for tab navigation
            <Box display="flex" alignItems="center" justifyContent="space-between" gap={2}>
              <Select
                value={String(activeTab)}
                onChange={(e) => {
                  setActiveTab(Number(e.target.value));
                }}
                sx={{ minWidth: 220 }}
              >
                {tabLabels.map((label, idx) => (
                  <MenuItem key={label} value={String(idx)}>
                    {label}
                  </MenuItem>
                ))}
              </Select>

              <Button
                onClick={handleNavigateToBooking}
                variant="contained"
                sx={{
                  backgroundColor: brandBlue,
                  color: '#fff',
                  textTransform: 'none',
                  fontWeight: 'bold',
                  '&:hover': {
                    backgroundColor: '#1c2231',
                  },
                }}
              >
                Book With A Tutor
              </Button>
            </Box>
          )}

          {/* =================== TAB PANELS =================== */}

          <TabPanel value={activeTab} index={0}>
            <SectionContainer>
              <SectionTitle variant="h6">Recent Appointments</SectionTitle>
              <Divider sx={{ marginBottom: '16px' }} />

              <Box display="flex" flexDirection="column" alignItems="center">
                {/* Up Arrow */}
                <IconButton
                  onClick={handlePrevAppointment}
                  disabled={startIndex === 0}
                  sx={{ mb: 2 }}
                >
                  <KeyboardArrowUpIcon fontSize="large" />
                </IconButton>

                {/* Desktop: Slide with ±10% offset; Mobile: no animation */}
                {!isMobile ? (
                  <Slide
                    key={startIndex}
                    in
                    direction={scrollDirection === 'down' ? 'down' : 'up'}
                    timeout={300}
                    mountOnEnter
                    unmountOnExit
                    onEnter={(node) => {
                      const offset = '10%';
                      node.style.transform =
                        scrollDirection === 'down'
                          ? `translateY(-${offset})`
                          : `translateY(${offset})`;
                    }}
                    onEntering={(node) => {
                      node.style.transform = 'translateY(0%)';
                    }}
                  >
                    <Box sx={{ maxWidth: 400 }}>
                      {appointmentsToShow.length > 0 ? (
                        appointmentsToShow.map((appt, index) => {
                          // Parse the date
                          const parsedDate = appt.date ? new Date(appt.date) : null;
                          const formattedDate = parsedDate
                            ? parsedDate.toLocaleDateString(undefined, {
                                year: 'numeric',
                                month: 'long',
                                day: 'numeric',
                              })
                            : 'N/A';

                          // Format homework percentage
                          const percentage = `${appt.percentage ?? 0}%`;

                          // Helper to convert decimal hours to Hours/Minutes
                          const formatDuration = (decimalHours) => {
                            if (!decimalHours || isNaN(decimalHours)) return 'N/A';
                            const totalMinutes = Math.round(decimalHours * 60);
                            const h = Math.floor(totalMinutes / 60);
                            const m = totalMinutes % 60;

                            if (h === 0 && m > 0) {
                              return `${m} Minute${m === 1 ? '' : 's'}`;
                            } else if (h > 0 && m === 0) {
                              return `${h} Hour${h === 1 ? '' : 's'}`;
                            } else if (h > 0 && m > 0) {
                              return `${h} Hour${h === 1 ? '' : 's'} and ${m} Minute${m === 1 ? '' : 's'}`;
                            } else {
                              return 'N/A';
                            }
                          };

                          const displayDuration = formatDuration(Number(appt.duration));

                          // Use attendance from the backend directly
                          const status = appt.attendance || 'N/A';

                          return (
                            <AppointmentCard key={index}>
                              <Typography variant="subtitle1" sx={{ fontWeight: 'bold', mb: 1 }}>
                                Appointment Date: {formattedDate}
                              </Typography>
                              <Typography variant="body2" sx={{ color: '#333', mb: 0.5 }}>
                                Homework Completed: {percentage}
                              </Typography>
                              <Typography variant="body2" sx={{ color: '#333', mb: 0.5 }}>
                                Duration: {displayDuration}
                              </Typography>
                              <Typography variant="body2" sx={{ color: '#333' }}>
                                Status: {status}
                              </Typography>
                            </AppointmentCard>
                          );
                        })
                      ) : (
                        <Typography variant="body2" color="textSecondary">
                          No recent appointments available.
                        </Typography>
                      )}
                    </Box>
                  </Slide>
                ) : (
                  // Mobile: No animation
                  <Box sx={{ maxWidth: '100%' }}>
                    {appointmentsToShow.length > 0 ? (
                      appointmentsToShow.map((appt, index) => {
                        // Parse the date
                        const parsedDate = appt.date ? new Date(appt.date) : null;
                        const formattedDate = parsedDate
                          ? parsedDate.toLocaleDateString(undefined, {
                              year: 'numeric',
                              month: 'long',
                              day: 'numeric',
                            })
                          : 'N/A';

                        // Format homework percentage
                        const percentage = `${appt.percentage ?? 0}%`;

                        // Helper to convert decimal hours to Hours/Minutes
                        const formatDuration = (decimalHours) => {
                          if (!decimalHours || isNaN(decimalHours)) return 'N/A';
                          const totalMinutes = Math.round(decimalHours * 60);
                          const h = Math.floor(totalMinutes / 60);
                          const m = totalMinutes % 60;

                          if (h === 0 && m > 0) {
                            return `${m} Minute${m === 1 ? '' : 's'}`;
                          } else if (h > 0 && m === 0) {
                            return `${h} Hour${h === 1 ? '' : 's'}`;
                          } else if (h > 0 && m > 0) {
                            return `${h} Hour${h === 1 ? '' : 's'} and ${m} Minute${m === 1 ? '' : 's'}`;
                          } else {
                            return 'N/A';
                          }
                        };

                        const displayDuration = formatDuration(Number(appt.duration));

                        // Use attendance from the backend directly
                        const status = appt.attendance || 'N/A';

                        return (
                          <AppointmentCard key={index}>
                            <Typography variant="subtitle1" sx={{ fontWeight: 'bold', mb: 1 }}>
                              Appointment Date: {formattedDate}
                            </Typography>
                            <Typography variant="body2" sx={{ color: '#333', mb: 0.5 }}>
                              Homework Completed: {percentage}
                            </Typography>
                            <Typography variant="body2" sx={{ color: '#333', mb: 0.5 }}>
                              Duration: {displayDuration}
                            </Typography>
                            <Typography variant="body2" sx={{ color: '#333' }}>
                              Status: {status}
                            </Typography>
                          </AppointmentCard>
                        );
                      })
                    ) : (
                      <Typography variant="body2" color="textSecondary">
                        No recent appointments available.
                      </Typography>
                    )}
                  </Box>
                )}

                {/* Down Arrow */}
                <IconButton
                  onClick={() => handleNextAppointment(sortedAppointments.length)}
                  disabled={startIndex + 3 >= sortedAppointments.length}
                  sx={{ mt: 2 }}
                >
                  <KeyboardArrowDownIcon fontSize="large" />
                </IconButton>
              </Box>
            </SectionContainer>
          </TabPanel>



         {/* ============== School Goals ============== */}
         <TabPanel value={activeTab} index={1}>
          <SectionContainer>
            <Divider sx={{ marginBottom: '16px' }} />

            {(studentData.goals || []).length > 0 ? (
              <TableContainer
                sx={{
                  overflowX: {
                    xs: 'auto',  // Scroll on small screens if needed
                    md: 'visible',
                  },
                }}
              >
                <Table>
                  <TableHead>
                    <TableRow>
                      {/* Center the School column header */}
                      <TableCell
                        align="center"
                        sx={{
                          fontWeight: 600,
                          whiteSpace: 'nowrap',
                        }}
                      >
                        School
                      </TableCell>

                      {/* Center the Percentile column header */}
                      <TableCell
                        align="center"
                        sx={{
                          fontWeight: 600,
                          whiteSpace: 'nowrap',
                        }}
                      >
                        Percentile (25th, 50th, 75th)
                      </TableCell>
                    </TableRow>
                  </TableHead>

                  <TableBody>
                    {studentData.goals.map((goal, index) => {
                      const schoolName = goal.university || goal.College || 'N/A';

                      // Ensure ACT_percentiles is joined by comma + space
                      const actPercentile =
                        goal.ACT_percentiles && goal.ACT_percentiles !== 'N/A'
                          ? Array.isArray(goal.ACT_percentiles)
                            ? goal.ACT_percentiles.join(', ')
                            : goal.ACT_percentiles
                          : 'No data';

                      // Ensure SAT_percentiles is joined by comma + space
                      const satPercentile =
                        goal.SAT_percentiles && goal.SAT_percentiles !== 'N/A'
                          ? Array.isArray(goal.SAT_percentiles)
                            ? goal.SAT_percentiles.join(', ')
                            : goal.SAT_percentiles
                          : 'No data';

                      return (
                        <TableRow key={index}>
                          {/* Center the School column values */}
                          <TableCell align="center">{schoolName}</TableCell>

                          {/* Center the ACT/SAT box */}
                          <TableCell align="center">
                            <Box
                              sx={{
                                display: 'flex',
                                alignItems: 'center',
                                justifyContent: 'center',  // Center horizontally
                              }}
                            >
                              <Box sx={{ mr: 2, textAlign: 'center' }}>
                                <Typography variant="subtitle2">ACT</Typography>
                                <Typography variant="body2" sx={{ whiteSpace: 'nowrap' }}>
                                  {actPercentile}
                                </Typography>
                              </Box>

                              <Divider orientation="vertical" flexItem />

                              <Box sx={{ ml: 2, textAlign: 'center' }}>
                                <Typography variant="subtitle2">SAT</Typography>
                                <Typography variant="body2" sx={{ whiteSpace: 'nowrap' }}>
                                  {satPercentile}
                                </Typography>
                              </Box>
                            </Box>
                          </TableCell>
                        </TableRow>
                      );
                    })}
                  </TableBody>
                </Table>
              </TableContainer>
            ) : (
              <Typography variant="body2" color="textSecondary">
                No school goals available.
              </Typography>
            )}
          </SectionContainer>
        </TabPanel>


          {/* ============== Student Profile ============== */}
          <TabPanel value={activeTab} index={2}>
            <SectionContainer>
              <SectionTitle variant="h6">Student Profile</SectionTitle>
              <Divider sx={{ marginBottom: '16px' }} />
              <TableContainer>
                <Table>
                  <TableBody>
                    <TableRow>
                      <TableCell sx={{ fontWeight: 600 }}>Name</TableCell>
                      <TableCell>{studentData.personal?.name || 'N/A'}</TableCell>
                    </TableRow>
                    <TableRow>
                      <TableCell sx={{ fontWeight: 600 }}>Grade</TableCell>
                      <TableCell>{studentData.personal?.grade || 'N/A'}</TableCell>
                    </TableRow>
                    <TableRow>
                      <TableCell sx={{ fontWeight: 600 }}>High School</TableCell>
                      <TableCell>{studentData.personal?.high_school || 'N/A'}</TableCell>
                    </TableRow>
                    <TableRow>
                      <TableCell sx={{ fontWeight: 600 }}>Accommodations</TableCell>
                      <TableCell>{studentData.personal?.accommodations || 'N/A'}</TableCell>
                    </TableRow>
                  </TableBody>
                </Table>
              </TableContainer>
            </SectionContainer>
          </TabPanel>

          {/* ============== Test Data ============== */}
          <TabPanel value={activeTab} index={3}>
          <Grid container spacing={4} direction={isMobile ? 'column' : 'row'}>
            {/* Right side: Test Scores (or top on mobile) */}
            <Grid item xs={12} md={10} order={{ xs: 1, md: 2 }}>
              <SectionContainer>
                {/* <SectionTitle variant="h6">Test Scores</SectionTitle>
                <Divider sx={{ marginBottom: '16px' }} /> */}

                {/* ================== SAT Scores ================== */}
                <Typography variant="h6" sx={{ fontWeight: 600, marginTop: '16px' }}>
                  SAT Scores
                </Typography>

                {lineDataSAT.length > 0 ? (
                  <Box sx={{ mt: 2, height: 300, width: '100%' }}>
                    <ResponsiveContainer width="100%" height="100%">
                      <LineChart
                        data={lineDataSAT}
                        margin={{ top: 20, right: 20, left: 20, bottom: 20 }}
                      >
                        <CartesianGrid stroke="#ccc" strokeDasharray="5 5" />
                        <XAxis dataKey="name" />
                        <YAxis
                          domain={[
                            (dataMin) => Math.max(0, dataMin - 2), 
                            (dataMax) => dataMax + 2,
                          ]}
                          allowDecimals={false}
                        />
                        <Tooltip />
                        <Legend />
                        <Line
                          type="monotone"
                          dataKey="EBRW"
                          stroke={brandGold}
                          strokeWidth={2}
                        />
                        <Line
                          type="monotone"
                          dataKey="Math"
                          stroke={brandGoldLight}
                          strokeWidth={2}
                        />
                        <Line
                          type="monotone"
                          dataKey="Reading"
                          stroke={brandGoldLighter}
                          strokeWidth={2}
                        />
                        <Line
                          type="monotone"
                          dataKey="Writing"
                          stroke={brandBlueLight}
                          strokeWidth={2}
                        />
                        <Line
                          type="monotone"
                          dataKey="Total"
                          stroke={brandBlue}
                          strokeWidth={3}
                          dot
                        />
                      </LineChart>
                    </ResponsiveContainer>
                  </Box>
                ) : (
                  <Typography variant="body2" color="textSecondary" sx={{ mt: 2 }}>
                    No SAT data to graph.
                  </Typography>
                )}

                {/* Desktop vs. Mobile: SAT */}
                {!isMobile ? (
                  /* ========== DESKTOP TABLE (SAT) ========== */
                  <Box sx={{ mt: 4 }}>
                    <TableContainer>
                      <Table>
                        <TableHead>
                          <TableRow>
                            <TableCell>Date</TableCell>
                            <TableCell>Test (Type)</TableCell>
                            <TableCell>EBRW</TableCell>
                            <TableCell>Math</TableCell>
                            <TableCell>Reading</TableCell>
                            <TableCell>Writing</TableCell>
                            <TableCell>Total</TableCell>
                          </TableRow>
                        </TableHead>
                        <TableBody>
                          {[...satTests]
                            .sort((a, b) => new Date(b.date || 0) - new Date(a.date || 0))
                            .map((testDoc, index, arr) => {
                              const { EBRW, Math, Reading, Writing, SAT_Total } = renderSATScores(testDoc);

                              // The older test is next in the array
                              const prevDoc = arr[index + 1];

                              const dateStr = testDoc.date || 'N/A';
                              const testLabel = `${testDoc.test || 'N/A'} (${testDoc.type || 'N/A'})`;

                              return (
                                <React.Fragment key={index}>
                                  {/* Row #1: raw scores */}
                                  <TableRow>
                                    <TableCell>{dateStr}</TableCell>
                                    <TableCell>{testLabel}</TableCell>
                                    <TableCell>{EBRW}</TableCell>
                                    <TableCell>{Math}</TableCell>
                                    <TableCell>{Reading}</TableCell>
                                    <TableCell>{Writing}</TableCell>
                                    <TableCell>{SAT_Total}</TableCell>
                                  </TableRow>

                                  {/* Row #2: Score Change */}
                                  <TableRow>
                                    <TableCell colSpan={2} sx={{ fontWeight: 600 }}>
                                      {prevDoc ? 'Score Change (%)' : 'Baseline'}
                                    </TableCell>
                                    <TableCell>
                                      {prevDoc
                                        ? renderChangeChip(
                                            parseScore(prevDoc.SAT_Scores?.[0] || 0),
                                            parseScore(EBRW)
                                          )
                                        : '—'}
                                    </TableCell>
                                    <TableCell>
                                      {prevDoc
                                        ? renderChangeChip(
                                            parseScore(prevDoc.SAT_Scores?.[1] || 0),
                                            parseScore(Math)
                                          )
                                        : '—'}
                                    </TableCell>
                                    <TableCell>
                                      {prevDoc
                                        ? renderChangeChip(
                                            parseScore(prevDoc.SAT_Scores?.[2] || 0),
                                            parseScore(Reading)
                                          )
                                        : '—'}
                                    </TableCell>
                                    <TableCell>
                                      {prevDoc
                                        ? renderChangeChip(
                                            parseScore(prevDoc.SAT_Scores?.[3] || 0),
                                            parseScore(Writing)
                                          )
                                        : '—'}
                                    </TableCell>
                                    <TableCell>
                                      {prevDoc
                                        ? renderChangeChip(
                                            parseScore(prevDoc.SAT_Scores?.[4] || 0),
                                            parseScore(SAT_Total)
                                          )
                                        : '—'}
                                    </TableCell>
                                  </TableRow>

                                  <TableRow>
                                    <TableCell colSpan={7} sx={{ px: 0 }}>
                                      <Divider
                                        sx={{
                                          my: 2,
                                          borderColor: 'black',
                                          borderWidth: 2,
                                        }}
                                      />
                                    </TableCell>
                                  </TableRow>
                                </React.Fragment>
                              );
                            })}
                        </TableBody>
                      </Table>
                    </TableContainer>
                  </Box>
                ) : (
                  /* ========== MOBILE CARDS (SAT) ========== */
                  <Box sx={{ mt: 4 }}>
                    {[...satTests]
                      .sort((a, b) => new Date(b.date || 0) - new Date(a.date || 0))
                      .map((testDoc, index, arr) => {
                        const { EBRW, Math, Reading, Writing, SAT_Total } = renderSATScores(testDoc);
                        const prevDoc = arr[index + 1];
                        const dateStr = testDoc.date || 'N/A';
                        const testLabel = `${testDoc.test || 'N/A'} (${testDoc.type || 'N/A'})`;

                        return (
                          <Paper
                            key={index}
                            sx={{
                              mb: 2,
                              p: 2,
                              borderRadius: '12px',
                              boxShadow: '0 2px 12px rgba(0,0,0,0.1)',
                            }}
                          >
                            <Typography variant="subtitle1" sx={{ fontWeight: 'bold', mb: 1 }}>
                              {dateStr} — {testLabel}
                            </Typography>

                            <Typography variant="body2">
                              <strong>EBRW:</strong> {EBRW}
                            </Typography>
                            <Typography variant="body2">
                              <strong>Math:</strong> {Math}
                            </Typography>
                            <Typography variant="body2">
                              <strong>Reading:</strong> {Reading}
                            </Typography>
                            <Typography variant="body2">
                              <strong>Writing:</strong> {Writing}
                            </Typography>
                            <Typography variant="body2">
                              <strong>Total:</strong> {SAT_Total}
                            </Typography>

                            <Box sx={{ mt: 1, fontWeight: 600 }}>
                              {prevDoc ? (
                                <Typography
                                  variant="body2"
                                  sx={{ mb: 0.5, fontWeight: 'bold' }}
                                >
                                  Score Change (%):
                                </Typography>
                              ) : (
                                <Typography variant="body2" sx={{ mb: 0.5 }}>
                                  Baseline
                                </Typography>
                              )}

                              {prevDoc && (
                                <Box sx={{ display: 'flex', gap: 2, flexWrap: 'wrap' }}>
                                  <Box display="flex" alignItems="center" gap={0.5}>
                                    <Typography variant="caption" sx={{ fontWeight: 600 }}>
                                      EBRW:
                                    </Typography>
                                    {renderChangeChip(
                                      parseScore(prevDoc.SAT_Scores?.[0] || 0),
                                      parseScore(EBRW)
                                    )}
                                  </Box>
                                  <Box display="flex" alignItems="center" gap={0.5}>
                                    <Typography variant="caption" sx={{ fontWeight: 600 }}>
                                      Math:
                                    </Typography>
                                    {renderChangeChip(
                                      parseScore(prevDoc.SAT_Scores?.[1] || 0),
                                      parseScore(Math)
                                    )}
                                  </Box>
                                  <Box display="flex" alignItems="center" gap={0.5}>
                                    <Typography variant="caption" sx={{ fontWeight: 600 }}>
                                      Reading:
                                    </Typography>
                                    {renderChangeChip(
                                      parseScore(prevDoc.SAT_Scores?.[2] || 0),
                                      parseScore(Reading)
                                    )}
                                  </Box>
                                  <Box display="flex" alignItems="center" gap={0.5}>
                                    <Typography variant="caption" sx={{ fontWeight: 600 }}>
                                      Writing:
                                    </Typography>
                                    {renderChangeChip(
                                      parseScore(prevDoc.SAT_Scores?.[3] || 0),
                                      parseScore(Writing)
                                    )}
                                  </Box>
                                  <Box display="flex" alignItems="center" gap={0.5}>
                                    <Typography variant="caption" sx={{ fontWeight: 600 }}>
                                      Total:
                                    </Typography>
                                    {renderChangeChip(
                                      parseScore(prevDoc.SAT_Scores?.[4] || 0),
                                      parseScore(SAT_Total)
                                    )}
                                  </Box>
                                </Box>
                              )}
                            </Box>
                          </Paper>
                        );
                      })}
                  </Box>
                )}

                {/* ================== ACT Scores ================== */}
                <Typography variant="h6" sx={{ fontWeight: 600, marginTop: '32px' }}>
                  ACT Scores
                </Typography>

                {lineDataACT.length > 0 ? (
                  <Box sx={{ mt: 2, height: 300, width: '100%' }}>
                    <ResponsiveContainer width="100%" height="100%">
                      <LineChart
                        data={lineDataACT}
                        margin={{ top: 20, right: 20, left: 20, bottom: 20 }}
                      >
                        <CartesianGrid stroke="#ccc" strokeDasharray="5 5" />
                        <XAxis dataKey="name" />
                        <YAxis
                          domain={[
                            (dataMin) => Math.max(0, dataMin - 2),
                            (dataMax) => dataMax + 2,
                          ]}
                          allowDecimals={false}
                        />
                        <Tooltip />
                        <Legend />
                        <Line
                          type="monotone"
                          dataKey="English"
                          stroke={brandGold}
                          strokeWidth={2}
                        />
                        <Line
                          type="monotone"
                          dataKey="Math"
                          stroke={brandGoldLight}
                          strokeWidth={2}
                        />
                        <Line
                          type="monotone"
                          dataKey="Reading"
                          stroke={brandGoldLighter}
                          strokeWidth={2}
                        />
                        <Line
                          type="monotone"
                          dataKey="Science"
                          stroke={brandBlueLight}
                          strokeWidth={2}
                        />
                        <Line
                          type="monotone"
                          dataKey="Total"
                          stroke={brandBlue}
                          strokeWidth={3}
                          dot
                        />
                      </LineChart>
                    </ResponsiveContainer>
                  </Box>
                ) : (
                  <Typography variant="body2" color="textSecondary" sx={{ mt: 2 }}>
                    No ACT data to graph.
                  </Typography>
                )}

                {/* Desktop vs. Mobile: ACT */}
                {!isMobile ? (
                  /* ========== DESKTOP TABLE (ACT) ========== */
                  <Box sx={{ mt: 4 }}>
                    <TableContainer>
                      <Table>
                        <TableHead>
                          <TableRow>
                            <TableCell>Date</TableCell>
                            <TableCell>Test (Type)</TableCell>
                            <TableCell>English</TableCell>
                            <TableCell>Math</TableCell>
                            <TableCell>Reading</TableCell>
                            <TableCell>Science</TableCell>
                            <TableCell>Total</TableCell>
                          </TableRow>
                        </TableHead>
                        <TableBody>
                          {[...actTests]
                            .sort((a, b) => new Date(b.date || 0) - new Date(a.date || 0))
                            .map((testDoc, index, arr) => {
                              const { English, MathVal, Reading, Science, ACT_Total } = renderACTScores(
                                testDoc
                              );
                              const prevDoc = arr[index + 1];

                              // If there's an older test, compute the differences
                              let oldScores = null;
                              if (prevDoc) {
                                oldScores = renderACTScores(prevDoc);
                              }

                              const dateStr = testDoc.date || 'N/A';
                              const testLabel = `${testDoc.test || 'N/A'} (${testDoc.type || 'N/A'})`;

                              return (
                                <React.Fragment key={index}>
                                  {/* Row #1: raw scores */}
                                  <TableRow>
                                    <TableCell>{dateStr}</TableCell>
                                    <TableCell>{testLabel}</TableCell>
                                    <TableCell>{English}</TableCell>
                                    <TableCell>{MathVal}</TableCell>
                                    <TableCell>{Reading}</TableCell>
                                    <TableCell>{Science}</TableCell>
                                    <TableCell>{ACT_Total}</TableCell>
                                  </TableRow>

                                  {/* Row #2: Score Change */}
                                  <TableRow>
                                    <TableCell colSpan={2} sx={{ fontWeight: 600 }}>
                                      {prevDoc ? 'Score Change (%)' : 'Baseline'}
                                    </TableCell>
                                    <TableCell>
                                      {prevDoc
                                        ? renderChangeChip(
                                            parseScore(oldScores?.English),
                                            parseScore(English)
                                          )
                                        : '—'}
                                    </TableCell>
                                    <TableCell>
                                      {prevDoc
                                        ? renderChangeChip(
                                            parseScore(oldScores?.MathVal),
                                            parseScore(MathVal)
                                          )
                                        : '—'}
                                    </TableCell>
                                    <TableCell>
                                      {prevDoc
                                        ? renderChangeChip(
                                            parseScore(oldScores?.Reading),
                                            parseScore(Reading)
                                          )
                                        : '—'}
                                    </TableCell>
                                    <TableCell>
                                      {prevDoc
                                        ? renderChangeChip(
                                            parseScore(oldScores?.Science),
                                            parseScore(Science)
                                          )
                                        : '—'}
                                    </TableCell>
                                    <TableCell>
                                      {prevDoc
                                        ? renderChangeChip(
                                            parseScore(oldScores?.ACT_Total),
                                            parseScore(ACT_Total)
                                          )
                                        : '—'}
                                    </TableCell>
                                  </TableRow>

                                  <TableRow>
                                    <TableCell colSpan={7} sx={{ px: 0 }}>
                                      <Divider
                                        sx={{
                                          my: 2,
                                          borderColor: 'black',
                                          borderWidth: 2,
                                        }}
                                      />
                                    </TableCell>
                                  </TableRow>
                                </React.Fragment>
                              );
                            })}
                        </TableBody>
                      </Table>
                    </TableContainer>
                  </Box>
                ) : (
                  /* ========== MOBILE CARDS (ACT) ========== */
                  <Box sx={{ mt: 4 }}>
                    {[...actTests]
                      .sort((a, b) => new Date(b.date || 0) - new Date(a.date || 0))
                      .map((testDoc, index, arr) => {
                        const { English, MathVal, Reading, Science, ACT_Total } = renderACTScores(
                          testDoc
                        );
                        const prevDoc = arr[index + 1];

                        const dateStr = testDoc.date || 'N/A';
                        const testLabel = `${testDoc.test || 'N/A'} (${testDoc.type || 'N/A'})`;

                        // For each field, only compute diffs if prevDoc exists
                        function makeDiffObject(oldScoreIndex, newScoreStr) {
                          const oldVal = prevDoc
                            ? parseScore(prevDoc.ACT_Scores?.[oldScoreIndex])
                            : null;
                          const newVal = parseScore(newScoreStr);
                          if (oldVal === null || newVal === null) return null;
                          const diff = newVal - oldVal;
                          return diff !== 0 ? { oldVal, newVal } : null;
                        }

                        const changes = [
                          { label: 'English', diffs: makeDiffObject(0, English) },
                          { label: 'Math', diffs: makeDiffObject(1, MathVal) },
                          { label: 'Reading', diffs: makeDiffObject(2, Reading) },
                          { label: 'Science', diffs: makeDiffObject(3, Science) },
                          { label: 'Total', diffs: makeDiffObject(4, ACT_Total) },
                        ].filter((item) => item.diffs !== null);

                        return (
                          <Paper
                            key={index}
                            sx={{
                              mb: 2,
                              p: 2,
                              borderRadius: '12px',
                              boxShadow: '0 2px 12px rgba(0,0,0,0.1)',
                            }}
                          >
                            <Typography variant="subtitle1" sx={{ fontWeight: 'bold', mb: 1 }}>
                              {dateStr} — {testLabel}
                            </Typography>

                            <Typography variant="body2">
                              <strong>English:</strong> {English}
                            </Typography>
                            <Typography variant="body2">
                              <strong>Math:</strong> {MathVal}
                            </Typography>
                            <Typography variant="body2">
                              <strong>Reading:</strong> {Reading}
                            </Typography>
                            <Typography variant="body2">
                              <strong>Science:</strong> {Science}
                            </Typography>
                            <Typography variant="body2">
                              <strong>Total:</strong> {ACT_Total}
                            </Typography>

                            {prevDoc && changes.length > 0 ? (
                              <Box sx={{ mt: 1, fontWeight: 600 }}>
                                <Typography variant="body2" sx={{ mb: 0.5, fontWeight: 'bold' }}>
                                  Score Change (%):
                                </Typography>
                                <Box sx={{ display: 'flex', gap: 2, flexWrap: 'wrap' }}>
                                  {changes.map((item, idx) => (
                                    <Box
                                      display="flex"
                                      alignItems="center"
                                      gap={0.5}
                                      key={idx}
                                    >
                                      <Typography variant="caption" sx={{ fontWeight: 600 }}>
                                        {item.label}:
                                      </Typography>
                                      {renderChangeChip(
                                        item.diffs.oldVal,
                                        item.diffs.newVal
                                      )}
                                    </Box>
                                  ))}
                                </Box>
                              </Box>
                            ) : !prevDoc ? (
                              <Typography variant="body2" sx={{ mt: 1 }}>
                                Baseline
                              </Typography>
                            ) : null}
                          </Paper>
                        );
                      })}
                  </Box>
                )}
              </SectionContainer>
            </Grid>

            {/* The test dates on the left (desktop) or last (mobile) */}
            <Grid item xs={12} md={2} order={{ xs: 2, md: 1 }}>
              <SectionContainer>
                <SectionTitle variant="h6">Testing Dates</SectionTitle>
                <Divider sx={{ marginBottom: '16px' }} />
                {testDates.length > 0 ? (
                  testDates.map((test, index) => (
                    <Box key={index} sx={{ mb: 2 }}>
                      <ListItemText
                        primary={test.test_date || 'N/A'}
                        secondary={test.test_type || 'N/A'}
                      />
                      <Divider sx={{ my: 1 }} />
                    </Box>
                  ))
                ) : (
                  <Typography variant="body2" color="textSecondary">
                    No upcoming tests.
                  </Typography>
                )}
              </SectionContainer>
            </Grid>
          </Grid>
        </TabPanel>

        </ContentWrapper>
      </Container>
    </RootContainer>
  );
};

export default ParentDashboard;
