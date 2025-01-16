// src/components/BookingPage.js

import React, { useEffect, useState, useContext } from 'react';
import { API_BASE_URL } from '../config';
import { useNavigate } from 'react-router-dom';
import { AuthContext } from '../contexts/AuthContext';
import {
  Container,
  AppBar,
  Toolbar,
  Typography,
  Button,
  Box,
  Select,
  MenuItem,
  CircularProgress,
  Grid,
  Card,
  CardContent,
  Paper,
  Divider,
  Avatar,
  useMediaQuery,
  useTheme,
} from '@mui/material';
import Link from '@mui/material/Link';
import { styled } from '@mui/system';
import { tutorBookingLinks } from '../config/TutorBookingLinks';

// Tutor Images
import ben from '../assets/ben.jpg';
import edward from '../assets/edward.jpg';
import kieran from '../assets/kieran.jpg';
import kyra from '../assets/kyra.jpg';
import omar from '../assets/omar.jpg';
import patrick from '../assets/patrick.jpg';
import eli from '../assets/eli.jpg';

// ---- Brand Colors & Theme ----
const brandBlue = '#0e1027';
const brandGold = '#b29600';
const lightBackground = '#fafafa'; // Page background

// ---- Styled Components ----
const StyledAppBar = styled(AppBar)(() => ({
  backgroundColor: brandBlue,
}));

const HeroSection = styled(Box)(() => ({
  background: `linear-gradient(to bottom right, ${brandBlue} 30%, #2a2f45 90%)`,
  color: '#fff',
  borderRadius: '8px',
  padding: '40px',
  marginTop: '24px',
  marginBottom: '40px',
  boxShadow: '0 4px 20px rgba(0,0,0,0.1)',
}));

const InfoSection = styled(Paper)(() => ({
  padding: '24px',
  borderRadius: '16px',
  backgroundColor: '#fff',
  boxShadow: '0 4px 20px rgba(0,0,0,0.05)',
  marginBottom: '40px',
}));

const TutorCard = styled(Card)(() => ({
  borderRadius: '16px',
  boxShadow: '0 4px 10px rgba(0,0,0,0.1)',
  backgroundColor: '#fff',
}));

const SectionTitle = styled(Typography)(() => ({
  marginBottom: '16px',
  fontWeight: '600',
  color: brandBlue,
}));

// Map tutor names to subject descriptions:
const tutorSubjects = {
  edward: 'Math, Science, and Study Skills Tutor',
  eli: 'Reading, English/Grammar, and Science Tutor',
  ben: 'Reading, English/Grammar, and Science Tutor',
  patrick: 'Reading, English/Grammar, and Science Tutor',
  kyra: 'Math and Science Tutor',
  kieran: 'Math and Science Tutor',
  omar: 'High School / College Biology and Chemistry Tutor',
};

const tutorImages = {
  ben,
  edward,
  kieran,
  kyra,
  omar,
  patrick,
  eli,
};

