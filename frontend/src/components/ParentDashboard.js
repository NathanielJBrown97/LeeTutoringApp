// src/components/ParentDashboard.js

import React, { useEffect, useState, useContext } from 'react';
import { API_BASE_URL } from '../config';
import { AuthContext } from '../contexts/AuthContext';
import { useNavigate } from 'react-router-dom';
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
} from '@mui/material';
import { styled } from '@mui/system';

import KeyboardArrowUpIcon from '@mui/icons-material/KeyboardArrowUp';
import KeyboardArrowDownIcon from '@mui/icons-material/KeyboardArrowDown';

// -------------------- Brand Colors --------------------
const brandBlue = '#0e1027';
const brandGold = '#b29600';
const lightBackground = '#fafafa';

// The tab labels in an array (for desktop tabs or mobile dropdown).
const tabLabels = [
  'Recent Appointments',
  'School Goals',
  'Student Profile',
  'Test Data',
];

// -------------------- Styled Components --------------------
const StyledAppBar = styled(AppBar)(({ theme }) => ({
  backgroundColor: brandBlue,
  [theme.breakpoints.down('sm')]: {
    padding: theme.spacing(1),
  },
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

  // Sort appointments
  const sortedAppointments = [...(studentData.homeworkCompletion || [])].sort((a, b) => {
    const dateA = a.date ? new Date(a.date) : new Date(0);
    const dateB = b.date ? new Date(b.date) : new Date(0);
    return dateB - dateA;
  });
  const appointmentsToShow = sortedAppointments.slice(startIndex, startIndex + 3);

  // Filter test dates
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

  const testData = studentData.testData || [];
  const satTests = testData.filter((t) => {
    const upperTest = (t.test || '').toUpperCase();
    return upperTest.includes('SAT') || upperTest.includes('PSAT');
  });
  const actTests = testData.filter((t) => {
    const upperTest = (t.test || '').toUpperCase();
    return upperTest.includes('ACT');
  });

  // Score parsing
  const renderSATScores = (testDoc) => {
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
  };
  const renderACTScores = (testDoc) => {
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
    return { English, Math: MathVal, Reading, Science, ACT_Total };
  };

  // Different heading size on mobile
  const heroHeadingVariant = isMobile ? 'h4' : 'h3';

  return (
    <RootContainer>
      {/* ------------- AppBar ------------- */}
      <StyledAppBar position="static" elevation={3}>
        <Toolbar
          sx={{
            display: 'flex',
            flexDirection: isMobile ? 'column' : 'row',
            alignItems: isMobile ? 'flex-start' : 'center',
            justifyContent: 'space-between',
            gap: isMobile ? 1 : 0,
          }}
        >
          <Box display="flex" alignItems="center" gap={2}>
            <Typography variant={isMobile ? 'h6' : 'h5'} sx={{ fontWeight: 'bold' }}>
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

          <Button
            onClick={handleSignOut}
            variant="contained"
            sx={{
              backgroundColor: brandGold,
              color: '#fff',
              fontWeight: 'bold',
              textTransform: 'none',
              '&:hover': {
                backgroundColor: '#d4a100',
              },
              alignSelf: isMobile ? 'flex-start' : 'center',
            }}
          >
            Sign Out
          </Button>
        </Toolbar>
      </StyledAppBar>

      {/* ------------- Hero Section ------------- */}
      <Container maxWidth="xl">
        <HeroSection>
          <Box
            display="flex"
            alignItems="center"
            justifyContent={isMobile ? 'flex-start' : 'space-between'}
            flexWrap="wrap"
          >
            <Box mb={{ xs: 2, md: 0 }}>
              <Typography variant={heroHeadingVariant} sx={{ fontWeight: 700, marginBottom: '8px' }}>
                {studentData.personal?.name || 'Student'}
              </Typography>
              <Typography
                variant={isMobile ? 'body1' : 'h6'}
                sx={{ opacity: 0.9 }}
              >
                Track appointments, view previous and upcoming tests, and compare
                to your student's goal schools.
              </Typography>
            </Box>

            <Box sx={{ marginLeft: isMobile ? 0 : 'auto' }}>
              <Select
                value={selectedStudentID}
                onChange={handleStudentChange}
                variant="outlined"
                sx={{
                  minWidth: isMobile ? '150px' : '200px',
                  backgroundColor: '#fff',
                  borderRadius: '8px',
                  fontWeight: 500,
                }}
              >
                {associatedStudents.map((student) => (
                  <MenuItem key={student} value={student}>
                    {student}
                  </MenuItem>
                ))}
              </Select>
            </Box>
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
                sx={{ minWidth: 160 }}
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

          {/* Tab Panels */}
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

                {/* 
                  Desktop: Slide with ±10% offset
                  Mobile: No animation — just a <Box> containing the appointments
                */}
                {!isMobile ? (
                  <Slide
                    key={startIndex}
                    in
                    direction={scrollDirection === 'down' ? 'down' : 'up'}
                    timeout={300}
                    mountOnEnter
                    unmountOnExit
                    onEnter={(node) => {
                      const offset = '10%'; // Original partial slide
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
                          const parsedDate = appt.date ? new Date(appt.date) : null;
                          const formattedDate = parsedDate
                            ? parsedDate.toLocaleDateString(undefined, {
                                year: 'numeric',
                                month: 'long',
                                day: 'numeric',
                              })
                            : 'N/A';

                          const duration = '1 hr';
                          const status = 'Attended';
                          const percentage = appt.percentage || '0%';

                          return (
                            <AppointmentCard key={index}>
                              <Typography variant="subtitle1" sx={{ fontWeight: 'bold', mb: 1 }}>
                                Appointment Date: {formattedDate}
                              </Typography>
                              <Typography variant="body2" sx={{ color: '#333', mb: 0.5 }}>
                                Homework Completed: {percentage}
                              </Typography>
                              <Typography variant="body2" sx={{ color: '#333', mb: 0.5 }}>
                                Duration: {duration}
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
                        const parsedDate = appt.date ? new Date(appt.date) : null;
                        const formattedDate = parsedDate
                          ? parsedDate.toLocaleDateString(undefined, {
                              year: 'numeric',
                              month: 'long',
                              day: 'numeric',
                            })
                          : 'N/A';

                        const duration = '1 hr';
                        const status = 'Attended';
                        const percentage = appt.percentage || '0%';

                        return (
                          <AppointmentCard key={index}>
                            <Typography variant="subtitle1" sx={{ fontWeight: 'bold', mb: 1 }}>
                              Appointment Date: {formattedDate}
                            </Typography>
                            <Typography variant="body2" sx={{ color: '#333', mb: 0.5 }}>
                              Homework Completed: {percentage}
                            </Typography>
                            <Typography variant="body2" sx={{ color: '#333', mb: 0.5 }}>
                              Duration: {duration}
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

          <TabPanel value={activeTab} index={1}>
  <SectionContainer>
    <SectionTitle variant="h6">
      School Goals: 25th Percentile, 50th Percentile, 75th Percentile
    </SectionTitle>
    <Divider sx={{ marginBottom: '16px' }} />

    {(studentData.goals || []).length > 0 ? (
      <TableContainer>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell sx={{ fontWeight: 600 }}>School</TableCell>
              <TableCell sx={{ fontWeight: 600 }}>
                ACT Percentile
                <br />
                <Typography variant="caption">25th, 50th, 75th</Typography>
              </TableCell>
              <TableCell sx={{ fontWeight: 600 }}>
                SAT Percentile
                <br />
                <Typography variant="caption">25th, 50th, 75th</Typography>
              </TableCell>
            </TableRow>
          </TableHead>

          <TableBody>
            {studentData.goals.map((goal, index) => {
              const schoolName = goal.university || goal.College || 'N/A';

              // ---- ACT Column ----
              // Build an array exactly like your original code, but WITHOUT "ACT:" prefix
              let actArray = [];
              if (goal.ACT_percentiles && goal.ACT_percentiles !== 'N/A') {
                // For example: '23, 25, 29'
                // If multiple values need to be stored, push them, then join below
                actArray.push(goal.ACT_percentiles);
              }
              // Join everything with comma + space
              const actPercentile = actArray.length > 0
                ? actArray.join(', ')
                : 'No data';

              // ---- SAT Column ----
              let satArray = [];
              if (goal.SAT_percentiles && goal.SAT_percentiles !== 'N/A') {
                satArray.push(goal.SAT_percentiles);
              }
              const satPercentile = satArray.length > 0
                ? satArray.join(', ')
                : 'No data';

              return (
                <TableRow key={index}>
                  <TableCell>{schoolName}</TableCell>
                  <TableCell>{actPercentile}</TableCell>
                  <TableCell>{satPercentile}</TableCell>
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

          <TabPanel value={activeTab} index={3}>
            <Grid container spacing={4}>
              <Grid item xs={12} md={2}>
                <SectionContainer>
                  <SectionTitle variant="h6">Testing Dates</SectionTitle>
                  <Divider sx={{ marginBottom: '16px' }} />
                  {((studentData.testDates || []).length > 0 && (
                    (studentData.testDates || [])
                      .filter((test) => {
                        const testDateStr = test.test_date;
                        if (!testDateStr) return false;
                        const testDate = new Date(testDateStr);
                        if (isNaN(testDate.getTime())) return false;
                        const today = new Date();
                        today.setHours(0, 0, 0, 0);
                        return testDate >= today;
                      })
                      .sort((a, b) => new Date(a.test_date) - new Date(b.test_date))
                      .map((test, index) => (
                        <Box key={index} sx={{ mb: 2 }}>
                          <ListItemText
                            primary={test.test_date || 'N/A'}
                            secondary={test.test_type || 'N/A'}
                            sx={{ paddingLeft: 0 }}
                          />
                          <Divider sx={{ my: 1 }} />
                        </Box>
                      ))
                  )) || (
                    <Typography variant="body2" color="textSecondary">
                      No upcoming tests.
                    </Typography>
                  )}
                </SectionContainer>
              </Grid>

              <Grid item xs={12} md={10}>
                <SectionContainer>
                  <SectionTitle variant="h6">Test Scores</SectionTitle>
                  <Divider sx={{ marginBottom: '16px' }} />

                  {/* SAT/PSAT Scores */}
                  <Typography variant="h6" sx={{ fontWeight: 600, marginTop: '16px' }}>
                    SAT/PSAT Scores
                  </Typography>
                  <TableContainer sx={{ marginBottom: '24px' }}>
                    <Table>
                      <TableHead>
                        <TableRow>
                          <TableCell sx={{ fontWeight: 600, verticalAlign: 'top' }}>
                            Date
                          </TableCell>
                          <TableCell sx={{ fontWeight: 600, verticalAlign: 'top' }}>
                            Test (Type)
                          </TableCell>
                          <TableCell sx={{ fontWeight: 600, verticalAlign: 'top' }}>
                            EBRW
                          </TableCell>
                          <TableCell sx={{ fontWeight: 600, verticalAlign: 'top' }}>
                            Math
                          </TableCell>
                          <TableCell sx={{ fontWeight: 600, verticalAlign: 'top' }}>
                            Reading
                          </TableCell>
                          <TableCell sx={{ fontWeight: 600, verticalAlign: 'top' }}>
                            Writing
                          </TableCell>
                          <TableCell sx={{ fontWeight: 600, verticalAlign: 'top' }}>
                            SAT_Total
                          </TableCell>
                        </TableRow>
                      </TableHead>
                      <TableBody>
                        {satTests.length > 0 ? (
                          satTests.map((testDoc, index) => {
                            const { EBRW, Math, Reading, Writing, SAT_Total } =
                              renderSATScores(testDoc);
                            return (
                              <TableRow key={index}>
                                <TableCell sx={{ verticalAlign: 'top' }}>
                                  {testDoc.date || 'N/A'}
                                </TableCell>
                                <TableCell sx={{ verticalAlign: 'top' }}>
                                  {(testDoc.test || 'N/A')} ({testDoc.type || 'N/A'})
                                </TableCell>
                                <TableCell sx={{ verticalAlign: 'top' }}>{EBRW}</TableCell>
                                <TableCell sx={{ verticalAlign: 'top' }}>{Math}</TableCell>
                                <TableCell sx={{ verticalAlign: 'top' }}>{Reading}</TableCell>
                                <TableCell sx={{ verticalAlign: 'top' }}>{Writing}</TableCell>
                                <TableCell sx={{ verticalAlign: 'top' }}>{SAT_Total}</TableCell>
                              </TableRow>
                            );
                          })
                        ) : (
                          <TableRow>
                            <TableCell colSpan={7}>
                              <Typography variant="body2" color="textSecondary">
                                No SAT/PSAT tests available.
                              </Typography>
                            </TableCell>
                          </TableRow>
                        )}
                      </TableBody>
                    </Table>
                  </TableContainer>

                  {/* ACT Scores */}
                  <Typography variant="h6" sx={{ fontWeight: 600 }}>
                    ACT Scores
                  </Typography>
                  <TableContainer>
                    <Table>
                      <TableHead>
                        <TableRow>
                          <TableCell sx={{ fontWeight: 600, verticalAlign: 'top' }}>
                            Date
                          </TableCell>
                          <TableCell sx={{ fontWeight: 600, verticalAlign: 'top' }}>
                            Test (Type)
                          </TableCell>
                          <TableCell sx={{ fontWeight: 600, verticalAlign: 'top' }}>
                            English
                          </TableCell>
                          <TableCell sx={{ fontWeight: 600, verticalAlign: 'top' }}>
                            Math
                          </TableCell>
                          <TableCell sx={{ fontWeight: 600, verticalAlign: 'top' }}>
                            Reading
                          </TableCell>
                          <TableCell sx={{ fontWeight: 600, verticalAlign: 'top' }}>
                            Science
                          </TableCell>
                          <TableCell sx={{ fontWeight: 600, verticalAlign: 'top' }}>
                            ACT_Total
                          </TableCell>
                        </TableRow>
                      </TableHead>
                      <TableBody>
                        {actTests.length > 0 ? (
                          actTests.map((testDoc, index) => {
                            const {
                              English,
                              Math: MathVal,
                              Reading,
                              Science,
                              ACT_Total,
                            } = renderACTScores(testDoc);
                            return (
                              <TableRow key={index}>
                                <TableCell sx={{ verticalAlign: 'top' }}>
                                  {testDoc.date || 'N/A'}
                                </TableCell>
                                <TableCell sx={{ verticalAlign: 'top' }}>
                                  {(testDoc.test || 'N/A')} ({testDoc.type || 'N/A'})
                                </TableCell>
                                <TableCell sx={{ verticalAlign: 'top' }}>{English}</TableCell>
                                <TableCell sx={{ verticalAlign: 'top' }}>{MathVal}</TableCell>
                                <TableCell sx={{ verticalAlign: 'top' }}>{Reading}</TableCell>
                                <TableCell sx={{ verticalAlign: 'top' }}>{Science}</TableCell>
                                <TableCell sx={{ verticalAlign: 'top' }}>{ACT_Total}</TableCell>
                              </TableRow>
                            );
                          })
                        ) : (
                          <TableRow>
                            <TableCell colSpan={7}>
                              <Typography variant="body2" color="textSecondary">
                                No ACT tests available.
                              </Typography>
                            </TableCell>
                          </TableRow>
                        )}
                      </TableBody>
                    </Table>
                  </TableContainer>
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
