// src/components/StudentIntake.js

import React, { useState, useEffect } from 'react';
import { API_BASE_URL } from '../config';
import { useNavigate } from 'react-router-dom';

// MUI imports
import {
  Box,
  Card,
  CardContent,
  Typography,
  Button,
  TextField,
  CircularProgress,
  Divider,
  List,
  ListItem,
  ListItemText,
  Stack,
  useMediaQuery,
  useTheme,
} from '@mui/material';
import { styled } from '@mui/system';

// Background image
import loginImage from '../assets/login.jpg';

// Brand colors
const brandBlue = '#0e1027';
const brandGold = '#b29600';
const brandGoldLight = '#d4a100';

/* ===================== STYLED COMPONENTS ===================== */

/**
 * Root container:
 *  - Desktop: horizontal layout: left image (60%) + right form (40%)
 *  - Mobile: single-column with full background.
 *  - Lock scrolling and remove default margins/padding to fill screen fully.
 */
const RootContainer = styled(Box)(({ theme }) => ({
  display: 'flex',
  width: '100vw',
  height: '100vh',
  margin: 0,
  padding: 0,
  overflow: 'hidden', // lock scrolling

  [theme.breakpoints.down('md')]: {
    // On mobile, use the background image behind the entire screen
    backgroundImage: `url(${loginImage})`,
    backgroundSize: 'cover',
    backgroundPosition: 'center',
    backgroundRepeat: 'no-repeat',
    justifyContent: 'center',
    alignItems: 'center',
  },
}));

/** 
 * Left-side image container for desktop 
 * Hidden on mobile (md breakpoint and below).
 */
const ImageContainer = styled(Box)(({ theme }) => ({
  width: '60%',
  height: '100%',
  position: 'relative',
  backgroundImage: `url(${loginImage})`,
  backgroundSize: 'cover',
  backgroundPosition: 'center',
  backgroundRepeat: 'no-repeat',

  [theme.breakpoints.down('md')]: {
    display: 'none',
  },
}));

/** 
 * A semi-transparent overlay on the desktop image 
 */
const ImageOverlay = styled(Box)(() => ({
  position: 'absolute',
  inset: 0,
  backgroundColor: `${brandBlue}80`, // brandBlue at 50% opacity
}));

/** 
 * Right-side (desktop) or center (mobile) container 
 * for the actual form content (Card).
 */
const IntakeContainer = styled(Box)(({ theme }) => ({
  width: '40%',
  height: '100%',
  backgroundColor: '#f9f9f9',
  display: 'flex',
  justifyContent: 'center',
  alignItems: 'center',
  padding: theme.spacing(4),

  [theme.breakpoints.down('md')]: {
    width: '100%',
    height: 'auto',
    backgroundColor: 'transparent',
    padding: theme.spacing(2, 2, 6),
  },
}));

/** The card that holds the intake steps */
const StyledCard = styled(Card)(({ theme }) => ({
  width: '100%',
  maxWidth: 500,
  margin: 'auto',
  boxShadow: '0 4px 20px rgba(0, 0, 0, 0.15)',
  borderRadius: '16px',
  backgroundColor: '#fff',
}));

/** A brand-styled button */
const StyledButton = styled(Button, {
  shouldForwardProp: (prop) => prop !== 'hovercolor',
})(({ hovercolor }) => ({
  textTransform: 'none',
  fontWeight: 'bold',
  backgroundColor: brandBlue,
  color: '#fff',
  '&:hover': {
    backgroundColor: hovercolor || brandGold,
  },
}));

/* ===================== MAIN COMPONENT ===================== */