const BookingPage = () => {
  const [studentsData, setStudentsData] = useState([]);
  const [selectedStudentID, setSelectedStudentID] = useState(null);
  const [parentName, setParentName] = useState('Parent');
  const [parentPicture, setParentPicture] = useState(null);
  const [loading, setLoading] = useState(true);

  const authState = useContext(AuthContext);
  const navigate = useNavigate();

  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));

  // ---- Fetch Parent Data ----
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
      });
  }, [navigate]);

  // ---- Fetch Associated Students & Their Data ----
  useEffect(() => {
    const token = localStorage.getItem('authToken');

    // Parse the studentID from query parameters
    const params = new URLSearchParams(window.location.search);
    const queriedStudentID = params.get('studentID');

    fetch(`${API_BASE_URL}/api/associated-students`, {
      method: 'GET',
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
      .then(async (response) => {
        if (!response.ok) {
          const errorText = await response.text();
          throw new Error(`Error ${response.status}: ${errorText}`);
        }
        return response.json();
      })
      .then((data) => {
        const associatedStudents = data.associatedStudents || [];
        if (associatedStudents.length > 0) {
          // Fetch student data
          fetchStudentsData(associatedStudents, token, queriedStudentID);
        } else {
          // If no students, redirect to intake
          navigate('/studentintake');
        }
      })
      .catch((error) => {
        console.error('Error fetching associated students:', error);
        setLoading(false);
      });
  }, [navigate]);

  const fetchStudentsData = (studentIDs, token, queriedStudentID) => {
    const fetchPromises = studentIDs.map((studentID) =>
      fetch(`${API_BASE_URL}/api/students/${studentID}`, {
        method: 'GET',
        headers: {
          Authorization: `Bearer ${token}`,
        },
      })
        .then(async (response) => {
          if (!response.ok) {
            const errorText = await response.text();
            throw new Error(`Error ${response.status}: ${errorText}`);
          }
          return response.json();
        })
        .then((studentData) => ({
          studentID,
          personal: studentData.personal || {},
          business: studentData.business || {},
        }))
    );

    Promise.all(fetchPromises)
      .then((students) => {
        setStudentsData(students);
        // If queriedStudentID is provided and valid, select it. Otherwise, select the first.
        if (queriedStudentID && students.some((s) => s.studentID === queriedStudentID)) {
          setSelectedStudentID(queriedStudentID);
        } else {
          setSelectedStudentID(students[0]?.studentID);
        }
        setLoading(false);
      })
      .catch((error) => {
        console.error('Error fetching student data:', error);
        setLoading(false);
      });
  };

  // ---- Handlers ----
  const handleStudentChange = (event) => {
    const newStudentID = event.target.value;
    setSelectedStudentID(newStudentID);
    navigate(`/booking?studentID=${newStudentID}`);
  };

  const handleBackToDashboard = () => {
    if (selectedStudentID) {
      navigate(`/parentdashboard?studentID=${selectedStudentID}`);
    } else {
      navigate('/parentdashboard');
    }
  };

  const handleSignOut = () => {
    localStorage.removeItem('authToken');
    authState.updateToken(null);
    navigate('/');
  };

  if (loading) {
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

  // ---- Data Access ----
  const selectedStudent = studentsData.find(
    (student) => student.studentID === selectedStudentID
  );
  const tutors = selectedStudent?.business.associated_tutors || [];

  // Example hour calculations (replace with your actual logic/data)
  const totalHoursPurchased = selectedStudent?.business.totalHoursPurchased || 0;
  const totalHoursUsed = selectedStudent?.business.totalHoursUsed || 0;
  const totalHoursRemaining = totalHoursPurchased - totalHoursUsed;

  return (
    <Box sx={{ backgroundColor: lightBackground, minHeight: '100vh' }}>
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
              {/* Top row: Welcome + Sign Out */}
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
                <Typography variant="h5" sx={{ fontWeight: 'bold', whiteSpace: 'nowrap' }}>
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
          {/* Single row: "Booking for:" + dropdown */}
          <Box
            display="flex"
            alignItems="center"
            flexWrap="nowrap"
            gap={2}
            sx={{
              whiteSpace: 'nowrap',
            }}
          >
            <Typography
              variant={isMobile ? 'h6' : 'h5'}
              sx={{ fontWeight: 700, m: 0 }}
            >
              Booking for:
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
              {studentsData.map((student) => (
                <MenuItem key={student.studentID} value={student.studentID}>
                  {student.personal.name || 'Unnamed Student'}
                </MenuItem>
              ))}
            </Select>
          </Box>

          <Box mt={2}>
            <Typography
              variant={isMobile ? 'body1' : 'h6'}
              sx={{ opacity: 0.9 }}
            >
              Book your student's next appointment below.
            </Typography>
          </Box>
        </HeroSection>
      </Container>

      {/* ---------- Tutors & Hours Section ---------- */}
      <Container maxWidth="xl">
        <InfoSection>
          <Box
            display="flex"
            justifyContent="space-between"
            alignItems="center"
            sx={{ marginBottom: '16px' }}
          >
            {/* Updated Title: "so-and-so's Tutors:" */}
            <SectionTitle variant="h5" sx={{ m: 0 }}>
              {selectedStudent?.personal.name
                ? `${selectedStudent.personal.name}'s Tutors:`
                : 'Student’s Tutors:'}
            </SectionTitle>

            <Button
              onClick={handleBackToDashboard}
              sx={{
                textTransform: 'none',
                fontWeight: 'bold',
                color: brandBlue,
                border: `1px solid ${brandBlue}`,
                '&:hover': {
                  backgroundColor: brandBlue,
                  color: '#fff',
                },
              }}
            >
              Back to Dashboard
            </Button>
          </Box>

          <Divider sx={{ marginBottom: '24px' }} />

          {/*
            We set order={{ xs: 1, md: 2 }} on the hours breakdown so it appears FIRST on mobile,
            and order={{ xs: 2, md: 1 }} on the tutors so it appears SECOND on mobile.
            On desktop (md), they switch to their original positions.
          */}
          <Grid container spacing={4} alignItems="stretch">
            {/* ---------- Right Column: Hours Breakdown (FIRST on mobile) ---------- */}
            <Grid
              item
              xs={12}
              md={6}
              order={{ xs: 1, md: 2 }}
              sx={{
                display: 'flex',
                flexDirection: 'column',
              }}
            >
              <Box
                sx={{
                  border: `1px solid #ddd`,
                  borderRadius: '8px',
                  p: 3,
                  display: 'flex',
                  flexDirection: 'column',
                  justifyContent: 'space-between',
                  height: '100%',
                }}
              >
                <Box>
                  <Typography variant="h6" sx={{ fontWeight: 700, mb: 2 }}>
                    Hours Breakdown
                  </Typography>
                  <Typography variant="body1" sx={{ mb: 1 }}>
                    <strong>Total Hours Purchased:</strong> {totalHoursPurchased}
                  </Typography>
                  <Typography variant="body1" sx={{ mb: 1 }}>
                    <strong>Total Hours Used:</strong> {totalHoursUsed}
                  </Typography>
                  <Typography variant="body1">
                    <strong>Total Hours Remaining:</strong> {totalHoursRemaining}
                  </Typography>
                </Box>

                <Box sx={{ mt: 2 }}>
                  <Typography variant="body2" sx={{ fontStyle: 'italic' }}>
                  Please contact{' '}
                  <Link href="mailto:admin@leetutoring.com" underline="hover">
                    admin@leetutoring.com
                  </Link>{' '}
                  to purchase more hours if needed.
                </Typography>
                </Box>
              </Box>
            </Grid>

            {/* ---------- Left Column: Tutor Cards (SECOND on mobile) ---------- */}
            <Grid
              item
              xs={12}
              md={6}
              order={{ xs: 2, md: 1 }}
              sx={{
                display: 'flex',
                flexDirection: 'column',
              }}
            >
              {tutors.length > 0 ? (
                tutors.map((tutor, index) => {
                  const tutorKey = tutor.toLowerCase();
                  const tutorImage =
                    tutorImages[tutorKey] ||
                    'https://via.placeholder.com/300x200?text=Tutor+Image';
                  const tutorInfo = tutorBookingLinks[tutor] || {};
                  const subjectDescription = tutorSubjects[tutorKey] || 'Tutor';

                  return (
                    <Box key={index} mb={3}>
                      <TutorCard>
                        <CardContent>
                          <Grid container spacing={2} alignItems="stretch">
                            {/* Tutor Image */}
                            <Grid item xs={12} sm={4}>
                              <Box
                                sx={{
                                  width: '100%',
                                  height: '100%',
                                  borderRadius: '8px',
                                  overflow: 'hidden',
                                }}
                              >
                                <Box
                                  component="img"
                                  src={tutorImage}
                                  alt={tutor}
                                  sx={{
                                    width: '100%',
                                    height: '100%',
                                    objectFit: 'cover',
                                    objectPosition: 'center',
                                  }}
                                />
                              </Box>
                            </Grid>

                            {/* Tutor Info */}
                            <Grid
                              item
                              xs={12}
                              sm={8}
                              display="flex"
                              flexDirection="column"
                            >
                              <Typography variant="h6" sx={{ fontWeight: 600, mb: 1 }}>
                                {tutor}
                              </Typography>

                              <Typography variant="body1" sx={{ mb: 2 }}>
                                {subjectDescription}
                              </Typography>

                              {tutorInfo.individualLink ? (
                                <Button
                                  variant="contained"
                                  href={tutorInfo.individualLink}
                                  target="_blank"
                                  rel="noopener noreferrer"
                                  sx={{
                                    alignSelf: 'flex-start',
                                    textTransform: 'none',
                                    fontWeight: 'bold',
                                    backgroundColor: brandBlue,
                                    color: '#fff',
                                    '&:hover': {
                                      backgroundColor: '#1c2231',
                                    },
                                  }}
                                >
                                  Book Session
                                </Button>
                              ) : (
                                <Typography variant="body2" color="textSecondary">
                                  No booking link available.
                                </Typography>
                              )}
                            </Grid>
                          </Grid>
                        </CardContent>
                      </TutorCard>
                    </Box>
                  );
                })
              ) : (
                <Typography variant="body1" color="textSecondary">
                  No tutors available.
                </Typography>
              )}
            </Grid>
          </Grid>
        </InfoSection>
      </Container>
    </Box>
  );
};

export default BookingPage;
