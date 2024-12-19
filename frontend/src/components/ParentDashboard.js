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
  List,
  ListItem,
  ListItemText,
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
} from '@mui/material';
import { styled } from '@mui/system';

const navy = '#001F54';
const cream = '#FFF8E1';
const backgroundGray = '#f9f9f9';
const lightGray = '#f0f0f0';

const StyledAppBar = styled(AppBar)({
  backgroundColor: navy,
});

const HeroSection = styled(Box)({
  borderRadius: '8px',
  padding: '40px',
  marginTop: '24px',
  marginBottom: '24px',
  color: '#fff',
  background: `linear-gradient(to bottom right, ${navy}, #003f88)`,
  boxShadow: '0 4px 20px rgba(0,0,0,0.1)'
});

const ContentWrapper = styled(Box)({
  backgroundColor: backgroundGray,
  borderRadius: '16px',
  padding: '24px',
  marginBottom: '40px',
  boxShadow: '0 6px 20px rgba(0,0,0,0.1)',
});

const SectionContainer = styled(Paper)({
  padding: '24px',
  borderRadius: '16px',
  backgroundColor: '#fff',
  boxShadow: '0 4px 20px rgba(0,0,0,0.05)',
});

const SectionTitle = styled(Typography)({
  marginBottom: '16px',
  fontWeight: 600,
  color: navy,
});

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
  const [associatedStudents, setAssociatedStudents] = useState([]);
  const [selectedStudentID, setSelectedStudentID] = useState(null);
  const [studentData, setStudentData] = useState(null);
  const [parentName, setParentName] = useState('Parent');
  const [parentPicture, setParentPicture] = useState(null);
  const [activeTab, setActiveTab] = useState(0);
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();

  // Fetch Parent Data
  useEffect(() => {
    const token = localStorage.getItem('authToken');
    if (!token) {
      // Handle unauthenticated state
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
        // Optionally, set default values or handle error state
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
        // Parse the query parameter
        const params = new URLSearchParams(window.location.search);
        const queriedStudentID = params.get('studentID');

        if (queriedStudentID && data.associatedStudents.includes(queriedStudentID)) {
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
          console.log('Fetched Student Data:', data); // Debugging line
          setStudentData(data);
          setLoading(false);
        })
        .catch((err) => {
          console.error(err);
          setLoading(false);
        });
    }
  }, [selectedStudentID]);

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

  const handleTabChange = (event, newValue) => {
    setActiveTab(newValue);
  };

  // Navigation to Booking Page with current studentID
  const handleNavigateToBooking = () => {
    if (selectedStudentID) {
      navigate(`/booking?studentID=${selectedStudentID}`);
    } else {
      navigate('/booking');
    }
  };

  if (loading || !studentData) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" height="100vh">
        <CircularProgress />
      </Box>
    );
  }

  // Filter and sort test dates for upcoming tests
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
    .sort((a, b) => {
      const dateA = new Date(a.test_date);
      const dateB = new Date(b.test_date);
      return dateA - dateB;
    });

  const testData = studentData.testData || [];

  // Separate SAT/PSAT and ACT tests
  const satTests = testData.filter((t) => {
    const upperTest = (t.test || '').toUpperCase();
    return upperTest.includes('SAT') || upperTest.includes('PSAT');
  });

  const actTests = testData.filter((t) => {
    const upperTest = (t.test || '').toUpperCase();
    return upperTest.includes('ACT');
  });

  const tableCellProps = {
    sx: { verticalAlign: 'top' },
  };

  const renderSATScores = (testDoc) => {
    let EBRW = '';
    let Math = '';
    let Reading = '';
    let Writing = '';
    let SAT_Total = '';

    if (Array.isArray(testDoc.SAT_Scores) && testDoc.SAT_Scores.length > 0) {
      EBRW = testDoc.SAT_Scores[0] || '';
      Math = testDoc.SAT_Scores[1] || '';
      Reading = testDoc.SAT_Scores[2] || '';
      Writing = testDoc.SAT_Scores[3] || '';
      SAT_Total = testDoc.SAT_Scores[4] || '';
    } else if (Array.isArray(testDoc.SAT) && testDoc.SAT.length > 0) {
      EBRW = testDoc.SAT[0] || '';
      Math = testDoc.SAT[1] || '';
      Reading = testDoc.SAT[2] || '';
      Writing = testDoc.SAT[3] || '';
      SAT_Total = testDoc.SAT[4] || '';
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
      English = testDoc.ACT_Scores[0] || '';
      MathVal = testDoc.ACT_Scores[1] || '';
      Reading = testDoc.ACT_Scores[2] || '';
      Science = testDoc.ACT_Scores[3] || '';
      ACT_Total = testDoc.ACT_Scores[4] || '';
    } else if (Array.isArray(testDoc.ACT) && testDoc.ACT.length > 0) {
      English = testDoc.ACT[0] || '';
      MathVal = testDoc.ACT[1] || '';
      Reading = testDoc.ACT[2] || '';
      Science = testDoc.ACT[3] || '';
      ACT_Total = testDoc.ACT[4] || '';
    }

    return { English, Math: MathVal, Reading, Science, ACT_Total };
  };

  return (
    <Box sx={{ backgroundColor: '#fafafa', minHeight: '100vh' }}>
      <StyledAppBar position="static">
        <Toolbar sx={{ display: 'flex', justifyContent: 'space-between' }}>
          {/* Left Side: Welcome Message and Avatar */}
          <Box display="flex" alignItems="center" gap="16px">
            <Typography variant="h6" sx={{ fontWeight: 'bold' }}>
              Welcome, {parentName}!
            </Typography>
            <Avatar
              src={parentPicture || undefined}
              alt={parentName}
              sx={{ bgcolor: parentPicture ? 'transparent' : '#003f88', color: '#fff' }}
            >
              {!parentPicture && parentName.charAt(0).toUpperCase()}
            </Avatar>
          </Box>

          {/* Right Side: Navigation Buttons */}
          <Box display="flex" alignItems="center" gap="16px">
            <Button
              color="inherit"
              onClick={handleNavigateToBooking}
              sx={{
                textTransform: 'none',
                fontWeight: 'bold',
              }}
            >
              Book With A Tutor
            </Button>
            <Button
              color="inherit"
              onClick={handleSignOut}
              sx={{ textTransform: 'none', fontWeight: 'bold' }}
            >
              Sign Out
            </Button>
          </Box>
        </Toolbar>
      </StyledAppBar>

      <Container maxWidth="xl">
        <HeroSection>
          <Box display="flex" alignItems="center" justifyContent="space-between" flexWrap="wrap">
            <Box>
              <Typography variant="h3" sx={{ fontWeight: 700, marginBottom: '8px' }}>
                {studentData.personal?.name || 'Student'}
              </Typography>
              <Typography variant="h6" sx={{ opacity: 0.9 }}>
                Manage performance, view upcoming tests, track homework completion, and set goals.
              </Typography>
            </Box>

            <Box mt={{ xs: 2, md: 0 }}>
              <Select
                value={selectedStudentID}
                onChange={handleStudentChange}
                variant="outlined"
                sx={{
                  minWidth: '200px',
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

        <ContentWrapper>
          <Tabs
            value={activeTab}
            onChange={handleTabChange}
            textColor="primary"
            indicatorColor="primary"
            variant="scrollable"
            scrollButtons="auto"
            sx={{ marginBottom: '16px' }}
          >
            <Tab label="Overview" sx={{ textTransform: 'none', fontWeight: 'bold' }} />
            <Tab label="School Goals" sx={{ textTransform: 'none', fontWeight: 'bold' }} />
            <Tab label="Student Profile" sx={{ textTransform: 'none', fontWeight: 'bold' }} />
          </Tabs>

          {/* OVERVIEW TAB */}
          <TabPanel value={activeTab} index={0}>
            <Grid container spacing={4}>
              {/* Upcoming Test Dates */}
              <Grid item xs={12} md={2}>
                <SectionContainer>
                  <SectionTitle variant="h6">Testing Dates</SectionTitle>
                  <Divider sx={{ marginBottom: '16px' }} />
                  <List>
                    {testDates.length > 0 ? (
                      testDates.map((test, index) => (
                        <ListItem key={index} sx={{ paddingLeft: 0 }}>
                          <ListItemText
                            primary={test.test_date || 'N/A'}
                            secondary={test.test_type || 'N/A'}
                          />
                        </ListItem>
                      ))
                    ) : (
                      <Typography variant="body2" color="textSecondary">
                        No upcoming tests.
                      </Typography>
                    )}
                  </List>
                </SectionContainer>
              </Grid>

              {/* Test Scores */}
              <Grid item xs={12} md={8}>
                <SectionContainer>
                  <SectionTitle variant="h6">Test Scores</SectionTitle>
                  <Divider sx={{ marginBottom: '16px' }} />

                  {/* SAT/PSAT Scores Table */}
                  <Typography variant="h6" sx={{ fontWeight: 600, marginTop: '16px' }}>
                    SAT/PSAT Scores
                  </Typography>
                  <TableContainer sx={{ marginBottom: '24px' }}>
                    <Table>
                      <TableHead>
                        <TableRow>
                          <TableCell sx={{ fontWeight: 600 }} {...tableCellProps}>Date</TableCell>
                          <TableCell sx={{ fontWeight: 600 }} {...tableCellProps}>Test (Type)</TableCell>
                          <TableCell sx={{ fontWeight: 600 }} {...tableCellProps}>EBRW</TableCell>
                          <TableCell sx={{ fontWeight: 600 }} {...tableCellProps}>Math</TableCell>
                          <TableCell sx={{ fontWeight: 600 }} {...tableCellProps}>Reading</TableCell>
                          <TableCell sx={{ fontWeight: 600 }} {...tableCellProps}>Writing</TableCell>
                          <TableCell sx={{ fontWeight: 600 }} {...tableCellProps}>SAT_Total</TableCell>
                        </TableRow>
                      </TableHead>
                      <TableBody>
                        {satTests.length > 0 ? (
                          satTests.map((testDoc, index) => {
                            const scores = renderSATScores(testDoc);
                            return (
                              <TableRow key={index}>
                                <TableCell {...tableCellProps}>{testDoc.date || 'N/A'}</TableCell>
                                <TableCell {...tableCellProps}>
                                  {(testDoc.test || 'N/A')} ({testDoc.type || 'N/A'})
                                </TableCell>
                                <TableCell {...tableCellProps}>{scores.EBRW}</TableCell>
                                <TableCell {...tableCellProps}>{scores.Math}</TableCell>
                                <TableCell {...tableCellProps}>{scores.Reading}</TableCell>
                                <TableCell {...tableCellProps}>{scores.Writing}</TableCell>
                                <TableCell {...tableCellProps}>{scores.SAT_Total}</TableCell>
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

                  {/* ACT Scores Table */}
                  <Typography variant="h6" sx={{ fontWeight: 600 }}>
                    ACT Scores
                  </Typography>
                  <TableContainer>
                    <Table>
                      <TableHead>
                        <TableRow>
                          <TableCell sx={{ fontWeight: 600 }} {...tableCellProps}>Date</TableCell>
                          <TableCell sx={{ fontWeight: 600 }} {...tableCellProps}>Test (Type)</TableCell>
                          <TableCell sx={{ fontWeight: 600 }} {...tableCellProps}>English</TableCell>
                          <TableCell sx={{ fontWeight: 600 }} {...tableCellProps}>Math</TableCell>
                          <TableCell sx={{ fontWeight: 600 }} {...tableCellProps}>Reading</TableCell>
                          <TableCell sx={{ fontWeight: 600 }} {...tableCellProps}>Science</TableCell>
                          <TableCell sx={{ fontWeight: 600 }} {...tableCellProps}>ACT_Total</TableCell>
                        </TableRow>
                      </TableHead>
                      <TableBody>
                        {actTests.length > 0 ? (
                          actTests.map((testDoc, index) => {
                            const scores = renderACTScores(testDoc);
                            return (
                              <TableRow key={index}>
                                <TableCell {...tableCellProps}>{testDoc.date || 'N/A'}</TableCell>
                                <TableCell {...tableCellProps}>
                                  {(testDoc.test || 'N/A')} ({testDoc.type || 'N/A'})
                                </TableCell>
                                <TableCell {...tableCellProps}>{scores.English}</TableCell>
                                <TableCell {...tableCellProps}>{scores.Math}</TableCell>
                                <TableCell {...tableCellProps}>{scores.Reading}</TableCell>
                                <TableCell {...tableCellProps}>{scores.Science}</TableCell>
                                <TableCell {...tableCellProps}>{scores.ACT_Total}</TableCell>
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

              {/* Homework Completion */}
              <Grid item xs={12} md={2}>
              <SectionContainer>
                <SectionTitle variant="h6">Homework Completion</SectionTitle>
                <Divider sx={{ marginBottom: '16px' }} />
                <List>
                  {(() => {
                    // Clone the homeworkCompletion array to avoid mutating the original state
                    const sortedHomework = [...(studentData.homeworkCompletion || [])].sort((a, b) => {
                      const dateA = a.date ? new Date(a.date) : new Date(0); // Assign epoch if date is missing
                      const dateB = b.date ? new Date(b.date) : new Date(0);
                      return dateB - dateA; // Descending order
                    });

                    return sortedHomework.length > 0 ? (
                      sortedHomework.map((hw, index) => {
                        // Parse the date string into a Date object
                        const parsedDate = hw.date ? new Date(hw.date) : null;

                        // Format the date for better readability
                        const formattedDate = parsedDate
                          ? parsedDate.toLocaleDateString(undefined, {
                              year: 'numeric',
                              month: 'long',
                              day: 'numeric',
                            })
                          : 'N/A';

                        // Validate percentage
                        const percentage = hw.percentage ? hw.percentage : '0%';

                        return (
                          <ListItem key={index} sx={{ paddingLeft: 0 }}>
                            <ListItemText
                              primary={formattedDate}
                              secondary={`${percentage}`}
                            />
                          </ListItem>
                        );
                      })
                    ) : (
                      <Typography variant="body2" color="textSecondary">
                        No homework completion records available.
                      </Typography>
                    );
                  })()}
                </List>
              </SectionContainer>
              </Grid>
            </Grid>
          </TabPanel>
          
          
          {/* SCHOOL GOALS TAB */}
          <TabPanel value={activeTab} index={1}>
            <SectionContainer>
              <SectionTitle variant="h6">School Goals</SectionTitle>
              <Divider sx={{ marginBottom: '16px' }} />
              <List>
                {(studentData.goals || []).length > 0 ? (
                  studentData.goals.map((goal, index) => {
                    // Initialize an array to hold the percentiles
                    const percentiles = [];

                    // Check if ACT_percentiles exists and is not 'N/A'
                    if (goal.ACT_percentiles && goal.ACT_percentiles !== 'N/A') {
                      percentiles.push(`ACT: ${goal.ACT_percentiles}`);
                    }

                    // Check if SAT_percentiles exists and is not 'N/A'
                    if (goal.SAT_percentiles && goal.SAT_percentiles !== 'N/A') {
                      percentiles.push(`SAT: ${goal.SAT_percentiles}`);
                    }

                    // Combine the percentiles into a single string separated by commas
                    const secondaryText = percentiles.join(', ');

                    return (
                      <ListItem key={index} sx={{ paddingLeft: 0 }}>
                        <ListItemText
                          primary={goal.College || 'N/A'}
                          secondary={secondaryText || 'No percentiles available.'}
                        />
                      </ListItem>
                    );
                  })
                ) : (
                  <Typography variant="body2" color="textSecondary">
                    No school goals available.
                  </Typography>
                )}
              </List>
            </SectionContainer>
          </TabPanel>




          {/* STUDENT PROFILE TAB */}
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
        </ContentWrapper>
      </Container>
    </Box>
  );
};

export default ParentDashboard;
