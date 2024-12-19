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
} from '@mui/material';
import { styled } from '@mui/system';
import { tutorBookingLinks } from '../config/TutorBookingLinks';

import ben from '../assets/ben.jpg';
import edward from '../assets/edward.jpg';
import kieran from '../assets/kieran.jpg';
import kyra from '../assets/kyra.jpg';
import omar from '../assets/omar.jpg';
import patrick from '../assets/patrick.jpg';
import eli from '../assets/eli.jpg';

const navy = '#001F54';
const cream = '#FFF8E1';
const backgroundGray = '#f9f9f9';

// Styled Components
const StyledAppBar = styled(AppBar)({
  backgroundColor: navy,
});

const HeroSection = styled(Box)({
  background: `linear-gradient(to bottom right, ${navy} 40%, ${cream} 100%)`,
  color: '#fff',
  borderRadius: '8px',
  padding: '40px',
  marginTop: '24px',
  marginBottom: '40px',
  boxShadow: '0 4px 20px rgba(0,0,0,0.1)',
});

const InfoSection = styled(Paper)({
  padding: '24px',
  borderRadius: '16px',
  backgroundColor: '#fff',
  boxShadow: '0 4px 20px rgba(0,0,0,0.05)',
  marginBottom: '40px',
});

const TutorCard = styled(Card)({
  borderRadius: '16px',
  boxShadow: '0 4px 10px rgba(0,0,0,0.1)',
  backgroundColor: '#fff',
});

const SectionTitle = styled(Typography)({
  marginBottom: '16px',
  fontWeight: '600',
  color: navy,
});

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

    // Parse the studentID from the query parameters
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
          fetchStudentsData(associatedStudents, token, queriedStudentID);
        } else {
          // If no associated students, redirect to intake
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
        // If queriedStudentID is provided and it's in the list, select it. Otherwise, default to the first.
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
        sx={{ backgroundColor: backgroundGray }}
      >
        <CircularProgress />
      </Box>
    );
  }

  const selectedStudent = studentsData.find(
    (student) => student.studentID === selectedStudentID
  );

  const tutors = selectedStudent?.business.associated_tutors || [];

  return (
    <Box sx={{ backgroundColor: backgroundGray, minHeight: '100vh' }}>
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
              onClick={handleBackToDashboard}
              sx={{
                textTransform: 'none',
                fontWeight: 'bold',
              }}
            >
              Back to Dashboard
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

      <Container maxWidth="xl" sx={{ marginTop: '24px' }}>
        {/* Hero Section */}
        <HeroSection>
          <Box display="flex" alignItems="center" justifyContent="space-between" flexWrap="wrap">
            <Box>
              <Typography variant="h3" component="div" sx={{ fontWeight: 700, color: '#fff' }}>
                Booking for {selectedStudent?.personal.name || 'Unnamed Student'}
              </Typography>
              <Typography variant="h6" sx={{ color: '#fff', opacity: 0.9 }}>
                Manage tutoring sessions and view recent feedback from your student’s tutors.
              </Typography>
            </Box>

            <Box mt={{ xs: 2, md: 0 }}>
              <Select
                value={selectedStudentID}
                onChange={handleStudentChange}
                variant="outlined"
                sx={{
                  minWidth: '240px',
                  backgroundColor: '#fff',
                  borderRadius: '8px',
                  fontWeight: 500,
                }}
              >
                {studentsData.map((student) => (
                  <MenuItem key={student.studentID} value={student.studentID}>
                    {student.personal.name || 'Unnamed Student'}
                  </MenuItem>
                ))}
              </Select>
            </Box>
          </Box>
        </HeroSection>

        {/* Combined Feedback & Booking Section */}
        <InfoSection>
          <SectionTitle variant="h5">
            Feedback {selectedStudent?.personal.name ? `From ${selectedStudent.personal.name}'s Tutors:` : 'for Tutors:'}
          </SectionTitle>
          <Divider sx={{ marginBottom: '24px' }} />

          {tutors.length > 0 ? (
            <Grid container spacing={4}>
              {tutors.map((tutor, index) => {
                const tutorInfo = tutorBookingLinks[tutor] || {};
                const tutorKey = tutor.toLowerCase();
                const tutorImage =
                  tutorImages[tutorKey] || 'https://via.placeholder.com/300x200?text=Tutor+Image';

                return (
                  <Grid item xs={12} key={index}>
                    <TutorCard>
                      <CardContent>
                        <Grid container spacing={2} alignItems="stretch">
                          {/* Tutor Image */}
                          <Grid item xs={12} md={3}>
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

                          {/* Feedback & Booking */}
                          <Grid item xs={12} md={9} display="flex" flexDirection="column">
                            <Typography variant="h6" sx={{ fontWeight: 600, marginBottom: 2 }}>
                              An Update From {tutor}:
                            </Typography>

                            <Typography variant="body1" sx={{ flexGrow: 1 }}>
                              {tutorInfo.feedback ||
                                'No recent feedback available.'}
                            </Typography>

                            <Box display="flex" justifyContent="space-between" alignItems="center" sx={{ marginTop: '16px' }}>
                              <Typography variant="body2" color="textSecondary">
                                {tutorInfo.feedbackDate || 'Feedback Date: N/A'}
                              </Typography>

                              {tutorInfo.individualLink ? (
                                <Button
                                  variant="contained"
                                  color="primary"
                                  href={tutorInfo.individualLink}
                                  target="_blank"
                                  rel="noopener noreferrer"
                                  sx={{
                                    textTransform: 'none',
                                    fontWeight: 'bold',
                                    backgroundColor: navy,
                                    ':hover': { backgroundColor: '#002e7a' },
                                  }}
                                >
                                  Book Session
                                </Button>
                              ) : (
                                <Typography variant="body2" color="textSecondary">
                                  No booking link available.
                                </Typography>
                              )}
                            </Box>
                          </Grid>
                        </Grid>
                      </CardContent>
                    </TutorCard>
                  </Grid>
                );
              })}
            </Grid>
          ) : (
            <Typography variant="body1" color="textSecondary">
              No tutors available.
            </Typography>
          )}
        </InfoSection>
      </Container>
    </Box>
  );
};

export default BookingPage;