const StudentIntake = () => {
  const [numStudents, setNumStudents] = useState(1);
  const [studentIDs, setStudentIDs] = useState(['']);
  const [studentInfos, setStudentInfos] = useState([]);
  const [confirmationStep, setConfirmationStep] = useState(false);
  const [loading, setLoading] = useState(false);
  const [readyToProceed, setReadyToProceed] = useState(false);

  const navigate = useNavigate();
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));

  // Lock scrolling on component mount, restore on unmount
  useEffect(() => {
    document.body.style.overflow = 'hidden';
    return () => {
      document.body.style.overflow = 'auto';
    };
  }, []);

  /* -------------------- HANDLERS -------------------- */
  const handleNumStudentsChange = (e) => {
    const count = parseInt(e.target.value, 10) || 1;
    setNumStudents(count);
    setStudentIDs(Array(count).fill(''));
  };

  const handleStudentIDChange = (index, value) => {
    const newIDs = [...studentIDs];
    newIDs[index] = value;
    setStudentIDs(newIDs);
  };

  const handleSubmitStudentIDs = () => {
    setLoading(true);
    const token = localStorage.getItem('authToken');

    fetch(`${API_BASE_URL}/api/submitStudentIDs`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({ studentIds: studentIDs }),
    })
      .then(async (response) => {
        if (!response.ok) {
          const errorText = await response.text();
          throw new Error(`Error ${response.status}: ${errorText}`);
        }
        return response.json();
      })
      .then((data) => {
        console.log('Student Infos:', data);
        // Initialize 'confirmed' property to null
        const updatedInfos = data.studentInfos.map((info) => ({
          ...info,
          confirmed: null,
        }));
        setStudentInfos(updatedInfos);
        setConfirmationStep(true);
        setLoading(false);
      })
      .catch((error) => {
        console.error('Error submitting student IDs:', error);
        setLoading(false);
      });
  };

  const handleConfirmation = (index, confirmedValue) => {
    const updatedInfos = [...studentInfos];
    updatedInfos[index].confirmed = confirmedValue;
    setStudentInfos(updatedInfos);
  };

  useEffect(() => {
    // Check if all students have been confirmed or rejected
    if (confirmationStep) {
      const allConfirmed = studentInfos.every((info) => info.confirmed !== null);
      setReadyToProceed(allConfirmed);
    }
  }, [studentInfos, confirmationStep]);

  const handleProceed = () => {
    setLoading(true);
    const token = localStorage.getItem('authToken');

    // Extract confirmed student IDs
    const confirmedStudentIds = studentInfos
      .filter((info) => info.canLink && info.confirmed === true)
      .map((info) => info.studentId);

    // Proceed to link confirmed students
    fetch(`${API_BASE_URL}/api/confirmLinkStudents`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({ confirmedStudentIds }),
    })
      .then(async (response) => {
        if (!response.ok) {
          const errorText = await response.text();
          throw new Error(`Error ${response.status}: ${errorText}`);
        }
        return response.json();
      })
      .then((data) => {
        console.log('Students linked successfully:', data);
        setLoading(false);

        // Check if there are any students that were not confirmed
        const unconfirmedStudents = studentInfos.filter(
          (info) =>
            info.confirmed === false || (!info.canLink && info.confirmed !== true)
        );

        if (unconfirmedStudents.length > 0) {
          // Reset the intake form for unconfirmed students
          setNumStudents(unconfirmedStudents.length);
          setStudentIDs(Array(unconfirmedStudents.length).fill(''));
          setStudentInfos([]);
          setConfirmationStep(false);
          setReadyToProceed(false);
          alert('Some students were not confirmed. Please re-enter their IDs.');
        } else {
          // Redirect to Parent Dashboard
          navigate('/parentdashboard');
        }
      })
      .catch((error) => {
        console.error('Error confirming student links:', error);
        setLoading(false);
      });
  };

  /* -------------------- RENDER -------------------- */

  return (
    <RootContainer>
      {/* Desktop Image Section */}
      {!isMobile && (
        <ImageContainer>
          <ImageOverlay />
        </ImageContainer>
      )}

      <IntakeContainer>
        <StyledCard>
          <CardContent sx={{ p: isMobile ? 2 : 4 }}>
            {/* Heading & Instructions */}
            <Typography
              variant={isMobile ? 'h5' : 'h4'}
              sx={{ fontWeight: 'bold', color: brandBlue, mb: 1 }}
              align="center"
            >
              Let's get you linked to your student(s)!
            </Typography>

            <Divider sx={{ my: 2 }} />

            <Typography variant="body1" align="center" sx={{ mb: 3 }}>
              Please refer to the email from{' '}
              <strong style={{ color: brandGold }}>admin@leetutoring.com</strong>{' '}
              for your student ID(s). Copy/paste them below to link your student(s).
            </Typography>

            {/* If loading, show spinner */}
            {loading && (
              <Box display="flex" justifyContent="center" alignItems="center" mb={2}>
                <CircularProgress />
              </Box>
            )}

            {/* STEP 1: Enter # of students and IDs */}
            {!confirmationStep && (
              <Stack spacing={2}>
                <TextField
                  type="number"
                  label="Number of Students"
                  variant="outlined"
                  fullWidth
                  value={numStudents}
                  onChange={handleNumStudentsChange}
                  inputProps={{ min: 1 }}
                />

                {studentIDs.map((id, index) => (
                  <TextField
                    key={index}
                    label={`Student ID ${index + 1}`}
                    variant="outlined"
                    fullWidth
                    value={id}
                    onChange={(e) => handleStudentIDChange(index, e.target.value)}
                  />
                ))}

                <StyledButton
                  onClick={handleSubmitStudentIDs}
                  disabled={loading}
                  sx={{ mt: 1 }}
                >
                  Submit Student IDs
                </StyledButton>
              </Stack>
            )}

            {/* STEP 2: Confirm the results */}
            {confirmationStep && (
              <>
                <Typography
                  variant="h6"
                  align="center"
                  sx={{ fontWeight: 'bold', color: brandBlue, mb: 2, mt: 2 }}
                >
                  Confirm Your Students
                </Typography>

                <List>
                  {studentInfos.map((info, index) => (
                    <ListItem
                      key={index}
                      sx={{
                        flexDirection: 'column',
                        alignItems: 'flex-start',
                        border: '1px solid #ccc',
                        borderRadius: 2,
                        mb: 2,
                      }}
                    >
                      <ListItemText
                        primary={`Student ID: ${info.studentId}`}
                        secondary={`Student Name: ${info.studentName}`}
                        sx={{ mb: 1 }}
                      />
                      {!info.canLink && (
                        <Typography variant="body2" sx={{ color: 'red' }}>
                          Cannot link this student. Please check the ID.
                        </Typography>
                      )}

                      {info.canLink && info.confirmed === null && (
                        <Box sx={{ display: 'flex', gap: 2, mt: 1 }}>
                          <StyledButton
                            onClick={() => handleConfirmation(index, true)}
                            hovercolor="#28a745"
                            size="small"
                          >
                            Yes
                          </StyledButton>
                          <StyledButton
                            onClick={() => handleConfirmation(index, false)}
                            hovercolor="#d9480f"
                            size="small"
                          >
                            No
                          </StyledButton>
                        </Box>
                      )}

                      {info.confirmed === true && (
                        <Typography variant="body2" sx={{ color: 'green', mt: 1 }}>
                          Confirmed
                        </Typography>
                      )}
                      {info.confirmed === false && (
                        <Typography variant="body2" sx={{ color: 'orange', mt: 1 }}>
                          Not your student
                        </Typography>
                      )}
                    </ListItem>
                  ))}
                </List>

                <StyledButton
                  onClick={handleProceed}
                  disabled={!readyToProceed || loading}
                  sx={{ mt: 2 }}
                >
                  Proceed
                </StyledButton>
              </>
            )}
          </CardContent>
        </StyledCard>
      </IntakeContainer>
    </RootContainer>
  );
};

export default StudentIntake;
